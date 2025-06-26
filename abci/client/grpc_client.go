package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ Client = (*grpcClient)(nil)

type grpcClient struct {
	client cmtproto.ABCIApplicationClient
	conn   *grpc.ClientConn
	mtx    sync.Mutex
	logger Logger
}

// NewGRPCClient creates a new gRPC client
func NewGRPCClient(addr string, mustConnect bool) Client {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		if mustConnect {
			panic(fmt.Sprintf("failed to connect to gRPC server: %v", err))
		}
		return &grpcClient{conn: nil}
	}

	client := cmtproto.NewABCIApplicationClient(conn)
	return &grpcClient{
		client: client,
		conn:   conn,
	}
}

func (c *grpcClient) SetResponseCallback(cb Callback) {
	// gRPC client doesn't use callbacks in the same way
}

func (c *grpcClient) SetLogger(l Logger) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.logger = l
}

func (c *grpcClient) Error() error {
	if c.conn == nil {
		return ErrConnectionNotInitialized
	}
	return nil
}

// Mempool methods
func (c *grpcClient) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateTxData(req.Tx); err != nil {
		return nil, fmt.Errorf("CheckTx validation failed: %w", err)
	}

	pbReq := &cmtproto.RequestCheckTx{
		Tx:   req.Tx,
		Type: cmtproto.CheckTxType(req.Type),
	}
	pbRes, err := c.client.CheckTx(ctx, pbReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("CheckTx failed", "err", err)
		}
		return nil, fmt.Errorf("CheckTx failed: %w", err)
	}
	return &cmtabci.ResponseCheckTx{
		Code:      pbRes.Code,
		Data:      pbRes.Data,
		Log:       pbRes.Log,
		Info:      pbRes.Info,
		GasWanted: pbRes.GasWanted,
		GasUsed:   pbRes.GasUsed,
		Events:    fromProtoEvents(pbRes.Events),
		Codespace: pbRes.Codespace,
	}, nil
}

// Helper to convert proto events to abci events
func fromProtoEvents(events []*cmtproto.Event) []cmtabci.Event {
	if events == nil {
		return nil
	}
	result := make([]cmtabci.Event, len(events))
	for i, ev := range events {
		attrs := make([]cmtabci.EventAttribute, len(ev.Attributes))
		for j, attr := range ev.Attributes {
			attrs[j] = cmtabci.EventAttribute{
				Key:   attr.Key,
				Value: attr.Value,
				Index: attr.Index,
			}
		}
		result[i] = cmtabci.Event{
			Type:       ev.Type,
			Attributes: attrs,
		}
	}
	return result
}

func (c *grpcClient) CheckTxAsync(ctx context.Context, req *cmtabci.RequestCheckTx) *ReqRes {
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_CheckTx{CheckTx: req}})
	go func() {
		res, err := c.CheckTx(ctx, req)
		if err != nil {
			reqRes.ErrorCh <- err
		} else {
			reqRes.ResponseCh <- res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (c *grpcClient) Flush(ctx context.Context) error {
	// gRPC doesn't need explicit flushing
	return nil
}

// Consensus methods
func (c *grpcClient) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("FinalizeBlock validation failed: %w", err)
	}

	protoReq := &cmtproto.RequestFinalizeBlock{
		Height: req.Height,
		Txs:    req.Txs,
		Hash:   req.Hash,
		Header: req.Header,
	}

	res, err := c.client.FinalizeBlock(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("FinalizeBlock failed", "err", err)
		}
		return nil, fmt.Errorf("FinalizeBlock failed: %w", err)
	}

	return &cmtabci.ResponseFinalizeBlock{
		TxResults:             res.TxResults,
		ValidatorUpdates:      res.ValidatorUpdates,
		ConsensusParamUpdates: res.ConsensusParamUpdates,
		AppHash:               res.AppHash,
		Events:                res.Events,
	}, nil
}

func (c *grpcClient) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if req.MaxTxBytes <= 0 {
		return nil, fmt.Errorf("invalid max tx bytes: %d", req.MaxTxBytes)
	}

	protoReq := &cmtproto.RequestPrepareProposal{
		MaxTxBytes: req.MaxTxBytes,
		Txs:        req.Txs,
	}

	res, err := c.client.PrepareProposal(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("PrepareProposal failed", "err", err)
		}
		return nil, fmt.Errorf("PrepareProposal failed: %w", err)
	}

	return &cmtabci.ResponsePrepareProposal{
		Txs: res.Txs,
	}, nil
}

func (c *grpcClient) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ProcessProposal validation failed: %w", err)
	}

	protoReq := &cmtproto.RequestProcessProposal{
		Height: req.Height,
		Txs:    req.Txs,
		Hash:   req.Hash,
		Header: req.Header,
	}

	res, err := c.client.ProcessProposal(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ProcessProposal failed", "err", err)
		}
		return nil, fmt.Errorf("ProcessProposal failed: %w", err)
	}

	return &cmtabci.ResponseProcessProposal{
		Status: cmtabci.ResponseProcessProposal_Status(res.Status),
	}, nil
}

func (c *grpcClient) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ExtendVote validation failed: %w", err)
	}

	protoReq := &cmtproto.RequestExtendVote{
		Height: req.Height,
		Round:  req.Round,
		Hash:   req.Hash,
	}

	res, err := c.client.ExtendVote(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ExtendVote failed", "err", err)
		}
		return nil, fmt.Errorf("ExtendVote failed: %w", err)
	}

	return &cmtabci.ResponseExtendVote{
		VoteExtension: res.VoteExtension,
	}, nil
}

