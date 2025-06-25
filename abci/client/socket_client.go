package client

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"

	types "github.com/fluentum-chain/fluentum/abci/types"
	tmnet "github.com/fluentum-chain/fluentum/libs/net"
	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
	"github.com/fluentum-chain/fluentum/libs/timer"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

const (
	reqQueueSize    = 256 // TODO make configurable
	flushThrottleMS = 20  // Don't wait longer than...
)

// This is goroutine-safe, but users should beware that the application in
// general is not meant to be interfaced with concurrent callers.
type socketClient struct {
	service.BaseService

	addr        string
	mustConnect bool
	conn        net.Conn

	reqQueue   chan *ReqRes
	flushTimer *timer.ThrottleTimer

	mtx     tmsync.Mutex
	err     error
	reqSent *list.List                          // list of requests sent, waiting for response
	resCb   func(*abci.Request, *abci.Response) // called on all requests, if set.
}

var _ Client = (*socketClient)(nil)

// NewSocketClient creates a new socket client, which connects to a given
// address. If mustConnect is true, the client will return an error upon start
// if it fails to connect.
func NewSocketClient(addr string, mustConnect bool) Client {
	cli := &socketClient{
		reqQueue:    make(chan *ReqRes, reqQueueSize),
		flushTimer:  timer.NewThrottleTimer("socketClient", flushThrottleMS),
		mustConnect: mustConnect,

		addr:    addr,
		reqSent: list.New(),
		resCb:   nil,
	}
	cli.BaseService = *service.NewBaseService(nil, "socketClient", cli)
	return cli
}

// OnStart implements Service by connecting to the server and spawning reading
// and writing goroutines.
func (cli *socketClient) OnStart() error {
	var (
		err  error
		conn net.Conn
	)

	for {
		conn, err = tmnet.Connect(cli.addr)
		if err != nil {
			if cli.mustConnect {
				return err
			}
			cli.Logger.Error(fmt.Sprintf("abci.socketClient failed to connect to %v.  Retrying after %vs...",
				cli.addr, dialRetryIntervalSeconds), "err", err)
			time.Sleep(time.Second * dialRetryIntervalSeconds)
			continue
		}
		cli.conn = conn

		go cli.sendRequestsRoutine(conn)
		go cli.recvResponseRoutine(conn)

		return nil
	}
}

// OnStop implements Service by closing connection and flushing all queues.
func (cli *socketClient) OnStop() {
	if cli.conn != nil {
		cli.conn.Close()
	}

	cli.flushQueue()
	cli.flushTimer.Stop()
}

// Error returns an error if the client was stopped abruptly.
func (cli *socketClient) Error() error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	return cli.err
}

// SetResponseCallback sets a callback, which will be executed for each
// non-error & non-empty response from the server.
//
// NOTE: callback may get internally generated flush responses.
func (cli *socketClient) SetResponseCallback(resCb Callback) {
	cli.mtx.Lock()
	cli.resCb = resCb
	cli.mtx.Unlock()
}

//----------------------------------------

func (cli *socketClient) sendRequestsRoutine(conn io.Writer) {
	w := bufio.NewWriter(conn)
	for {
		select {
		case reqres := <-cli.reqQueue:
			// cli.Logger.Debug("Sent request", "requestType", reflect.TypeOf(reqres.Request), "request", reqres.Request)

			cli.willSendReq(reqres)
			err := types.WriteMessage(reqres.Request, w)
			if err != nil {
				cli.stopForError(fmt.Errorf("write to buffer: %w", err))
				return
			}

			// If it's a flush request, flush the current buffer.
			if _, ok := reqres.Request.Value.(*abci.Request_Flush); ok {
				err = w.Flush()
				if err != nil {
					cli.stopForError(fmt.Errorf("flush buffer: %w", err))
					return
				}
			}
		case <-cli.flushTimer.Ch: // flush queue
			select {
			case cli.reqQueue <- NewReqRes(&abci.Request{
				Value: &abci.Request_Flush{Flush: &abci.RequestFlush{}},
			}):
			default:
				// Probably will fill the buffer, or retry later.
			}
		case <-cli.Quit():
			return
		}
	}
}

func (cli *socketClient) recvResponseRoutine(conn io.Reader) {
	r := bufio.NewReader(conn)
	for {
		var res = &abci.Response{}
		err := types.ReadMessage(r, res)
		if err != nil {
			cli.stopForError(fmt.Errorf("read message: %w", err))
			return
		}

		// cli.Logger.Debug("Received response", "responseType", reflect.TypeOf(res), "response", res)

		switch r := res.Value.(type) {
		case *abci.Response_Exception: // app responded with error
			// XXX After setting cli.err, release waiters (e.g. reqres.Done())
			cli.stopForError(errors.New(r.Exception.Error))
			return
		default:
			err := cli.didRecvResponse(res)
			if err != nil {
				cli.stopForError(err)
				return
			}
		}
	}
}

