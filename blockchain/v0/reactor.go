package v0

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"

	bc "github.com/fluentum-chain/fluentum/blockchain"
	"github.com/fluentum-chain/fluentum/libs/log"
	"github.com/fluentum-chain/fluentum/p2p"
	bcproto "github.com/fluentum-chain/fluentum/proto/tendermint/blockchain"
	sm "github.com/fluentum-chain/fluentum/state"
	"github.com/fluentum-chain/fluentum/store"
	"github.com/fluentum-chain/fluentum/types"
)

const (
	// BlockchainChannel is a channel for blocks and status updates (`BlockStore` height)
	BlockchainChannel = byte(0x40)

	trySyncIntervalMS = 10

	// stop syncing when last block's time is
	// within this much of the system time.
	// stopSyncingDurationMinutes = 10

	// ask for best height every 10s
	statusUpdateIntervalSeconds = 10
	// check if we should switch to consensus reactor
	switchToConsensusIntervalSeconds = 1
)

type consensusReactor interface {
	// for when we switch from blockchain reactor and fast sync to
	// the consensus machine
	SwitchToConsensus(state sm.State, skipWAL bool)
}

type peerError struct {
	err    error
	peerID p2p.ID
}

func (e peerError) Error() string {
	return fmt.Sprintf("error with peer %v: %s", e.peerID, e.err.Error())
}

// BlockchainReactor handles long-term catchup syncing.
type BlockchainReactor struct {
	p2p.BaseReactor

	// immutable
	initialState sm.State

	blockExec *sm.BlockExecutor
	store     *store.BlockStore
	pool      *BlockPool
	fastSync  bool

	requestsCh <-chan BlockRequest
	errorsCh   <-chan peerError
}

// NewBlockchainReactor returns new reactor instance.
func NewBlockchainReactor(state sm.State, blockExec *sm.BlockExecutor, store *store.BlockStore,
	fastSync bool) *BlockchainReactor {

	if state.LastBlockHeight != store.Height() {
		panic(fmt.Sprintf("state (%v) and store (%v) height mismatch", state.LastBlockHeight,
			store.Height()))
	}

	requestsCh := make(chan BlockRequest, maxTotalRequesters)

	const capacity = 1000                      // must be bigger than peers count
	errorsCh := make(chan peerError, capacity) // so we don't block in #Receive#pool.AddBlock

	startHeight := store.Height() + 1
	if startHeight == 1 {
		startHeight = state.InitialHeight
	}
	pool := NewBlockPool(startHeight, requestsCh, errorsCh)

	bcR := &BlockchainReactor{
		initialState: state,
		blockExec:    blockExec,
		store:        store,
		pool:         pool,
		fastSync:     fastSync,
		requestsCh:   requestsCh,
		errorsCh:     errorsCh,
	}
	bcR.BaseReactor = *p2p.NewBaseReactor("BlockchainReactor", bcR)
	return bcR
}

// SetLogger implements service.Service by setting the logger on reactor and pool.
func (bcR *BlockchainReactor) SetLogger(l log.Logger) {
	bcR.BaseService.Logger = l
	bcR.pool.Logger = l
}

// OnStart implements service.Service.
func (bcR *BlockchainReactor) OnStart() error {
	if bcR.fastSync {
		err := bcR.pool.Start()
		if err != nil {
			return err
		}
		go bcR.poolRoutine(false)
	}
	return nil
}

// SwitchToFastSync is called by the state sync reactor when switching to fast sync.
func (bcR *BlockchainReactor) SwitchToFastSync(state sm.State) error {
	bcR.fastSync = true
	bcR.initialState = state

	bcR.pool.height = state.LastBlockHeight + 1
	err := bcR.pool.Start()
	if err != nil {
		return err
	}
	go bcR.poolRoutine(true)
	return nil
}

// OnStop implements service.Service.
func (bcR *BlockchainReactor) OnStop() {
	if bcR.fastSync {
		if err := bcR.pool.Stop(); err != nil {
			bcR.Logger.Error("Error stopping pool", "err", err)
		}
	}
}

// GetChannels implements Reactor
func (bcR *BlockchainReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  BlockchainChannel,
			Priority:            5,
			SendQueueCapacity:   1000,
			RecvBufferCapacity:  50 * 4096,
			RecvMessageCapacity: bc.MaxMsgSize,
			MessageType:         &bcproto.Message{},
		},
	}
}

// AddPeer implements Reactor by sending our state to peer.
func (bcR *BlockchainReactor) AddPeer(peer p2p.Peer) {
	p2p.SendEnvelopeShim(peer, p2p.Envelope{ //nolint: staticcheck
		ChannelID: BlockchainChannel,
		Message: &bcproto.StatusResponse{
			Base:   bcR.store.Base(),
			Height: bcR.store.Height(),
		},
	}, bcR.Logger)
	// it's OK if send fails. will try later in poolRoutine

	// peer is added to the pool once we receive the first
	// bcStatusResponseMessage from the peer and call pool.SetPeerRange
}

