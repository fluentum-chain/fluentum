package state

import (
	"context"
	"fmt"
	"time"

	cometbftabci "github.com/cometbft/cometbft/abci/types"
	cometproto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	abci "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/crypto/encoding"
	"github.com/fluentum-chain/fluentum/libs/fail"
	"github.com/fluentum-chain/fluentum/libs/log"
	mempl "github.com/fluentum-chain/fluentum/mempool"
	protocrypto "github.com/fluentum-chain/fluentum/proto/tendermint/crypto"
	tmproto "github.com/fluentum-chain/fluentum/proto/tendermint/types"
	"github.com/fluentum-chain/fluentum/proxy"
	"github.com/fluentum-chain/fluentum/types"
)

//-----------------------------------------------------------------------------
// BlockExecutor handles block execution and state updates.
// It exposes ApplyBlock(), which validates & executes the block, updates state w/ ABCI responses,
// then commits and updates the mempool atomically, then saves state.

// BlockExecutor provides the context and accessories for properly executing a block.
type BlockExecutor struct {
	// save state, validators, consensus params, abci responses here
	store Store

	// execute the app against this
	proxyApp proxy.AppConnConsensus

	// events
	eventBus types.BlockEventPublisher

	// manage the mempool lock during commit
	// and update both with block results after commit.
	mempool mempl.Mempool
	evpool  EvidencePool

	logger log.Logger

	metrics *Metrics
}

type BlockExecutorOption func(executor *BlockExecutor)

func BlockExecutorWithMetrics(metrics *Metrics) BlockExecutorOption {
	return func(blockExec *BlockExecutor) {
		blockExec.metrics = metrics
	}
}

// NewBlockExecutor returns a new BlockExecutor with a NopEventBus.
// Call SetEventBus to provide one.
func NewBlockExecutor(
	stateStore Store,
	logger log.Logger,
	proxyApp proxy.AppConnConsensus,
	mempool mempl.Mempool,
	evpool EvidencePool,
	options ...BlockExecutorOption,
) *BlockExecutor {
	res := &BlockExecutor{
		store:    stateStore,
		proxyApp: proxyApp,
		eventBus: types.NopEventBus{},
		mempool:  mempool,
		evpool:   evpool,
		logger:   logger,
		metrics:  NopMetrics(),
	}

	for _, option := range options {
		option(res)
	}

	return res
}

func (blockExec *BlockExecutor) Store() Store {
	return blockExec.store
}

// SetEventBus - sets the event bus for publishing block related events.
// If not called, it defaults to types.NopEventBus.
func (blockExec *BlockExecutor) SetEventBus(eventBus types.BlockEventPublisher) {
	blockExec.eventBus = eventBus
}

// CreateProposalBlock calls state.MakeBlock with evidence from the evpool
// and txs from the mempool. The max bytes must be big enough to fit the commit.
// Up to 1/10th of the block space is allcoated for maximum sized evidence.
// The rest is given to txs, up to the max gas.
func (blockExec *BlockExecutor) CreateProposalBlock(
	height int64,
	state State, commit *types.Commit,
	proposerAddr []byte,
) (*types.Block, *types.PartSet) {

	maxBytes := state.ConsensusParams.Block.MaxBytes
	maxGas := state.ConsensusParams.Block.MaxGas

	evidence, evSize := blockExec.evpool.PendingEvidence(state.ConsensusParams.Evidence.MaxBytes)

	// Fetch a limited amount of valid txs
	maxDataBytes := types.MaxDataBytes(maxBytes, evSize, state.Validators.Size())

	txs := blockExec.mempool.ReapMaxBytesMaxGas(maxDataBytes, maxGas)

	return state.MakeBlock(height, txs, commit, evidence, proposerAddr)
}

// ValidateBlock validates the given block against the given state.
// If the block is invalid, it returns an error.
// Validation does not mutate state, but does require historical information from the stateDB,
// ie. to verify evidence from a validator at an old height.
func (blockExec *BlockExecutor) ValidateBlock(state State, block *types.Block) error {
	err := validateBlock(state, block)
	if err != nil {
		return err
	}
	return blockExec.evpool.CheckEvidence(block.Evidence.Evidence)
}