func (cli *socketClient) willSendReq(reqres *ReqRes) {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	cli.reqSent.PushBack(reqres)
}

func (cli *socketClient) didRecvResponse(res *abci.Response) error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// Get the first ReqRes.
	next := cli.reqSent.Front()
	if next == nil {
		return fmt.Errorf("unexpected %v when nothing expected", reflect.TypeOf(res.Value))
	}

	reqres := next.Value.(*ReqRes)
	if !resMatchesReq(reqres.Request, res) {
		return fmt.Errorf("unexpected %v when response to %v expected",
			reflect.TypeOf(res.Value), reflect.TypeOf(reqres.Request.Value))
	}

	reqres.Response = res
	reqres.Done()            // release waiters
	cli.reqSent.Remove(next) // pop first item from linked list

	// Notify client listener if set (global callback).
	if cli.resCb != nil {
		cli.resCb(reqres.Request, res)
	}

	// Notify reqRes listener if set (request specific callback).
	//
	// NOTE: It is possible this callback isn't set on the reqres object. At this
	// point, in which case it will be called after, when it is set.
	reqres.InvokeCallback()

	return nil
}

//----------------------------------------

// EchoAsync sends an async Echo request
func (cli *socketClient) EchoAsync(msg string) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_Echo{Echo: &abci.RequestEcho{Message: msg}},
	})
}

// FlushAsync sends an async Flush request
func (cli *socketClient) FlushAsync() *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_Flush{Flush: &abci.RequestFlush{}},
	})
}

// InfoAsync sends an async Info request
func (cli *socketClient) InfoAsync(req abci.RequestInfo) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_Info{Info: &req},
	})
}

// SetOptionAsync sends an async SetOption request
func (cli *socketClient) SetOptionAsync(req abci.RequestSetOption) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_SetOption{SetOption: &req},
	})
}

// CheckTxAsync sends an async CheckTx request
func (cli *socketClient) CheckTxAsync(req abci.RequestCheckTx) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_CheckTx{CheckTx: &req},
	})
}

// QueryAsync sends an async Query request
func (cli *socketClient) QueryAsync(req abci.RequestQuery) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_Query{Query: &req},
	})
}

// CommitAsync sends an async Commit request
func (cli *socketClient) CommitAsync() *ReqRes {
	return cli.queueRequest(&abci.Request{})
}

// InitChainAsync sends an async InitChain request
func (cli *socketClient) InitChainAsync(req abci.RequestInitChain) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_InitChain{InitChain: &req},
	})
}

// ListSnapshotsAsync sends an async ListSnapshots request
func (cli *socketClient) ListSnapshotsAsync(req abci.RequestListSnapshots) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_ListSnapshots{ListSnapshots: &req},
	})
}

// OfferSnapshotAsync sends an async OfferSnapshot request
func (cli *socketClient) OfferSnapshotAsync(req abci.RequestOfferSnapshot) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_OfferSnapshot{OfferSnapshot: &req},
	})
}

// LoadSnapshotChunkAsync sends an async LoadSnapshotChunk request
func (cli *socketClient) LoadSnapshotChunkAsync(req abci.RequestLoadSnapshotChunk) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_LoadSnapshotChunk{LoadSnapshotChunk: &req},
	})
}

// ApplySnapshotChunkAsync sends an async ApplySnapshotChunk request
func (cli *socketClient) ApplySnapshotChunkAsync(req abci.RequestApplySnapshotChunk) *ReqRes {
	return cli.queueRequest(&abci.Request{
		Value: &abci.Request_ApplySnapshotChunk{ApplySnapshotChunk: &req},
	})
}

//----------------------------------------

// EchoSync sends a sync Echo request
func (cli *socketClient) EchoSync(msg string) (*abci.ResponseEcho, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_Echo{Echo: &abci.RequestEcho{Message: msg}},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_Echo); ok {
		return r.Echo, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// FlushSync sends a sync Flush request
func (cli *socketClient) FlushSync() error {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_Flush{Flush: &abci.RequestFlush{}},
	})
	cli.finishSyncCall(reqres)
	if _, ok := reqres.Response.Value.(*abci.Response_Flush); ok {
		return cli.Error()
	}
	return fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// InfoSync sends a sync Info request
