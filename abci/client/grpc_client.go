package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ Client = (*grpcClient)(nil)

type grpcClient struct {
	client cmtabci.ABCIServiceClient
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

	client := cmtabci.NewABCIServiceClient(conn)
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

	res, err := c.client.CheckTx(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("CheckTx failed", "err", err)
		}
		return nil, fmt.Errorf("CheckTx failed: %w", err)
	}

	return res, nil
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

	res, err := c.client.FinalizeBlock(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("FinalizeBlock failed", "err", err)
		}
		return nil, fmt.Errorf("FinalizeBlock failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if req.MaxTxBytes <= 0 {
		return nil, fmt.Errorf("invalid max tx bytes: %d", req.MaxTxBytes)
	}

	res, err := c.client.PrepareProposal(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("PrepareProposal failed", "err", err)
		}
		return nil, fmt.Errorf("PrepareProposal failed: %w", err)
	}

	// Validate response
	if len(res.Txs) > int(req.MaxTxBytes) {
		return nil, fmt.Errorf("proposal exceeds max tx bytes")
	}

	return res, nil
}

func (c *grpcClient) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ProcessProposal validation failed: %w", err)
	}

	res, err := c.client.ProcessProposal(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ProcessProposal failed", "err", err)
		}
		return nil, fmt.Errorf("ProcessProposal failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ExtendVote validation failed: %w", err)
	}

	res, err := c.client.ExtendVote(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ExtendVote failed", "err", err)
		}
		return nil, fmt.Errorf("ExtendVote failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("VerifyVoteExtension validation failed: %w", err)
	}

	res, err := c.client.VerifyVoteExtension(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("VerifyVoteExtension failed", "err", err)
		}
		return nil, fmt.Errorf("VerifyVoteExtension failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.Commit(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Commit failed", "err", err)
		}
		return nil, fmt.Errorf("Commit failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.InitChain(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("InitChain failed", "err", err)
		}
		return nil, fmt.Errorf("InitChain failed: %w", err)
	}

	return res, nil
}

// Query methods
func (c *grpcClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.Info(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Info failed", "err", err)
		}
		return nil, fmt.Errorf("Info failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.Query(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Query failed", "err", err)
		}
		return nil, fmt.Errorf("Query failed: %w", err)
	}

	return res, nil
}

// Snapshot methods
func (c *grpcClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.ListSnapshots(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ListSnapshots failed", "err", err)
		}
		return nil, fmt.Errorf("ListSnapshots failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.OfferSnapshot(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("OfferSnapshot failed", "err", err)
		}
		return nil, fmt.Errorf("OfferSnapshot failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.LoadSnapshotChunk(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("LoadSnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("LoadSnapshotChunk failed: %w", err)
	}

	return res, nil
}

func (c *grpcClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	ctx, cancel := contextWithTimeout(ctx, 0)
	defer cancel()

	res, err := c.client.ApplySnapshotChunk(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ApplySnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("ApplySnapshotChunk failed: %w", err)
	}

	return res, nil
} 