// ApplyBlock validates the block against the state, executes it against the app,
// fires the relevant events, commits the app, and saves the new state and responses.
// It returns the new state and the block height to retain (pruning older blocks).
// It's the only function that needs to be called
// from outside this package to process and commit an entire block.
// It takes a blockID to avoid recomputing the parts hash.
func (blockExec *BlockExecutor) ApplyBlock(
	state State, blockID types.BlockID, block *types.Block,
) (State, int64, error) {

	if err := validateBlock(state, block); err != nil {
		return state, 0, ErrInvalidBlock(err)
	}

	startTime := time.Now().UnixNano()
	ctx := context.Background()
	abciResponses, err := execBlockOnProxyApp(
		ctx, blockExec.logger, blockExec.proxyApp, block, blockExec.store, state.InitialHeight,
	)
	endTime := time.Now().UnixNano()
	blockExec.metrics.BlockProcessingTime.Observe(float64(endTime-startTime) / 1000000)
	if err != nil {
		return state, 0, ErrProxyAppConn(err)
	}

	fail.Fail() // XXX

	// Save the results before we commit.
	if err := blockExec.store.SaveABCIResponses(block.Height, abciResponses); err != nil {
		return state, 0, err
	}

	fail.Fail() // XXX

	// validate the validator updates and convert to tendermint types
	abciValUpdates := abciResponses.FinalizeBlock.ValidatorUpdates
	err = validateValidatorUpdates(abciValUpdates, state.ConsensusParams.Validator)
	if err != nil {
		return state, 0, fmt.Errorf("error in validator updates: %v", err)
	}

	// Convert CometBFT validator updates to Tendermint types
	var validatorUpdates []*types.Validator
	for _, update := range abciValUpdates {
		localPubKey := toLocalPubKey(update.PubKey)
		pubKey, err := encoding.PubKeyFromProto(*localPubKey)
		if err != nil {
			return state, 0, fmt.Errorf("invalid validator pubkey: %v", err)
		}
		validatorUpdates = append(validatorUpdates, &types.Validator{
			Address:     pubKey.Address(),
			PubKey:      pubKey,
			VotingPower: update.Power,
		})
	}

	if len(validatorUpdates) > 0 {
		blockExec.logger.Debug("updates to validators", "updates", types.ValidatorListString(validatorUpdates))
	}

	// Update the state with the block and responses.
	state, err = updateState(state, blockID, &block.Header, abciResponses, validatorUpdates)
	if err != nil {
		return state, 0, fmt.Errorf("commit failed for application: %v", err)
	}

	// Lock mempool, commit app state, update mempoool.
	appHash, retainHeight, err := blockExec.Commit(state, block, abciResponses.FinalizeBlock.TxResults)
	if err != nil {
		return state, 0, fmt.Errorf("commit failed for application: %v", err)
	}

	// Update evpool with the latest state.
	blockExec.evpool.Update(state, block.Evidence.Evidence)

	fail.Fail() // XXX

	// Update the state with the block and responses.
	state.AppHash = appHash
	if err := blockExec.store.Save(state); err != nil {
		return state, 0, err
	}

	fail.Fail() // XXX

	// Events are fired after everything else.
	// NOTE: if we crash between Commit and Save, events wont be fired during replay
	fireEvents(blockExec.logger, blockExec.eventBus, block, abciResponses, validatorUpdates)

	return state, retainHeight, nil
}

// Commit locks the mempool, runs the ABCI Commit message, and updates the
// mempool.
// It returns the result of calling abci.Commit (the AppHash) and the height to retain (if any).
// The Mempool must be locked during commit and update because state is
// typically reset on Commit and old txs must be replayed against committed
// state before new txs are run in the mempool, lest they be invalid.
func (blockExec *BlockExecutor) Commit(
	state State,
	block *types.Block,
	txResults []*cometbftabci.ExecTxResult,
) ([]byte, int64, error) {
	blockExec.mempool.Lock()
	defer blockExec.mempool.Unlock()

	// while mempool is Locked, flush to ensure all async requests have completed
	// in the ABCI app before Commit.
	err := blockExec.mempool.FlushAppConn()
	if err != nil {
		blockExec.logger.Error("client error during mempool.FlushAppConn", "err", err)
		return nil, 0, err
	}

	// Commit block, get hash back
	res, err := blockExec.proxyApp.CommitSync(context.Background(), &abci.CommitRequest{})
	if err != nil {
		blockExec.logger.Error("client error during proxyAppConn.CommitSync", "err", err)
		return nil, 0, err
	}

	blockExec.logger.Info(
		"committed state",
		"height", block.Height,
		"num_txs", len(block.Txs),
	)

	// Update mempool.
	err = blockExec.mempool.Update(
		block.Height,
		block.Txs,
		txResults,
		TxPreCheck(state),
		TxPostCheck(state),
	)

	return nil, res.RetainHeight, err
}

//---------------------------------------------------------
// Helper functions for executing blocks and updating state

