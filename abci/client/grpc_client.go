package client

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	tmnet "github.com/fluentum-chain/fluentum/libs/net"
	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

// Callback type for ABCI responses
type Callback func(*abci.Request, *abci.Response)

var _ Client = (*grpcClient)(nil)

// A stripped copy of the remoteClient that makes
// synchronous calls using grpc
type grpcClient struct {
	service.BaseService
	mustConnect bool

	client   abci.ABCIApplicationClient
	conn     *grpc.ClientConn
	chReqRes chan *ReqRes // dispatches "async" responses to callbacks *in order*, needed by mempool

	mtx   tmsync.Mutex
	addr  string
	err   error
	resCb Callback // listens to all callbacks
}

func NewGRPCClient(addr string, mustConnect bool) Client {
	cli := &grpcClient{
		addr:        addr,
		mustConnect: mustConnect,
		// Buffering the channel is needed to make calls appear asynchronous,
		// which is required when the caller makes multiple async calls before
		// processing callbacks (e.g. due to holding locks). 64 means that a
		// caller can make up to 64 async calls before a callback must be
		// processed (otherwise it deadlocks). It also means that we can make 64
		// gRPC calls while processing a slow callback at the channel head.
		chReqRes: make(chan *ReqRes, 64),
	}
	cli.BaseService = *service.NewBaseService(nil, "grpcClient", cli)
	return cli
}

func dialerFunc(ctx context.Context, addr string) (net.Conn, error) {
	return tmnet.Connect(addr)
}

func (cli *grpcClient) OnStart() error {
	if err := cli.BaseService.OnStart(); err != nil {
		return err
	}

	// This processes asynchronous request/response messages and dispatches
	// them to callbacks.
	go func() {
		// Use a separate function to use defer for mutex unlocks (this handles panics)
		callCb := func(reqres *ReqRes) {
			cli.mtx.Lock()
			defer cli.mtx.Unlock()

			reqres.Done()

			// Notify client listener if set
			if cli.resCb != nil {
				cli.resCb(reqres.Request, reqres.Response)
			}

			// Notify reqRes listener if set
			reqres.InvokeCallback()
		}
		for reqres := range cli.chReqRes {
			if reqres != nil {
				callCb(reqres)
			} else {
				cli.Logger.Error("Received nil reqres")
			}
		}
	}()

RETRY_LOOP:
	for {
		//nolint:staticcheck // SA1019 Existing use of deprecated but supported dial option.
		conn, err := grpc.Dial(cli.addr, grpc.WithInsecure(), grpc.WithContextDialer(dialerFunc))
		if err != nil {
			if cli.mustConnect {
				return err
			}
			cli.Logger.Error(fmt.Sprintf("abci.grpcClient failed to connect to %v.  Retrying...\n", cli.addr), "err", err)
			time.Sleep(time.Second * dialRetryIntervalSeconds)
			continue RETRY_LOOP
		}

		cli.Logger.Info("Dialed server. Waiting for echo.", "addr", cli.addr)
		client := abci.NewABCIApplicationClient(conn)
		cli.conn = conn

	ENSURE_CONNECTED:
		for {
			_, err := client.Echo(context.Background(), &abci.RequestEcho{Message: "hello"}, grpc.WaitForReady(true))
			if err == nil {
				break ENSURE_CONNECTED
			}
			cli.Logger.Error("Echo failed", "err", err)
			time.Sleep(time.Second * echoRetryIntervalSeconds)
		}

		cli.client = client
		return nil
	}
}

func (cli *grpcClient) OnStop() {
	cli.BaseService.OnStop()

	if cli.conn != nil {
		cli.conn.Close()
	}
	close(cli.chReqRes)
}

func (cli *grpcClient) StopForError(err error) {
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	if cli.err == nil {
		cli.err = err
	}
	cli.mtx.Unlock()

	cli.Logger.Error(fmt.Sprintf("Stopping abci.grpcClient for error: %v", err.Error()))
	if err := cli.Stop(); err != nil {
		cli.Logger.Error("Error stopping abci.grpcClient", "err", err)
	}
}

func (cli *grpcClient) Error() error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	return cli.err
}

// Set listener for all responses
// NOTE: callback may get internally generated flush responses.
func (cli *grpcClient) SetResponseCallback(resCb Callback) {
	cli.mtx.Lock()
	cli.resCb = resCb
	cli.mtx.Unlock()
}

//----------------------------------------
// GRPC calls are synchronous, but some callbacks expect to be called asynchronously
// (eg. the mempool expects to be able to lock to remove bad txs from cache).
// To accommodate, we finish each call in its own go-routine,
// which is expensive, but easy - if you want something better, use the socket protocol!
// maybe one day, if people really want it, we use grpc streams,
// but hopefully not :D

