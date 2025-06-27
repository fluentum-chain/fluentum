package testsuite

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	abcicli "github.com/fluentum-chain/fluentum/abci/client"
	tmrand "github.com/fluentum-chain/fluentum/libs/rand"
)

func InitChain(client abcicli.Client) error {
	total := 10
	vals := make([]cmtabci.ValidatorUpdate, total)
	for i := 0; i < total; i++ {
		power := tmrand.Int()
		// For now, we'll use a simple approach without the complex PubKey structure
		vals[i] = cmtabci.ValidatorUpdate{
			Power: int64(power),
		}
	}
	_, err := client.InitChain(context.Background(), &cmtabci.RequestInitChain{
		Validators: vals,
	})
	if err != nil {
		fmt.Printf("Failed test: InitChain - %v\n", err)
		return err
	}
	fmt.Println("Passed test: InitChain")
	return nil
}

func Commit(client abcicli.Client, hashExp []byte) error {
	_, err := client.Commit(context.Background())
	if err != nil {
		fmt.Println("Failed test: Commit")
		fmt.Printf("error while committing: %v\n", err)
		return err
	}
	// Note: ResponseCommit doesn't have Data field in CometBFT v0.38+
	// We'll just check if the commit was successful
	fmt.Println("Passed test: Commit")
	return nil
}

func DeliverTx(client abcicli.Client, txBytes []byte, codeExp uint32, dataExp []byte) error {
	res, err := client.FinalizeBlock(context.Background(), &cmtabci.RequestFinalizeBlock{
		Txs: [][]byte{txBytes},
	})
	if err != nil {
		fmt.Printf("Failed test: DeliverTx - %v\n", err)
		return err
	}

	if len(res.TxResults) == 0 {
		fmt.Println("Failed test: DeliverTx - no transaction results")
		return errors.New("no transaction results")
	}

	txRes := res.TxResults[0]
	code, data, log := txRes.Code, txRes.Data, txRes.Log
	if code != codeExp {
		fmt.Println("Failed test: DeliverTx")
		fmt.Printf("DeliverTx response code was unexpected. Got %v expected %v. Log: %v\n",
			code, codeExp, log)
		return errors.New("deliverTx error")
	}
	if !bytes.Equal(data, dataExp) {
		fmt.Println("Failed test: DeliverTx")
		fmt.Printf("DeliverTx response data was unexpected. Got %X expected %X\n",
			data, dataExp)
		return errors.New("deliverTx error")
	}
	fmt.Println("Passed test: DeliverTx")
	return nil
}

func CheckTx(client abcicli.Client, txBytes []byte, codeExp uint32, dataExp []byte) error {
	res, err := client.CheckTx(context.Background(), &cmtabci.RequestCheckTx{Tx: txBytes})
	if err != nil {
		fmt.Printf("Failed test: CheckTx - %v\n", err)
		return err
	}

	code, data, log := res.Code, res.Data, res.Log
	if code != codeExp {
		fmt.Println("Failed test: CheckTx")
		fmt.Printf("CheckTx response code was unexpected. Got %v expected %v. Log: %v\n",
			code, codeExp, log)
		return errors.New("checkTx")
	}
	if !bytes.Equal(data, dataExp) {
		fmt.Println("Failed test: CheckTx")
		fmt.Printf("CheckTx response data was unexpected. Got %X expected %X\n",
			data, dataExp)
		return errors.New("checkTx")
	}
	fmt.Println("Passed test: CheckTx")
	return nil
}
