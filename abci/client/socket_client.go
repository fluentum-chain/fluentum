package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/gogo/protobuf/proto"
)

var _ Client = (*socketClient)(nil)

type socketClient struct {
	conn      net.Conn
	mtx       sync.Mutex
	reqQueue  map[uint64]*ReqRes
	nextReqID uint64
	logger    Logger
	closed    bool
	err       error
}

func NewSocketClient(conn net.Conn, logger Logger) Client {
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
	defer c.removeRequest(reqRes.ID)

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
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.CheckTx(ctx, req)
		if err != nil {
			reqRes.Error = err
		} else {
			reqRes.Response = res
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
	defer c.removeRequest(reqRes.ID)

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
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.FinalizeBlock(ctx, req)
		if err != nil {
			reqRes.Error = err
		} else {
			reqRes.Response = res
		}
		reqRes.Done()
	}()
	return reqRes
}

// Commit implements the Commit ABCI method
func (c *socketClient) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	c.mtx.Lock()
	if c.closed {
		c.mtx.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mtx.Unlock()

	reqRes := c.queueRequest(&cmtabci.Request{
		Value: &cmtabci.Request_Commit{Commit: req},
	})
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
func (c *socketClient) CommitAsync(ctx context.Context, req *cmtabci.RequestCommit) *ReqRes {
	reqRes := NewReqRes(req)
	go func() {
		res, err := c.Commit(ctx, req)
		if err != nil {
			reqRes.Error = err
		} else {
			reqRes.Response = res
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
	defer c.removeRequest(reqRes.ID)

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
	defer c.removeRequest(reqRes.ID)

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
	defer c.removeRequest(reqRes.ID)

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
	defer c.removeRequest(reqRes.ID)

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
	defer c.removeRequest(reqRes.ID)

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
	defer c.removeRequest(reqRes.ID)

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
	
	// Close all pending requests
	for _, reqRes := range c.reqQueue {
		reqRes.ErrorCh <- fmt.Errorf("client closed")
		close(reqRes.ResponseCh)
		close(reqRes.ErrorCh)
	}
	
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

		// Dispatch response
		if reqRes := c.getRequest(res.RequestID); reqRes != nil {
			select {
			case reqRes.ResponseCh <- res.Value:
			default:
				if c.logger != nil {
					c.logger.Error("Response channel full, dropping response")
				}
			}
		} else {
			if c.logger != nil {
				c.logger.Error("No request found for response", "requestID", res.RequestID)
			}
		}
	}
}

// Error returns the current error state of the client
func (c *socketClient) Error() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.err
} 