func (cli *grpcClient) EchoAsync(msg string) *ReqRes {
	// req := abci.ToRequestEcho(msg)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_Echo{
			Echo: &abci.RequestEcho{Message: msg},
		},
	}
	res, err := cli.client.Echo(context.Background(), req.GetEcho(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_Echo{Echo: res}})
}

func (cli *grpcClient) FlushAsync() *ReqRes {
	// req := abci.ToRequestFlush()
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_Flush{
			Flush: &abci.RequestFlush{},
		},
	}
	res, err := cli.client.Flush(context.Background(), req.GetFlush(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_Flush{Flush: res}})
}

func (cli *grpcClient) InfoAsync(params abci.RequestInfo) *ReqRes {
	// req := abci.ToRequestInfo(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_Info{
			Info: &params,
		},
	}
	res, err := cli.client.Info(context.Background(), req.GetInfo(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_Info{Info: res}})
}

func (cli *grpcClient) SetOptionAsync(params abci.RequestSetOption) *ReqRes {
	// req := abci.ToRequestSetOption(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_SetOption{
			SetOption: &params,
		},
	}
	res, err := cli.client.SetOption(context.Background(), req.GetSetOption(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_SetOption{SetOption: res}})
}

func (cli *grpcClient) CheckTxAsync(params abci.RequestCheckTx) *ReqRes {
	// req := abci.ToRequestCheckTx(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_CheckTx{
			CheckTx: &params,
		},
	}
	res, err := cli.client.CheckTx(context.Background(), req.GetCheckTx(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_CheckTx{CheckTx: res}})
}

func (cli *grpcClient) QueryAsync(params abci.RequestQuery) *ReqRes {
	// req := abci.ToRequestQuery(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_Query{
			Query: &params,
		},
	}
	res, err := cli.client.Query(context.Background(), req.GetQuery(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_Query{Query: res}})
}

func (cli *grpcClient) CommitAsync() *ReqRes {
	// TODO: Fix this when Request_Commit is added to protobuf
	// req := &abci.Request{
	// 	Value: &abci.Request_Commit{
	// 		Commit: &abci.RequestCommit{},
	// 	},
	// }
	// res, err := cli.client.Commit(context.Background(), req.GetCommit(), grpc.WaitForReady(true))
	// if err != nil {
	// 	cli.StopForError(err)
	// }
	// return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_Commit{Commit: res}})
	return nil
}

func (cli *grpcClient) InitChainAsync(params abci.RequestInitChain) *ReqRes {
	// req := abci.ToRequestInitChain(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_InitChain{
			InitChain: &params,
		},
	}
	res, err := cli.client.InitChain(context.Background(), req.GetInitChain(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_InitChain{InitChain: res}})
}

func (cli *grpcClient) ListSnapshotsAsync(params abci.RequestListSnapshots) *ReqRes {
	// req := abci.ToRequestListSnapshots(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_ListSnapshots{
			ListSnapshots: &params,
		},
	}
	res, err := cli.client.ListSnapshots(context.Background(), req.GetListSnapshots(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_ListSnapshots{ListSnapshots: res}})
}

func (cli *grpcClient) OfferSnapshotAsync(params abci.RequestOfferSnapshot) *ReqRes {
	// req := abci.ToRequestOfferSnapshot(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_OfferSnapshot{
			OfferSnapshot: &params,
		},
	}
	res, err := cli.client.OfferSnapshot(context.Background(), req.GetOfferSnapshot(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_OfferSnapshot{OfferSnapshot: res}})
}

func (cli *grpcClient) LoadSnapshotChunkAsync(params abci.RequestLoadSnapshotChunk) *ReqRes {
	// req := abci.ToRequestLoadSnapshotChunk(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_LoadSnapshotChunk{
			LoadSnapshotChunk: &params,
		},
	}
	res, err := cli.client.LoadSnapshotChunk(context.Background(), req.GetLoadSnapshotChunk(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_LoadSnapshotChunk{LoadSnapshotChunk: res}})
}

func (cli *grpcClient) ApplySnapshotChunkAsync(params abci.RequestApplySnapshotChunk) *ReqRes {
	// req := abci.ToRequestApplySnapshotChunk(params)
	// TODO: Fix this when protobuf is properly generated
	req := &abci.Request{
		Value: &abci.Request_ApplySnapshotChunk{
			ApplySnapshotChunk: &params,
		},
	}
	res, err := cli.client.ApplySnapshotChunk(context.Background(), req.GetApplySnapshotChunk(), grpc.WaitForReady(true))
	if err != nil {
		cli.StopForError(err)
	}
	return cli.finishAsyncCall(req, &abci.Response{Value: &abci.Response_ApplySnapshotChunk{ApplySnapshotChunk: res}})
}

