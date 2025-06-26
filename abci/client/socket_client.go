package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"
)

var _ Client = (*socketClient)(nil)

type socketClient struct {
	conn     net.Conn
	mtx      sync.Mutex
	callback Callback
	logger   Logger
}

func NewSocketClient(addr string, mustConnect bool) Client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		if mustConnect {
			panic(fmt.Sprintf("failed to connect to socket server: %v", err))
		}
		return &socketClient{conn: nil}
	}

	return &socketClient{
		conn: conn,
	}
}

func (c *socketClient) SetResponseCallback(cb Callback) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.callback = cb
}

func (c *socketClient) SetLogger(l Logger) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.logger = l
}

func (c *socketClient) Error() error {
	if c.conn == nil {
		return ErrConnectionNotInitialized
	}
	return nil
}

// Helper methods for socket communication
func (c *socketClient) sendRequest(method string, reqBytes []byte) error {
	// Send length prefix
	length := uint32(len(reqBytes))
	if err := binary.Write(c.conn, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	// Send request
	if _, err := c.conn.Write(reqBytes); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

func (c *socketClient) receiveResponse() ([]byte, error) {
	// Read length prefix
	var length uint32
	if err := binary.Read(c.conn, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}

	// Read response
	resBytes := make([]byte, length)
	if _, err := c.conn.Read(resBytes); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return resBytes, nil
}

// Mempool methods
func (c *socketClient) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if err := validateTxData(req.Tx); err != nil {
		return nil, fmt.Errorf("CheckTx validation failed: %w", err)
	}

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("CheckTx", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseCheckTx
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) CheckTxAsync(ctx context.Context, req *cmtabci.RequestCheckTx) *ReqRes {
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

func (c *socketClient) Flush(ctx context.Context) error {
	// Socket client doesn't need explicit flushing
	return nil
}

// Consensus methods
func (c *socketClient) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("FinalizeBlock validation failed: %w", err)
	}

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("FinalizeBlock", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseFinalizeBlock
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if req.MaxTxBytes <= 0 {
		return nil, fmt.Errorf("invalid max tx bytes: %d", req.MaxTxBytes)
	}

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("PrepareProposal", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponsePrepareProposal
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ProcessProposal validation failed: %w", err)
	}

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("ProcessProposal", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseProcessProposal
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("ExtendVote validation failed: %w", err)
	}

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("ExtendVote", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseExtendVote
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if err := validateBlockHeight(req.Height); err != nil {
		return nil, fmt.Errorf("VerifyVoteExtension validation failed: %w", err)
	}

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("VerifyVoteExtension", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseVerifyVoteExtension
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("Commit", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseCommit
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("InitChain", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseInitChain
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

// Query methods
func (c *socketClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("Info", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseInfo
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("Query", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseQuery
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

// Snapshot methods
func (c *socketClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("ListSnapshots", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseListSnapshots
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("OfferSnapshot", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseOfferSnapshot
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("LoadSnapshotChunk", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseLoadSnapshotChunk
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
}

func (c *socketClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Serialize request
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}

	// Send request
	if err := c.sendRequest("ApplySnapshotChunk", reqBytes); err != nil {
		return nil, err
	}

	// Receive response
	resBytes, err := c.receiveResponse()
	if err != nil {
		return nil, err
	}

	// Deserialize
	var res cmtabci.ResponseApplySnapshotChunk
	if err := proto.Unmarshal(resBytes, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &res, nil
} 