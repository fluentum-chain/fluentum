package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	tmlog "github.com/fluentum-chain/fluentum/libs/log"
	"github.com/gogo/protobuf/proto"
)

var _ Client = (*socketClient)(nil)

type socketClient struct {
	conn      net.Conn
	mtx       sync.Mutex
	reqQueue  map[uint64]*ReqRes
	nextReqID uint64
	logger    tmlog.Logger
	closed    bool
	err       error
}

func NewSocketClient(conn net.Conn, logger tmlog.Logger) Client {
	client := &socketClient{
		conn:     conn,
		reqQueue: make(map[uint64]*ReqRes),
		logger:   logger,
	}
	go client.recvRoutine()
	return client
}

// CheckTx implements the CheckTx ABCI method
func (c *socketClient) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_CheckTx{CheckTx: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if checkTxRes, ok := res.(*cmtabci.ResponseCheckTx); ok {
			return checkTxRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// CheckTxAsync implements async CheckTx
func (c *socketClient) CheckTxAsync(ctx context.Context, req *cmtabci.RequestCheckTx) *ReqRes {
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_CheckTx{CheckTx: req}})
	go func() {
		res, err := c.CheckTx(ctx, req)
		if err != nil {
			reqRes.Error = err
		} else {
			reqRes.ResponseCh <- res
		}
		reqRes.Done()
	}()
	return reqRes
}

// FinalizeBlock implements the FinalizeBlock ABCI method
func (c *socketClient) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_FinalizeBlock{FinalizeBlock: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if finalizeRes, ok := res.(*cmtabci.ResponseFinalizeBlock); ok {
			return finalizeRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// FinalizeBlockAsync implements async FinalizeBlock
func (c *socketClient) FinalizeBlockAsync(ctx context.Context, req *cmtabci.RequestFinalizeBlock) *ReqRes {
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_FinalizeBlock{FinalizeBlock: req}})
	go func() {
		res, err := c.FinalizeBlock(ctx, req)
		if err != nil {
			reqRes.Error = err
		} else {
			reqRes.ResponseCh <- res
		}
		reqRes.Done()
	}()
	return reqRes
}

