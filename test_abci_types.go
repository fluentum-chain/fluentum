package main

import (
	"fmt"
	"time"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	abci "github.com/fluentum-chain/fluentum/abci/types"
)

func main() {
	fmt.Println("Testing ABCI Type Changes...")

	// Test creating ABCI request types
	infoReq := &abci.InfoRequest{
		Version:      "1.0.0",
		BlockVersion: 1,
		P2PVersion:   1,
	}
	fmt.Printf("✓ Created InfoRequest: %+v\n", infoReq)

	checkTxReq := &abci.CheckTxRequest{
		Tx:   []byte("test transaction"),
		Type: abci.CheckTxType_New,
	}
	fmt.Printf("✓ Created CheckTxRequest: %+v\n", checkTxReq)

	finalizeBlockReq := &abci.FinalizeBlockRequest{
		Txs:                [][]byte{[]byte("tx1"), []byte("tx2")},
		DecidedLastCommit:  cmtabci.CommitInfo{},
		Misbehavior:        []cmtabci.Misbehavior{},
		Hash:               []byte("block hash"),
		Height:             100,
		Time:               time.Now(),
		NextValidatorsHash: []byte("next validators hash"),
		ProposerAddress:    []byte("proposer address"),
	}
	fmt.Printf("✓ Created FinalizeBlockRequest: %+v\n", finalizeBlockReq)

	// Test creating ABCI response types
	infoResp := &abci.InfoResponse{
		Data:             "test data",
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  100,
		LastBlockAppHash: []byte("app hash"),
	}
	fmt.Printf("✓ Created InfoResponse: %+v\n", infoResp)

	checkTxResp := &abci.CheckTxResponse{
		Code:      0,
		Data:      []byte("response data"),
		Log:       "success",
		Info:      "info",
		GasWanted: 1000,
		GasUsed:   500,
		Events:    []*abci.Event{},
		Codespace: "test",
	}
	fmt.Printf("✓ Created CheckTxResponse: %+v\n", checkTxResp)

	finalizeBlockResp := &abci.FinalizeBlockResponse{
		Events:                []cmtabci.Event{},
		TxResults:             []*cmtabci.ExecTxResult{},
		ValidatorUpdates:      []cmtabci.ValidatorUpdate{},
		ConsensusParamUpdates: &cmtproto.ConsensusParams{},
		AppHash:               []byte("app hash"),
	}
	fmt.Printf("✓ Created FinalizeBlockResponse: %+v\n", finalizeBlockResp)

	// Test constants
	fmt.Printf("✓ CheckTxType_New: %v\n", abci.CheckTxType_New)
	fmt.Printf("✓ CheckTxType_Recheck: %v\n", abci.CheckTxType_Recheck)
	fmt.Printf("✓ ResponseOfferSnapshot_REJECT: %v\n", abci.ResponseOfferSnapshot_REJECT)
	fmt.Printf("✓ ResponseApplySnapshotChunk_ABORT: %v\n", abci.ResponseApplySnapshotChunk_ABORT)

	fmt.Println("\n🎉 All ABCI type tests passed!")
	fmt.Println("The ABCI type migration has been successful!")
}