// Executes block's transactions on proxyAppConn using ABCI++ FinalizeBlock.
// Returns a structure with the FinalizeBlock response and validator updates.
func execBlockOnProxyApp(
	ctx context.Context,
	logger log.Logger,
	proxyAppConn proxy.AppConnConsensus,
	block *types.Block,
	store Store,
	initialHeight int64,
) (*ABCIResponses, error) {
	// Convert evidence to CometBFT ABCI format
	var byzantineValidators []*cometbftabci.Misbehavior
	for _, ev := range block.Evidence.Evidence {
		for _, evi := range ev.ABCI() {
			if pb, ok := any(evi).(*cometbftabci.Misbehavior); ok {
				byzantineValidators = append(byzantineValidators, pb)
			}
		}
	}

	txs := make([][]byte, len(block.Txs))
	for i, tx := range block.Txs {
		txs[i] = tx
	}

	// Convert []*Misbehavior to []Misbehavior
	var byzVals []cometbftabci.Misbehavior
	for _, pb := range byzantineValidators {
		if pb != nil {
			byzVals = append(byzVals, *pb)
		}
	}

	finalizeBlockReq := &cometbftabci.RequestFinalizeBlock{
		Hash:               block.Hash(),
		Height:             block.Height,
		Time:               block.Time,
		Txs:                txs,
		ProposerAddress:    block.ProposerAddress,
		DecidedLastCommit:  *getBeginBlockValidatorInfo(block, store, initialHeight),
		Misbehavior:        byzVals,
		NextValidatorsHash: block.NextValidatorsHash,
	}
	finalizeBlockResp, err := proxyAppConn.FinalizeBlock(ctx, finalizeBlockReq)
	if err != nil {
		logger.Error("error in proxyAppConn.FinalizeBlock", "err", err)
		return nil, err
	}

	abciResponses := &ABCIResponses{
		FinalizeBlock: finalizeBlockResp,
	}

	return abciResponses, nil
}

func getBeginBlockValidatorInfo(block *types.Block, store Store,
	initialHeight int64) *cometbftabci.CommitInfo {
	voteInfos := make([]cometbftabci.VoteInfo, block.LastCommit.Size())
	if block.Height > initialHeight {
		lastValSet, err := store.LoadValidators(block.Height - 1)
		if err != nil {
			panic(err)
		}
		commitSize := block.LastCommit.Size()
		valSetLen := len(lastValSet.Validators)
		if commitSize != valSetLen {
			panic(fmt.Sprintf(
				"commit size (%d) doesn't match valset length (%d) at height %d\n\n%v\n\n%v",
				commitSize, valSetLen, block.Height, block.LastCommit.Signatures, lastValSet.Validators,
			))
		}
		for i, val := range lastValSet.Validators {
			voteInfos[i] = cometbftabci.VoteInfo{
				Validator: cometbftabci.Validator{
					Address: val.PubKey.Address(),
					Power:   val.VotingPower,
				},
				BlockIdFlag: 2, // COMMIT
			}
		}
	}
	return &cometbftabci.CommitInfo{
		Round: block.LastCommit.Round,
		Votes: voteInfos,
	}
}

func validateValidatorUpdates(abciUpdates []cometbftabci.ValidatorUpdate, params tmproto.ValidatorParams) error {
	for _, valUpdate := range abciUpdates {
		localPubKey := toLocalPubKey(valUpdate.PubKey)
		pk, err := encoding.PubKeyFromProto(*localPubKey)
		if err != nil {
			return err
		}
		if !types.IsValidPubkeyType(params, pk.Type()) {
			return fmt.Errorf("validator %v is using pubkey %s, which is unsupported for consensus", valUpdate, pk.Type())
		}
	}
	return nil
}

