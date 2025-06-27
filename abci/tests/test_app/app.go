package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
)

func startClient(abciType string) abcicli.Client {
	// Start client
	client, err := abcicli.NewClient("tcp://127.0.0.1:26658", abciType, true)
	if err != nil {
		panic(err.Error())
	}
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	client.SetLogger(logger.With("module", "abcicli"))
	if err := client.Start(); err != nil {
		panicf("connecting to abci_app: %v", err.Error())
	}

	return client
}

func commit(client abcicli.Client, hashExp []byte) {
	res, err := client.Commit(context.Background(), &types.RequestCommit{})
	if err != nil {
		panicf("client error: %v", err)
	}
	fmt.Printf("Commit RetainHeight: %d\n", res.RetainHeight)
	// No Data field available in this CometBFT version
}

func deliverTx(client abcicli.Client, txBytes []byte, codeExp uint32, dataExp []byte) {
	req := &types.RequestFinalizeBlock{
		Txs: [][]byte{txBytes},
	}
	res, err := client.FinalizeBlock(context.Background(), req)
	if err != nil {
		panicf("client error: %v", err)
	}
	if len(res.TxResults) == 0 {
		panicf("No transaction results returned")
	}
	txResult := res.TxResults[0]
	if txResult.Code != codeExp {
		panicf("DeliverTx response code was unexpected. Got %v expected %v. Log: %v", txResult.Code, codeExp, txResult.Log)
	}
	if !bytes.Equal(txResult.Data, dataExp) {
		panicf("DeliverTx response data was unexpected. Got %X expected %X", txResult.Data, dataExp)
	}
}

/*func checkTx(client abcicli.Client, txBytes []byte, codeExp uint32, dataExp []byte) {
	res, err := client.CheckTxSync(txBytes)
	if err != nil {
		panicf("client error: %v", err)
	}
	if res.IsErr() {
		panicf("checking tx %X: %v\nlog: %v", txBytes, res.Log)
	}
	if res.Code != codeExp {
		panicf("CheckTx response code was unexpected. Got %v expected %v. Log: %v",
			res.Code, codeExp, res.Log)
	}
	if !bytes.Equal(res.Data, dataExp) {
		panicf("CheckTx response data was unexpected. Got %X expected %X",
			res.Data, dataExp)
	}
}*/

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