func (cli *socketClient) InfoSync(req abci.RequestInfo) (*abci.ResponseInfo, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_Info{Info: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_Info); ok {
		return r.Info, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// SetOptionSync sends a sync SetOption request
func (cli *socketClient) SetOptionSync(req abci.RequestSetOption) (*abci.ResponseSetOption, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_SetOption{SetOption: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_SetOption); ok {
		return r.SetOption, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// CheckTxSync sends a sync CheckTx request
func (cli *socketClient) CheckTxSync(req abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_CheckTx{CheckTx: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_CheckTx); ok {
		return r.CheckTx, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// QuerySync sends a sync Query request
func (cli *socketClient) QuerySync(req abci.RequestQuery) (*abci.ResponseQuery, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_Query{Query: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_Query); ok {
		return r.Query, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// CommitSync sends a sync Commit request
func (cli *socketClient) CommitSync() (*abci.ResponseCommit, error) {
	reqres := cli.queueRequest(&abci.Request{})
	cli.finishSyncCall(reqres)
	// Since we can't use abci.Response_Commit, just return a stub response
	return &abci.ResponseCommit{}, cli.Error()
}

// InitChainSync sends a sync InitChain request
func (cli *socketClient) InitChainSync(req abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_InitChain{InitChain: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_InitChain); ok {
		return r.InitChain, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// ListSnapshotsSync sends a sync ListSnapshots request
func (cli *socketClient) ListSnapshotsSync(req abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_ListSnapshots{ListSnapshots: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_ListSnapshots); ok {
		return r.ListSnapshots, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// OfferSnapshotSync sends a sync OfferSnapshot request
func (cli *socketClient) OfferSnapshotSync(req abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_OfferSnapshot{OfferSnapshot: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_OfferSnapshot); ok {
		return r.OfferSnapshot, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// LoadSnapshotChunkSync sends a sync LoadSnapshotChunk request
func (cli *socketClient) LoadSnapshotChunkSync(req abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_LoadSnapshotChunk{LoadSnapshotChunk: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_LoadSnapshotChunk); ok {
		return r.LoadSnapshotChunk, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

// ApplySnapshotChunkSync sends a sync ApplySnapshotChunk request
func (cli *socketClient) ApplySnapshotChunkSync(req abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	reqres := cli.queueRequest(&abci.Request{
		Value: &abci.Request_ApplySnapshotChunk{ApplySnapshotChunk: &req},
	})
	cli.finishSyncCall(reqres)
	if r, ok := reqres.Response.Value.(*abci.Response_ApplySnapshotChunk); ok {
		return r.ApplySnapshotChunk, cli.Error()
	}
	return nil, fmt.Errorf("unexpected response type: %T", reqres.Response.Value)
}

//----------------------------------------

func (cli *socketClient) queueRequest(req *abci.Request) *ReqRes {
	reqres := NewReqRes(req)

	// TODO: set cli.err if reqQueue times out
	cli.reqQueue <- reqres

	// Maybe auto-flush, or unset auto-flush
	switch req.Value.(type) {
	case *abci.Request_Flush:
		cli.flushTimer.Unset()
	default:
		cli.flushTimer.Set()
	}

	return reqres
}

func (cli *socketClient) flushQueue() {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// mark all in-flight messages as resolved (they will get cli.Error())
	for req := cli.reqSent.Front(); req != nil; req = req.Next() {
		reqres := req.Value.(*ReqRes)
		reqres.Done()
	}

	// mark all queued messages as resolved
LOOP:
	for {
		select {
		case reqres := <-cli.reqQueue:
			reqres.Done()
		default:
			break LOOP
		}
	}
}

//----------------------------------------

func resMatchesReq(req *abci.Request, res *abci.Response) (ok bool) {
	switch req.Value.(type) {
	case *abci.Request_Echo:
		_, ok = res.Value.(*abci.Response_Echo)
	case *abci.Request_Flush:
		_, ok = res.Value.(*abci.Response_Flush)
	case *abci.Request_Info:
		_, ok = res.Value.(*abci.Response_Info)
	case *abci.Request_SetOption:
		_, ok = res.Value.(*abci.Response_SetOption)
	case *abci.Request_CheckTx:
		_, ok = res.Value.(*abci.Response_CheckTx)
	case *abci.Request_Query:
		_, ok = res.Value.(*abci.Response_Query)
	case *abci.Request_InitChain:
		_, ok = res.Value.(*abci.Response_InitChain)
	case *abci.Request_ApplySnapshotChunk:
		_, ok = res.Value.(*abci.Response_ApplySnapshotChunk)
	case *abci.Request_LoadSnapshotChunk:
		_, ok = res.Value.(*abci.Response_LoadSnapshotChunk)
	case *abci.Request_ListSnapshots:
		_, ok = res.Value.(*abci.Response_ListSnapshots)
	case *abci.Request_OfferSnapshot:
		_, ok = res.Value.(*abci.Response_OfferSnapshot)
	}
	return ok
}

func (cli *socketClient) stopForError(err error) {
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	if cli.err == nil {
		cli.err = err
	}
	cli.mtx.Unlock()

	cli.Logger.Error(fmt.Sprintf("Stopping abci.socketClient for error: %v", err.Error()))
	if err := cli.Stop(); err != nil {
		cli.Logger.Error("Error stopping abci.socketClient", "err", err)
	}
}

func (cli *socketClient) finishSyncCall(reqres *ReqRes) {
	reqres.Wait()
}