// Commit implements the Commit ABCI method
func (c *socketClient) Commit(ctx context.Context) (*cmtabci.ResponseCommit, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	req := &cmtabci.Request{
		Value: &cmtabci.Request_Commit{Commit: &cmtabci.RequestCommit{}},
	}
	reqRes := c.queueRequest(req)
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if commitRes, ok := res.(*cmtabci.ResponseCommit); ok {
			return commitRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// CommitAsync implements async Commit
func (c *socketClient) CommitAsync(ctx context.Context) *ReqRes {
	req := &cmtabci.Request{
		Value: &cmtabci.Request_Commit{Commit: &cmtabci.RequestCommit{}},
	}
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.Commit(ctx)
		if err != nil {
			reqRes.Error = err
		} else {
			reqRes.ResponseCh <- res
		}
		reqRes.Done()
	}()
	return reqRes
}

// Info implements the Info ABCI method
func (c *socketClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_Info{Info: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if infoRes, ok := res.(*cmtabci.ResponseInfo); ok {
			return infoRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// Query implements the Query ABCI method
func (c *socketClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_Query{Query: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if queryRes, ok := res.(*cmtabci.ResponseQuery); ok {
			return queryRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// InitChain implements the InitChain ABCI method
func (c *socketClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_InitChain{InitChain: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if initRes, ok := res.(*cmtabci.ResponseInitChain); ok {
			return initRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// PrepareProposal implements the PrepareProposal ABCI method
func (c *socketClient) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_PrepareProposal{PrepareProposal: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if prepRes, ok := res.(*cmtabci.ResponsePrepareProposal); ok {
			return prepRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// ProcessProposal implements the ProcessProposal ABCI method
func (c *socketClient) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_ProcessProposal{ProcessProposal: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if procRes, ok := res.(*cmtabci.ResponseProcessProposal); ok {
			return procRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// ExtendVote implements the ExtendVote ABCI method
func (c *socketClient) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_ExtendVote{ExtendVote: req},
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if extRes, ok := res.(*cmtabci.ResponseExtendVote); ok {
			return extRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// VerifyVoteExtension implements the VerifyVoteExtension ABCI method
func (c *socketClient) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_VerifyVoteExtension{VerifyVoteExtension: req},
	})
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if verifyRes, ok := res.(*cmtabci.ResponseVerifyVoteExtension); ok {
			return verifyRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// ListSnapshots implements the ListSnapshots ABCI method
func (c *socketClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_ListSnapshots{ListSnapshots: req},
	})
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if listRes, ok := res.(*cmtabci.ResponseListSnapshots); ok {
			return listRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// OfferSnapshot implements the OfferSnapshot ABCI method
func (c *socketClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_OfferSnapshot{OfferSnapshot: req},
	})
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if offerRes, ok := res.(*cmtabci.ResponseOfferSnapshot); ok {
			return offerRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// LoadSnapshotChunk implements the LoadSnapshotChunk ABCI method
func (c *socketClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_LoadSnapshotChunk{LoadSnapshotChunk: req},
	})
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if loadRes, ok := res.(*cmtabci.ResponseLoadSnapshotChunk); ok {
			return loadRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// ApplySnapshotChunk implements the ApplySnapshotChunk ABCI method
func (c *socketClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_ApplySnapshotChunk{ApplySnapshotChunk: req},
	})
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if applyRes, ok := res.(*cmtabci.ResponseApplySnapshotChunk); ok {
			return applyRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// Close closes the socket client
func (c *socketClient) Close() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

// Core socket handling methods

func (c *socketClient) queueRequest(req *cmtabci.Request) *ReqRes {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.nextReqID++
	reqID := c.nextReqID

	reqRes := NewReqRes(req)
	reqRes.ID = reqID
	c.reqQueue[reqID] = reqRes

	// Send the request
	if err := c.sendRequest(req); err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to send request", "error", err)
		}
		reqRes.ErrorCh <- err
	}

	return reqRes
}

func (c *socketClient) removeRequest(reqID uint64) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if reqRes, exists := c.reqQueue[reqID]; exists {
		close(reqRes.ResponseCh)
		close(reqRes.ErrorCh)
		delete(c.reqQueue, reqID)
	}
}

func (c *socketClient) getRequest(reqID uint64) *ReqRes {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.reqQueue[reqID]
}

func (c *socketClient) sendRequest(req *cmtabci.Request) error {
	// Marshal the request
	msg, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write message length (4 bytes big-endian)
	msgLen := uint32(len(msg))
	if err := binary.Write(c.conn, binary.BigEndian, msgLen); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// Write message
	if _, err = c.conn.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (c *socketClient) recvRoutine() {
	defer func() {
		c.mtx.Lock()
		c.closed = true
		c.mtx.Unlock()
	}()

	for {
		// Read message length
		var msgLen uint32
		if err := binary.Read(c.conn, binary.BigEndian, &msgLen); err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to read message length", "error", err)
			}
			break
		}

		// Read message
		msg := make([]byte, msgLen)
		if _, err := io.ReadFull(c.conn, msg); err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to read message", "error", err)
			}
			break
		}

		// Unmarshal response
		var res cmtabci.Response
		if err := proto.Unmarshal(msg, &res); err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to unmarshal response", "error", err)
			}
			continue
		}

		// For now, we'll use a simple approach - find the first request in queue
		// In a real implementation, you'd need to track request IDs properly
		c.mtx.Lock()
		for reqID, reqRes := range c.reqQueue {
			select {
			case reqRes.ResponseCh <- res.Value:
				delete(c.reqQueue, reqID)
			default:
				if c.logger != nil {
					c.logger.Error("Response channel full, dropping response")
				}
			}
			break // Process only the first request for now
		}
		c.mtx.Unlock()
	}
}

// Error returns the current error state of the client
func (c *socketClient) Error() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.err
}

func (c *socketClient) SetLogger(logger tmlog.Logger) {
	c.logger = logger
}

func (c *socketClient) Echo(ctx context.Context, msg string) (*cmtabci.ResponseEcho, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	req := &cmtabci.Request{
		Value: &cmtabci.Request_Echo{Echo: &cmtabci.RequestEcho{Message: msg}},
	}
	reqRes := c.queueRequest(req)
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-reqRes.ResponseCh:
		if echoRes, ok := res.(*cmtabci.ResponseEcho); ok {
			return echoRes, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case err := <-reqRes.ErrorCh:
		return nil, err
	}
}

// Flush implements the Flush ABCI method
func (c *socketClient) Flush(ctx context.Context) error {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	req := &cmtabci.Request{
		Value: &cmtabci.Request_Flush{Flush: &cmtabci.RequestFlush{}},
	}
	reqRes := c.queueRequest(req)
	defer c.removeRequest(reqRes.ID)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-reqRes.ResponseCh:
		return nil
	case err := <-reqRes.ErrorCh:
		return err
	}
}

// FlushAsync implements async Flush
func (c *socketClient) FlushAsync(ctx context.Context) *ReqRes {
	req := &cmtabci.Request{
		Value: &cmtabci.Request_Flush{Flush: &cmtabci.RequestFlush{}},
	}
	reqRes := NewReqRes(req)
	go func() {
		err := c.Flush(ctx)
		if err != nil {
			reqRes.Error = err
		} else {
			reqRes.ResponseCh <- &cmtabci.ResponseFlush{}
		}
		reqRes.Done()
	}()
	return reqRes
}

func (c *socketClient) SetResponseCallback(cb Callback) {
	// Not implemented for socket client
}

// Start starts the client
func (c *socketClient) Start() error {
	// Socket client is already started when created
	return nil
}

// Stop stops the client
func (c *socketClient) Stop() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Quit returns a channel that is closed when the client is stopped
func (c *socketClient) Quit() <-chan struct{} {
	// Create a simple channel that's closed when Stop is called
	// For now, return a channel that's never closed since we don't have proper lifecycle management
	ch := make(chan struct{})
	// TODO: Implement proper lifecycle management
	return ch
}