// RemovePeer implements Reactor by removing peer from the pool.
func (bcR *BlockchainReactor) RemovePeer(peer p2p.Peer, reason interface{}) {
	bcR.pool.RemovePeer(peer.ID())
}

// respondToPeer loads a block and sends it to the requesting peer,
// if we have it. Otherwise, we'll respond saying we don't have it.
func (bcR *BlockchainReactor) respondToPeer(msg *bcproto.BlockRequest,
	src p2p.Peer) (queued bool) {

	block := bcR.store.LoadBlock(msg.Height)
	if block != nil {
		bl, err := block.ToProto()
		if err != nil {
			bcR.Logger.Error("could not convert msg to protobuf", "err", err)
			return false
		}
		return p2p.TrySendEnvelopeShim(src, p2p.Envelope{ //nolint: staticcheck
			ChannelID: BlockchainChannel,
			Message:   &bcproto.BlockResponse{Block: bl},
		}, bcR.Logger)
	}

	return p2p.TrySendEnvelopeShim(src, p2p.Envelope{ //nolint: staticcheck
		ChannelID: BlockchainChannel,
		Message:   &bcproto.NoBlockResponse{Height: msg.Height},
	}, bcR.Logger)
}

func (bcR *BlockchainReactor) ReceiveEnvelope(e p2p.Envelope) {
	if err := bc.ValidateMsg(e.Message); err != nil {
		bcR.Logger.Error("Peer sent us invalid msg", "peer", e.Src, "msg", e.Message, "err", err)
		bcR.Switch.StopPeerForError(e.Src, err)
		return
	}

	bcR.Logger.Debug("Receive", "e.Src", e.Src, "chID", e.ChannelID, "msg", e.Message)

	switch msg := e.Message.(type) {
	case *bcproto.BlockRequest:
		bcR.respondToPeer(msg, e.Src)
	case *bcproto.BlockResponse:
		bi, err := types.BlockFromProto(msg.Block)
		if err != nil {
			bcR.Logger.Error("Block content is invalid", "err", err)
			return
		}
		bcR.pool.AddBlock(e.Src.ID(), bi, proto.Size(msg.Block))
	case *bcproto.StatusRequest:
		// Send peer our state.
		p2p.TrySendEnvelopeShim(e.Src, p2p.Envelope{ //nolint: staticcheck
			ChannelID: BlockchainChannel,
			Message: &bcproto.StatusResponse{
				Height: bcR.store.Height(),
				Base:   bcR.store.Base(),
			},
		}, bcR.Logger)
	case *bcproto.StatusResponse:
		// Got a peer status. Unverified.
		bcR.pool.SetPeerRange(e.Src.ID(), msg.Base, msg.Height)
	case *bcproto.NoBlockResponse:
		bcR.Logger.Debug("Peer does not have requested block", "peer", e.Src, "height", msg.Height)
	default:
		bcR.Logger.Error(fmt.Sprintf("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

func (bcR *BlockchainReactor) Receive(chID byte, peer p2p.Peer, msgBytes []byte) {
	msg := &bcproto.Message{}
	err := proto.Unmarshal(msgBytes, msg)
	if err != nil {
		panic(err)
	}
	uw, err := msg.Unwrap()
	if err != nil {
		panic(err)
	}
	bcR.ReceiveEnvelope(p2p.Envelope{
		ChannelID: chID,
		Src:       peer,
		Message:   uw,
	})
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
func (bcR *BlockchainReactor) poolRoutine(stateSynced bool) {

	trySyncTicker := time.NewTicker(trySyncIntervalMS * time.Millisecond)
	defer trySyncTicker.Stop()

	statusUpdateTicker := time.NewTicker(statusUpdateIntervalSeconds * time.Second)
	defer statusUpdateTicker.Stop()

	switchToConsensusTicker := time.NewTicker(switchToConsensusIntervalSeconds * time.Second)
	defer switchToConsensusTicker.Stop()

	blocksSynced := uint64(0)

	chainID := bcR.initialState.ChainID
	state := bcR.initialState

	lastHundred := time.Now()
	lastRate := 0.0

	didProcessCh := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-bcR.Quit():
				return
			case <-bcR.pool.Quit():
				return
			case request := <-bcR.requestsCh:
				peer := bcR.Switch.Peers().Get(request.PeerID)
				if peer == nil {
					continue
				}
				queued := p2p.TrySendEnvelopeShim(peer, p2p.Envelope{ //nolint: staticcheck
					ChannelID: BlockchainChannel,
					Message:   &bcproto.BlockRequest{Height: request.Height},
				}, bcR.Logger)
				if !queued {
					bcR.Logger.Debug("Send queue is full, drop block request", "peer", peer.ID(), "height", request.Height)
				}
			case err := <-bcR.errorsCh:
				peer := bcR.Switch.Peers().Get(err.peerID)
				if peer != nil {
					bcR.Switch.StopPeerForError(peer, err)
				}

			case <-statusUpdateTicker.C:
				// ask for status updates
				go bcR.BroadcastStatusRequest() //nolint: errcheck

			}
		}
	}()

FOR_LOOP:
	for {
		select {
		case <-switchToConsensusTicker.C:
			height, numPending, lenRequesters := bcR.pool.GetStatus()
			outbound, inbound, _ := bcR.Switch.NumPeers()
			bcR.Logger.Debug("Consensus ticker", "numPending", numPending, "total", lenRequesters,
				"outbound", outbound, "inbound", inbound)
			if bcR.pool.IsCaughtUp() {
				bcR.Logger.Info("Time to switch to consensus reactor!", "height", height)
				if err := bcR.pool.Stop(); err != nil {
					bcR.Logger.Error("Error stopping pool", "err", err)
				}
				conR, ok := bcR.Switch.Reactor("CONSENSUS").(consensusReactor)
				if ok {
					conR.SwitchToConsensus(state, blocksSynced > 0 || stateSynced)
				}
				// else {
				// should only happen during testing
				// }

				break FOR_LOOP
			}

		case <-trySyncTicker.C: // chan time
			select {
			case didProcessCh <- struct{}{}:
			default:
			}

		case <-didProcessCh:
			// NOTE: It is a subtle mistake to process more than a single block
			// at a time (e.g. 10) here, because we only TrySend 1 request per
			// loop.  The ratio mismatch can result in starving of blocks, a
			// sudden burst of requests and responses, and repeat.
			// Consequently, it is better to split these routines rather than
			// coupling them as it's written here.  TODO uncouple from request
			// routine.

			// See if there are any blocks to sync.
			first, second := bcR.pool.PeekTwoBlocks()
			// bcR.Logger.Info("TrySync peeked", "first", first, "second", second)
			if first == nil || second == nil {
				// We need both to sync the first block.
				continue FOR_LOOP
			} else {
				// Try again quickly next loop.
				didProcessCh <- struct{}{}
			}

			firstParts := first.MakePartSet(types.BlockPartSizeBytes)
			firstPartSetHeader := firstParts.Header()
			firstID := types.BlockID{Hash: first.Hash(), PartSetHeader: firstPartSetHeader}
			// Finally, verify the first block using the second's commit
			// NOTE: we can probably make this more efficient, but note that calling
			// first.Hash() doesn't verify the tx contents, so MakePartSet() is
			// currently necessary.
			err := state.Validators.VerifyCommitLight(
				chainID, firstID, first.Height, second.LastCommit)

			if err == nil {
				// validate the block before we persist it
				err = bcR.blockExec.ValidateBlock(state, first)
			}

			if err != nil {
				bcR.Logger.Error("Error in validation", "err", err)
				peerID := bcR.pool.RedoRequest(first.Height)
				peer := bcR.Switch.Peers().Get(peerID)
				if peer != nil {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					bcR.Switch.StopPeerForError(peer, fmt.Errorf("blockchainReactor validation error: %v", err))
				}
				peerID2 := bcR.pool.RedoRequest(second.Height)
				peer2 := bcR.Switch.Peers().Get(peerID2)
				if peer2 != nil && peer2 != peer {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					bcR.Switch.StopPeerForError(peer2, fmt.Errorf("blockchainReactor validation error: %v", err))
				}
				continue FOR_LOOP
			}

			bcR.pool.PopRequest()

			// TODO: batch saves so we dont persist to disk every block
			bcR.store.SaveBlock(first, firstParts, second.LastCommit)

			// TODO: same thing for app - but we would need a way to
			// get the hash without persisting the state
			state, _, err = bcR.blockExec.ApplyBlock(state, firstID, first)
			if err != nil {
				// TODO This is bad, are we zombie?
				panic(fmt.Sprintf("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
			}
			blocksSynced++

			if blocksSynced%100 == 0 {
				lastRate = 0.9*lastRate + 0.1*(100/time.Since(lastHundred).Seconds())
				bcR.Logger.Info("Fast Sync Rate", "height", bcR.pool.height,
					"max_peer_height", bcR.pool.MaxPeerHeight(), "blocks/s", lastRate)
				lastHundred = time.Now()
			}

			continue FOR_LOOP

		case <-bcR.Quit():
			break FOR_LOOP
		}
	}
}

// BroadcastStatusRequest broadcasts `BlockStore` base and height.
func (bcR *BlockchainReactor) BroadcastStatusRequest() error {
	bcR.Switch.BroadcastEnvelope(p2p.Envelope{
		ChannelID: BlockchainChannel,
		Message:   &bcproto.StatusRequest{},
	})
	return nil
}
