package coregrpc

import (
	"context"

	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
	coregrpc "github.com/fluentum-chain/fluentum/proto/tendermint/rpc/grpc"
	core "github.com/fluentum-chain/fluentum/rpc/core"
	rpctypes "github.com/fluentum-chain/fluentum/rpc/jsonrpc/types"
)

type broadcastAPI struct {
	coregrpc.UnimplementedBroadcastAPIServer
}

func (bapi *broadcastAPI) Ping(ctx context.Context, req *coregrpc.RequestPing) (*coregrpc.ResponsePing, error) {
	// kvstore so we can check if the server is up
	return &coregrpc.ResponsePing{}, nil
}

func (bapi *broadcastAPI) BroadcastTx(ctx context.Context, req *coregrpc.RequestBroadcastTx) (*coregrpc.ResponseBroadcastTx, error) {
	// NOTE: there's no way to get client's remote address
	// see https://stackoverflow.com/questions/33684570/session-and-remote-ip-address-in-grpc-go
	res, err := core.BroadcastTxCommit(&rpctypes.Context{}, req.Tx)
	if err != nil {
		return nil, err
	}

	return &coregrpc.ResponseBroadcastTx{
		CheckTx: &abci.ResponseCheckTx{
			Code: res.CheckTx.Code,
			Data: res.CheckTx.Data,
			Log:  res.CheckTx.Log,
		},
	}, nil
}