func (c *grpcClient) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("VerifyVoteExtension validation failed: %w", err)
	}

	protoReq := &cmtproto.RequestVerifyVoteExtension{
		Height:        req.Height,
		Round:         req.Round,
		Hash:          req.Hash,
		VoteExtension: req.VoteExtension,
	}

	res, err := c.client.VerifyVoteExtension(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("VerifyVoteExtension failed", "err", err)
		}
		return nil, fmt.Errorf("VerifyVoteExtension failed: %w", err)
	}

	return &cmtabci.ResponseVerifyVoteExtension{
		Status: cmtabci.ResponseVerifyVoteExtension_Status(res.Status),
	}, nil
}

func (c *grpcClient) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestCommit{}
	res, err := c.client.Commit(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Commit failed", "err", err)
		}
		return nil, fmt.Errorf("Commit failed: %w", err)
	}

	return &cmtabci.ResponseCommit{
		Data:         res.Data,
		RetainHeight: res.RetainHeight,
	}, nil
}

func (c *grpcClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestInitChain{
		Time:    req.Time,
		ChainId: req.ChainId,
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{
				MaxBytes: req.ConsensusParams.Block.MaxBytes,
				MaxGas:   req.ConsensusParams.Block.MaxGas,
			},
			Evidence: &cmtproto.EvidenceParams{
				MaxAgeNumBlocks: req.ConsensusParams.Evidence.MaxAgeNumBlocks,
				MaxAgeDuration:  req.ConsensusParams.Evidence.MaxAgeDuration,
				MaxBytes:        req.ConsensusParams.Evidence.MaxBytes,
			},
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: req.ConsensusParams.Validator.PubKeyTypes,
			},
			Version: &cmtproto.VersionParams{
				AppVersion: req.ConsensusParams.Version.AppVersion,
			},
		},
		Validators: req.Validators,
		AppStateBytes: req.AppStateBytes,
		InitialHeight: req.InitialHeight,
	}

	res, err := c.client.InitChain(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("InitChain failed", "err", err)
		}
		return nil, fmt.Errorf("InitChain failed: %w", err)
	}

	return &cmtabci.ResponseInitChain{
		ConsensusParams: res.ConsensusParams,
		Validators:      res.Validators,
		AppHash:         res.AppHash,
	}, nil
}

// Query methods
func (c *grpcClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestInfo{
		Version: req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion: req.P2PVersion,
		AbciVersion: req.AbciVersion,
	}

	res, err := c.client.Info(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Info failed", "err", err)
		}
		return nil, fmt.Errorf("Info failed: %w", err)
	}

	return &cmtabci.ResponseInfo{
		Data:             res.Data,
		Version:          res.Version,
		AppVersion:       res.AppVersion,
		BlockVersion:     res.BlockVersion,
		P2PVersion:       res.P2PVersion,
		AbciVersion:      res.AbciVersion,
	}, nil
}

func (c *grpcClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestQuery{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}

	res, err := c.client.Query(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Query failed", "err", err)
		}
		return nil, fmt.Errorf("Query failed: %w", err)
	}

	return &cmtabci.ResponseQuery{
		Code:   res.Code,
		Log:    res.Log,
		Info:   res.Info,
		Index:  res.Index,
		Key:    res.Key,
		Value:  res.Value,
		ProofOps: res.ProofOps,
		Height: res.Height,
		Codespace: res.Codespace,
	}, nil
}

// Snapshot methods
func (c *grpcClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestListSnapshots{}
	res, err := c.client.ListSnapshots(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ListSnapshots failed", "err", err)
		}
		return nil, fmt.Errorf("ListSnapshots failed: %w", err)
	}

	return &cmtabci.ResponseListSnapshots{
		Snapshots: res.Snapshots,
	}, nil
}

func (c *grpcClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestOfferSnapshot{
		Snapshot: req.Snapshot,
		AppHash:  req.AppHash,
	}

	res, err := c.client.OfferSnapshot(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("OfferSnapshot failed", "err", err)
		}
		return nil, fmt.Errorf("OfferSnapshot failed: %w", err)
	}

	return &cmtabci.ResponseOfferSnapshot{
		Result: cmtabci.ResponseOfferSnapshot_Result(res.Result),
	}, nil
}

func (c *grpcClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestLoadSnapshotChunk{
		Height: req.Height,
		Format: req.Format,
		Chunk:  req.Chunk,
	}

	res, err := c.client.LoadSnapshotChunk(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("LoadSnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("LoadSnapshotChunk failed: %w", err)
	}

	return &cmtabci.ResponseLoadSnapshotChunk{
		Chunk: res.Chunk,
	}, nil
}

func (c *grpcClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	protoReq := &cmtproto.RequestApplySnapshotChunk{
		Index:  req.Index,
		Chunk:  req.Chunk,
		Sender: req.Sender,
	}

	res, err := c.client.ApplySnapshotChunk(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ApplySnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("ApplySnapshotChunk failed: %w", err)
	}

	return &cmtabci.ResponseApplySnapshotChunk{
		Result:        cmtabci.ResponseApplySnapshotChunk_Result(res.Result),
		RefetchChunks: res.RefetchChunks,
		RejectSenders: res.RejectSenders,
	}, nil
} 