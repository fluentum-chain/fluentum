package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"context"

	tmjson "github.com/fluentum-chain/fluentum/libs/json"
	proto_grpc "github.com/fluentum-chain/fluentum/proto/tendermint/rpc/grpc"
	coregrpc "github.com/fluentum-chain/fluentum/rpc/grpc"
)

var grpcAddr = "tcp://localhost:36656"

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Must enter a transaction to send (hex)")
		os.Exit(1)
	}
	tx := args[1]
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		fmt.Println("Invalid hex", err)
		os.Exit(1)
	}

	clientGRPC := coregrpc.StartGRPCClient(grpcAddr)
	res, err := clientGRPC.BroadcastTx(context.Background(), &proto_grpc.RequestBroadcastTx{Tx: txBytes})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bz, err := tmjson.Marshal(res)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(bz))
}