// finishAsyncCall creates a ReqRes for an async call, and immediately populates it
// with the response. We don't complete it until it's been ordered via the channel.
func (cli *grpcClient) finishAsyncCall(req *abci.Request, res *abci.Response) *ReqRes {
	reqres := NewReqRes(req)
	reqres.Response = res
	cli.chReqRes <- reqres // use channel for async responses, since they must be ordered
	return reqres
}

//----------------------------------------

func (cli *grpcClient) FlushSync() error {
	reqres := cli.FlushAsync()
	cli.finishSyncCall(reqres).GetFlush()
	return cli.Error()
}

func (cli *grpcClient) EchoSync(msg string) (*abci.ResponseEcho, error) {
	reqres := cli.EchoAsync(msg)
	// StopForError should already have been called if error is set
	return cli.finishSyncCall(reqres).GetEcho(), cli.Error()
}

func (cli *grpcClient) InfoSync(req abci.RequestInfo) (*abci.ResponseInfo, error) {
	reqres := cli.InfoAsync(req)
	return cli.finishSyncCall(reqres).GetInfo(), cli.Error()
}

func (cli *grpcClient) SetOptionSync(req abci.RequestSetOption) (*abci.ResponseSetOption, error) {
	reqres := cli.SetOptionAsync(req)
	return reqres.Response.GetSetOption(), cli.Error()
}

func (cli *grpcClient) CheckTxSync(params abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	reqres := cli.CheckTxAsync(params)
	return cli.finishSyncCall(reqres).GetCheckTx(), cli.Error()
}

func (cli *grpcClient) QuerySync(req abci.RequestQuery) (*abci.ResponseQuery, error) {
	reqres := cli.QueryAsync(req)
	return cli.finishSyncCall(reqres).GetQuery(), cli.Error()
}

func (cli *grpcClient) CommitSync() (*abci.ResponseCommit, error) {
	// TODO: Fix this when Request_Commit is added to protobuf
	// reqres := cli.CommitAsync()
	// return cli.finishSyncCall(reqres).GetCommit(), cli.Error()
	return nil, nil
}

func (cli *grpcClient) InitChainSync(params abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	reqres := cli.InitChainAsync(params)
	return cli.finishSyncCall(reqres).GetInitChain(), cli.Error()
}

func (cli *grpcClient) ListSnapshotsSync(params abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	reqres := cli.ListSnapshotsAsync(params)
	return cli.finishSyncCall(reqres).GetListSnapshots(), cli.Error()
}

func (cli *grpcClient) OfferSnapshotSync(params abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	reqres := cli.OfferSnapshotAsync(params)
	return cli.finishSyncCall(reqres).GetOfferSnapshot(), cli.Error()
}

func (cli *grpcClient) LoadSnapshotChunkSync(
	params abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	reqres := cli.LoadSnapshotChunkAsync(params)
	return cli.finishSyncCall(reqres).GetLoadSnapshotChunk(), cli.Error()
}

func (cli *grpcClient) ApplySnapshotChunkSync(
	params abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	reqres := cli.ApplySnapshotChunkAsync(params)
	return cli.finishSyncCall(reqres).GetApplySnapshotChunk(), cli.Error()
}

// finishSyncCall waits for an async call to complete. It is necessary to call all
// sync calls asynchronously as well, to maintain call and response ordering via
// the channel, and this method will wait until the async call completes.
func (cli *grpcClient) finishSyncCall(reqres *ReqRes) *abci.Response {
	// It's possible that the callback is called twice, since the callback can
	// be called immediately on SetCallback() in addition to after it has been
	// set. This is because completing the ReqRes happens in a separate critical
	// section from the one where the callback is called: there is a race where
	// SetCallback() is called between completing the ReqRes and dispatching the
	// callback.
	//
	// We also buffer the channel with 1 response, since SetCallback() will be
	// called synchronously if the reqres is already completed, in which case
	// it will block on sending to the channel since it hasn't gotten around to
	// receiving from it yet.
	//
	// ReqRes should really handle callback dispatch internally, to guarantee
	// that it's only called once and avoid the above race conditions.
	var once sync.Once
	ch := make(chan *abci.Response, 1)
	reqres.SetCallback(func(res *abci.Response) {
		once.Do(func() {
			ch <- res
		})
	})
	return <-ch
}