// updateState returns a new State updated according to the header and responses.
func updateState(
	state State,
	blockID types.BlockID,
	header *types.Header,
	abciResponses *ABCIResponses,
	validatorUpdates []*types.Validator,
) (State, error) {

	// Copy the valset so we can apply changes from EndBlock
	// and update s.LastValidators and s.Validators.
	nValSet := state.NextValidators.Copy()

	// Update the validator set with the latest abciResponses.
	lastHeightValsChanged := state.LastHeightValidatorsChanged
	if len(validatorUpdates) > 0 {
		err := nValSet.UpdateWithChangeSet(validatorUpdates)
		if err != nil {
			return state, fmt.Errorf("error changing validator set: %v", err)
		}
		// Change results from this height but only applies to the next next height.
		lastHeightValsChanged = header.Height + 1 + 1
	}

	// Update validator proposer priority and set state variables.
	nValSet.IncrementProposerPriority(1)

	// Update the params with the latest abciResponses.
	nextParams := state.ConsensusParams
	lastHeightParamsChanged := state.LastHeightConsensusParamsChanged
	if abciResponses.FinalizeBlock.ConsensusParamUpdates != nil {
		// TODO: convert ConsensusParamUpdates to the correct type if needed
		nextParams = types.UpdateConsensusParams(state.ConsensusParams, nil)
		err := types.ValidateConsensusParams(nextParams)
		if err != nil {
			return state, fmt.Errorf("error updating consensus params: %v", err)
		}
		state.Version.Consensus.App = nextParams.Version.AppVersion
		lastHeightParamsChanged = header.Height + 1
	}

	nextVersion := state.Version

	// NOTE: the AppHash has not been populated.
	// It will be filled on state.Save.
	return State{
		Version:                          nextVersion,
		ChainID:                          state.ChainID,
		InitialHeight:                    state.InitialHeight,
		LastBlockHeight:                  header.Height,
		LastBlockID:                      blockID,
		LastBlockTime:                    header.Time,
		NextValidators:                   nValSet,
		Validators:                       state.NextValidators.Copy(),
		LastValidators:                   state.Validators.Copy(),
		LastHeightValidatorsChanged:      lastHeightValsChanged,
		ConsensusParams:                  nextParams,
		LastHeightConsensusParamsChanged: lastHeightParamsChanged,
		LastResultsHash:                  ABCIResponsesResultsHash(abciResponses),
		AppHash:                          nil,
	}, nil
}

// Fire NewBlock, NewBlockHeader.
// Fire TxEvent for every tx.
// NOTE: if Tendermint crashes before commit, some or all of these events may be published again.
func fireEvents(
	logger log.Logger,
	eventBus types.BlockEventPublisher,
	block *types.Block,
	abciResponses *ABCIResponses,
	validatorUpdates []*types.Validator,
) {
	if err := eventBus.PublishEventNewBlock(types.EventDataNewBlock{
		Block: block,
	}); err != nil {
		logger.Error("failed publishing new block", "err", err)
	}

	if err := eventBus.PublishEventNewBlockHeader(types.EventDataNewBlockHeader{
		Header:              block.Header,
		NumTxs:              int64(len(block.Txs)),
		ResultFinalizeBlock: abciResponses.FinalizeBlock,
	}); err != nil {
		logger.Error("failed publishing new block header", "err", err)
	}

	if len(block.Evidence.Evidence) != 0 {
		for _, ev := range block.Evidence.Evidence {
			if err := eventBus.PublishEventNewEvidence(types.EventDataNewEvidence{
				Evidence: ev,
				Height:   block.Height,
			}); err != nil {
				logger.Error("failed publishing new evidence", "err", err)
			}
		}
	}

	for i, tx := range block.Data.Txs {
		if err := eventBus.PublishEventTx(types.EventDataTx{TxResult: abci.TxResult{
			Height: block.Height,
			Index:  uint32(i),
			Tx:     tx,
			Result: *(abciResponses.FinalizeBlock.TxResults[i]),
		}}); err != nil {
			logger.Error("failed publishing event TX", "err", err)
		}
	}

	if len(validatorUpdates) > 0 {
		if err := eventBus.PublishEventValidatorSetUpdates(
			types.EventDataValidatorSetUpdates{ValidatorUpdates: validatorUpdates}); err != nil {
			logger.Error("failed publishing event", "err", err)
		}
	}
}

//----------------------------------------------------------------------------------------------------
// Execute block without state. TODO: eliminate

// ExecCommitBlock executes and commits a block on the proxyApp without validating or mutating the state.
// It returns the application root hash (result of abci.Commit).
func ExecCommitBlock(
	appConnConsensus proxy.AppConnConsensus,
	block *types.Block,
	logger log.Logger,
	store Store,
	initialHeight int64,
) ([]byte, error) {
	_, err := execBlockOnProxyApp(context.Background(), logger, appConnConsensus, block, store, initialHeight)
	if err != nil {
		logger.Error("failed executing block on proxy app", "height", block.Height, "err", err)
		return nil, err
	}
	// Commit block, get hash back
	res, err := appConnConsensus.CommitSync(context.Background(), nil)
	if err != nil {
		logger.Error("client error during proxyAppConn.CommitSync", "err", res)
		return nil, err
	}
	return nil, nil
}

// Helper to convert CometBFT proto PublicKey to local proto PublicKey
func toLocalPubKey(pk cometproto.PublicKey) *protocrypto.PublicKey {
	// For now, return a simple implementation
	// In a real implementation, you'd need to handle different key types properly
	return &protocrypto.PublicKey{
		Sum: &protocrypto.PublicKey_Ed25519{
			Ed25519: []byte{}, // TODO: Extract actual key bytes from pk
		},
	}
}
