package client

import (
	"context"
	"fmt"
	"sync"

	cometbftabci "github.com/cometbft/cometbft/abci/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/abci"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ Client = (*grpcClient)(nil)

type grpcClient struct {
	client cometbftproto.ABCIServiceClient
	conn   *grpc.ClientConn
	mtx    sync.Mutex
	logger Logger
}

// NewGRPCClient creates a new gRPC client
func NewGRPCClient(addr string, logger Logger) (Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &grpcClient{
		client: cometbftproto.NewABCIServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
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

// Echo method for testing
func (c *grpcClient) Echo(ctx context.Context, msg string) (*cometbftabci.ResponseEcho, error) {
	resp, err := c.client.Echo(ctx, &cometbftproto.RequestEcho{Message: msg})
	if err != nil {
		return nil, err
	}
	return &cometbftabci.ResponseEcho{Message: resp.Message}, nil
}

// Flush method
func (c *grpcClient) Flush(ctx context.Context) error {
	_, err := c.client.Flush(ctx, &cometbftproto.RequestFlush{})
	return err
}

// Mempool methods
func (c *grpcClient) CheckTx(ctx context.Context, req *cometbftabci.RequestCheckTx) (*cometbftabci.ResponseCheckTx, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestCheckTx{
		Tx:   req.Tx,
		Type: cometbftproto.CheckTxType(req.Type),
	}

	resp, err := c.client.CheckTx(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("CheckTx failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC CheckTx failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseCheckTx{
		Code:      resp.Code,
		Data:      resp.Data,
		Log:       resp.Log,
		Info:      resp.Info,
		GasWanted: resp.GasWanted,
		GasUsed:   resp.GasUsed,
		Events:    resp.Events,
		Codespace: resp.Codespace,
		Sender:    resp.Sender,
		Priority:  resp.Priority,
		MempoolError: resp.MempoolError,
	}, nil
}

func (c *grpcClient) CheckTxAsync(ctx context.Context, req *cometbftabci.RequestCheckTx) *ReqRes {
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.CheckTx(ctx, req)
		if err != nil {
			reqRes.Response = nil
			reqRes.Error = err
		} else {
			reqRes.Response = res
		}
		reqRes.Done()
	}()
	return reqRes
}

// Consensus methods
func (c *grpcClient) FinalizeBlock(ctx context.Context, req *cometbftabci.RequestFinalizeBlock) (*cometbftabci.ResponseFinalizeBlock, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestFinalizeBlock{
		Txs:                req.Txs,
		DecidedLastCommit:  req.DecidedLastCommit,
		Misbehavior:        req.Misbehavior,
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}

	resp, err := c.client.FinalizeBlock(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("FinalizeBlock failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC FinalizeBlock failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseFinalizeBlock{
		Events:                resp.Events,
		TxResults:             resp.TxResults,
		ValidatorUpdates:      resp.ValidatorUpdates,
		ConsensusParamUpdates: resp.ConsensusParamUpdates,
		AppHash:               resp.AppHash,
		RetainHeight:          resp.RetainHeight,
	}, nil
}

func (c *grpcClient) FinalizeBlockAsync(ctx context.Context, req *cometbftabci.RequestFinalizeBlock) *ReqRes {
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.FinalizeBlock(ctx, req)
		if err != nil {
			reqRes.Response = nil
			reqRes.Error = err
		} else {
			reqRes.Response = res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (c *grpcClient) PrepareProposal(ctx context.Context, req *cometbftabci.RequestPrepareProposal) (*cometbftabci.ResponsePrepareProposal, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestPrepareProposal{
		MaxTxBytes:    req.MaxTxBytes,
		Txs:           req.Txs,
		LocalLastCommit: req.LocalLastCommit,
		Misbehavior:   req.Misbehavior,
		Height:        req.Height,
		Time:          req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}

	resp, err := c.client.PrepareProposal(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("PrepareProposal failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC PrepareProposal failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponsePrepareProposal{
		Txs: resp.Txs,
	}, nil
}

func (c *grpcClient) ProcessProposal(ctx context.Context, req *cometbftabci.RequestProcessProposal) (*cometbftabci.ResponseProcessProposal, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestProcessProposal{
		Txs:                req.Txs,
		ProposedLastCommit: req.ProposedLastCommit,
		Misbehavior:        req.Misbehavior,
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}

	resp, err := c.client.ProcessProposal(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ProcessProposal failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ProcessProposal failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseProcessProposal{
		Status: resp.Status,
	}, nil
}

func (c *grpcClient) ExtendVote(ctx context.Context, req *cometbftabci.RequestExtendVote) (*cometbftabci.ResponseExtendVote, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestExtendVote{
		Hash:   req.Hash,
		Height: req.Height,
	}

	resp, err := c.client.ExtendVote(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ExtendVote failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ExtendVote failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseExtendVote{
		VoteExtension: resp.VoteExtension,
	}, nil
}

func (c *grpcClient) VerifyVoteExtension(ctx context.Context, req *cometbftabci.RequestVerifyVoteExtension) (*cometbftabci.ResponseVerifyVoteExtension, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestVerifyVoteExtension{
		Hash:             req.Hash,
		ValidatorAddress: req.ValidatorAddress,
		Height:           req.Height,
		VoteExtension:    req.VoteExtension,
	}

	resp, err := c.client.VerifyVoteExtension(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("VerifyVoteExtension failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC VerifyVoteExtension failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseVerifyVoteExtension{
		Status: resp.Status,
	}, nil
}

func (c *grpcClient) Commit(ctx context.Context) (*cometbftabci.ResponseCommit, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	resp, err := c.client.Commit(ctx, &cometbftproto.RequestCommit{})
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Commit failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Commit failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseCommit{
		Data:         resp.Data,
		RetainHeight: resp.RetainHeight,
	}, nil
}

func (c *grpcClient) CommitAsync(ctx context.Context) *ReqRes {
	reqRes := NewReqRes(nil)
	go func() {
		res, err := c.Commit(ctx)
		if err != nil {
			reqRes.Response = nil
			reqRes.Error = err
		} else {
			reqRes.Response = res
		}
		reqRes.Done()
	}()
	return reqRes
}

func (c *grpcClient) InitChain(ctx context.Context, req *cometbftabci.RequestInitChain) (*cometbftabci.ResponseInitChain, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestInitChain{
		Time:            req.Time,
		ChainId:         req.ChainId,
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppStateBytes:   req.AppStateBytes,
		InitialHeight:   req.InitialHeight,
	}

	resp, err := c.client.InitChain(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("InitChain failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC InitChain failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseInitChain{
		ConsensusParams: resp.ConsensusParams,
		Validators:      resp.Validators,
		AppHash:         resp.AppHash,
	}, nil
}

func (c *grpcClient) Info(ctx context.Context, req *cometbftabci.RequestInfo) (*cometbftabci.ResponseInfo, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestInfo{
		Version:      req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion:   req.P2PVersion,
		AbciVersion:  req.AbciVersion,
	}

	resp, err := c.client.Info(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Info failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Info failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseInfo{
		Data:             resp.Data,
		Version:          resp.Version,
		AppVersion:       resp.AppVersion,
		LastBlockHeight:  resp.LastBlockHeight,
		LastBlockAppHash: resp.LastBlockAppHash,
	}, nil
}

func (c *grpcClient) Query(ctx context.Context, req *cometbftabci.RequestQuery) (*cometbftabci.ResponseQuery, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestQuery{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}

	resp, err := c.client.Query(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Query failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Query failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseQuery{
		Code:   resp.Code,
		Log:    resp.Log,
		Info:   resp.Info,
		Index:  resp.Index,
		Key:    resp.Key,
		Value:  resp.Value,
		ProofOps: resp.ProofOps,
		Height: resp.Height,
		Codespace: resp.Codespace,
	}, nil
}

func (c *grpcClient) ListSnapshots(ctx context.Context, req *cometbftabci.RequestListSnapshots) (*cometbftabci.ResponseListSnapshots, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.ListSnapshots(ctx, &cometbftproto.RequestListSnapshots{})
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ListSnapshots failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC ListSnapshots failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseListSnapshots{
		Snapshots: resp.Snapshots,
	}, nil
}

func (c *grpcClient) OfferSnapshot(ctx context.Context, req *cometbftabci.RequestOfferSnapshot) (*cometbftabci.ResponseOfferSnapshot, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestOfferSnapshot{
		Snapshot: req.Snapshot,
		AppHash:  req.AppHash,
	}

	resp, err := c.client.OfferSnapshot(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("OfferSnapshot failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC OfferSnapshot failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseOfferSnapshot{
		Result: resp.Result,
	}, nil
}

func (c *grpcClient) LoadSnapshotChunk(ctx context.Context, req *cometbftabci.RequestLoadSnapshotChunk) (*cometbftabci.ResponseLoadSnapshotChunk, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestLoadSnapshotChunk{
		Height: req.Height,
		Format: req.Format,
		Chunk:  req.Chunk,
	}

	resp, err := c.client.LoadSnapshotChunk(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("LoadSnapshotChunk failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC LoadSnapshotChunk failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseLoadSnapshotChunk{
		Chunk: resp.Chunk,
	}, nil
}

func (c *grpcClient) ApplySnapshotChunk(ctx context.Context, req *cometbftabci.RequestApplySnapshotChunk) (*cometbftabci.ResponseApplySnapshotChunk, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Convert to proto request
	protoReq := &cometbftproto.RequestApplySnapshotChunk{
		Index:  req.Index,
		Chunk:  req.Chunk,
		Sender: req.Sender,
	}

	resp, err := c.client.ApplySnapshotChunk(ctx, protoReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ApplySnapshotChunk failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC ApplySnapshotChunk failed: %w", err)
	}

	// Convert back to ABCI response
	return &cometbftabci.ResponseApplySnapshotChunk{
		Result:        resp.Result,
		RefetchChunks: resp.RefetchChunks,
		RejectSenders: resp.RejectSenders,
	}, nil
}

func (c *grpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
} 