package client

import (
	"context"
	"fmt"
	"sync"

	cometbftabciv1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ Client = (*grpcClient)(nil)

type grpcClient struct {
	client cometbftabciv1.ABCIServiceClient
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
		client: cometbftabciv1.NewABCIServiceClient(conn),
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
func (c *grpcClient) Echo(ctx context.Context, msg string) (*cometbftabciv1.EchoResponse, error) {
	return c.client.Echo(ctx, &cometbftabciv1.EchoRequest{Message: msg})
}

// Flush method
func (c *grpcClient) Flush(ctx context.Context) error {
	_, err := c.client.Flush(ctx, &cometbftabciv1.FlushRequest{})
	return err
}

// Mempool methods
func (c *grpcClient) CheckTx(ctx context.Context, req *cometbftabciv1.CheckTxRequest) (*cometbftabciv1.CheckTxResponse, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.CheckTx(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("CheckTx failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC CheckTx failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) CheckTxAsync(ctx context.Context, req *cometbftabciv1.CheckTxRequest) *ReqRes {
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
func (c *grpcClient) FinalizeBlock(ctx context.Context, req *cometbftabciv1.FinalizeBlockRequest) (*cometbftabciv1.FinalizeBlockResponse, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.FinalizeBlock(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("FinalizeBlock failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC FinalizeBlock failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) FinalizeBlockAsync(ctx context.Context, req *cometbftabciv1.FinalizeBlockRequest) *ReqRes {
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

func (c *grpcClient) PrepareProposal(ctx context.Context, req *cometbftabciv1.PrepareProposalRequest) (*cometbftabciv1.PrepareProposalResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.PrepareProposal(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("PrepareProposal failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC PrepareProposal failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) ProcessProposal(ctx context.Context, req *cometbftabciv1.ProcessProposalRequest) (*cometbftabciv1.ProcessProposalResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.ProcessProposal(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ProcessProposal failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ProcessProposal failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) ExtendVote(ctx context.Context, req *cometbftabciv1.ExtendVoteRequest) (*cometbftabciv1.ExtendVoteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.ExtendVote(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ExtendVote failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ExtendVote failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) VerifyVoteExtension(ctx context.Context, req *cometbftabciv1.VerifyVoteExtensionRequest) (*cometbftabciv1.VerifyVoteExtensionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.VerifyVoteExtension(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("VerifyVoteExtension failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC VerifyVoteExtension failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) Commit(ctx context.Context) (*cometbftabciv1.CommitResponse, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	resp, err := c.client.Commit(ctx, &cometbftabciv1.CommitRequest{})
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Commit failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Commit failed: %w", err)
	}

	return resp, nil
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

func (c *grpcClient) InitChain(ctx context.Context, req *cometbftabciv1.InitChainRequest) (*cometbftabciv1.InitChainResponse, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.InitChain(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("InitChain failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC InitChain failed: %w", err)
	}

	return resp, nil
}

// Query methods
func (c *grpcClient) Info(ctx context.Context, req *cometbftabciv1.InfoRequest) (*cometbftabciv1.InfoResponse, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.Info(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Info failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Info failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) Query(ctx context.Context, req *cometbftabciv1.QueryRequest) (*cometbftabciv1.QueryResponse, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.Query(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Query failed", "error", err)
		}
		return nil, fmt.Errorf("gRPC Query failed: %w", err)
	}

	return resp, nil
}

// Snapshot methods
func (c *grpcClient) ListSnapshots(ctx context.Context, req *cometbftabciv1.ListSnapshotsRequest) (*cometbftabciv1.ListSnapshotsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.ListSnapshots(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ListSnapshots failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ListSnapshots failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) OfferSnapshot(ctx context.Context, req *cometbftabciv1.OfferSnapshotRequest) (*cometbftabciv1.OfferSnapshotResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.OfferSnapshot(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("OfferSnapshot failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC OfferSnapshot failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) LoadSnapshotChunk(ctx context.Context, req *cometbftabciv1.LoadSnapshotChunkRequest) (*cometbftabciv1.LoadSnapshotChunkResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.LoadSnapshotChunk(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("LoadSnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC LoadSnapshotChunk failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) ApplySnapshotChunk(ctx context.Context, req *cometbftabciv1.ApplySnapshotChunkRequest) (*cometbftabciv1.ApplySnapshotChunkResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resp, err := c.client.ApplySnapshotChunk(ctx, req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("ApplySnapshotChunk failed", "err", err)
		}
		return nil, fmt.Errorf("gRPC ApplySnapshotChunk failed: %w", err)
	}

	return resp, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
} 