package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fluentum-chain/fluentum/abci"
	"github.com/fluentum-chain/fluentum/abci/types"
	cmtabci "github.com/cometbft/cometbft/abci/types"
)

func main() {
	fmt.Println("=== ABCI Application Interface Verification ===")
	
	// Run all verifications
	if err := abci.RunAllVerifications(); err != nil {
		log.Fatalf("Verification failed: %v", err)
	}
	
	fmt.Println("\n=== ABCI Application Demo ===")
	
	// Create a new application
	app := abci.NewMyApp("demo-chain")
	
	// Initialize the chain
	initReq := &types.RequestInitChain{
		ChainId:       "demo-chain",
		InitialHeight: 1,
		AppStateBytes: []byte("initial state"),
	}
	
	initRes, err := app.InitChain(context.Background(), initReq)
	if err != nil {
		log.Fatalf("InitChain failed: %v", err)
	}
	fmt.Printf("✓ Chain initialized: %s\n", initReq.ChainId)
	
	// Get application info
	infoRes, err := app.Info(context.Background(), &types.RequestInfo{})
	if err != nil {
		log.Fatalf("Info failed: %v", err)
	}
	fmt.Printf("✓ App info: %s\n", infoRes.Data)
	
	// Process some transactions
	txs := [][]byte{
		[]byte("SET user1=alice"),
		[]byte("SET user2=bob"),
		[]byte("SET balance1=100"),
		[]byte("SET balance2=200"),
	}
	
	// Check transactions
	for i, tx := range txs {
		checkRes, err := app.CheckTx(context.Background(), &types.RequestCheckTx{
			Tx:   tx,
			Type: types.CheckTxType_New,
		})
		if err != nil {
			log.Fatalf("CheckTx failed for tx %d: %v", i, err)
		}
		fmt.Printf("✓ Tx %d check: %s (gas: %d)\n", i+1, checkRes.Log, checkRes.GasWanted)
	}
	
	// Finalize block
	finalizeRes, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 1,
		Txs:    txs,
	})
	if err != nil {
		log.Fatalf("FinalizeBlock failed: %v", err)
	}
	fmt.Printf("✓ Block finalized: %d transactions processed\n", len(finalizeRes.TxResults))
	
	// Commit the block
	commitRes, err := app.Commit(context.Background(), &types.RequestCommit{})
	if err != nil {
		log.Fatalf("Commit failed: %v", err)
	}
	fmt.Printf("✓ Block committed: app hash = %x\n", commitRes.Data)
	
	// Query the state
	queryRes, err := app.Query(context.Background(), &types.RequestQuery{
		Path: "state",
		Data: []byte("user1"),
	})
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("✓ Query result: user1 = %s\n", string(queryRes.Value))
	
	// Process another block with more transactions
	txs2 := [][]byte{
		[]byte("SET user3=charlie"),
		[]byte("GET user1"),
		[]byte("GET balance1"),
	}
	
	finalizeRes2, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 2,
		Txs:    txs2,
	})
	if err != nil {
		log.Fatalf("FinalizeBlock failed: %v", err)
	}
	fmt.Printf("✓ Block 2 finalized: %d transactions processed\n", len(finalizeRes2.TxResults))
	
	// Show transaction results
	for i, result := range finalizeRes2.TxResults {
		fmt.Printf("  Tx %d: %s (gas: %d)\n", i+1, result.Log, result.GasUsed)
	}
	
	// Commit the second block
	commitRes2, err := app.Commit(context.Background(), &types.RequestCommit{})
	if err != nil {
		log.Fatalf("Commit failed: %v", err)
	}
	fmt.Printf("✓ Block 2 committed: app hash = %x\n", commitRes2.Data)
	
	// Test error handling
	fmt.Println("\n=== Error Handling Demo ===")
	
	// Test invalid transaction
	invalidTx := []byte("INVALID")
	checkInvalidRes, err := app.CheckTx(context.Background(), &types.RequestCheckTx{
		Tx:   invalidTx,
		Type: types.CheckTxType_New,
	})
	if err != nil {
		log.Fatalf("CheckTx failed: %v", err)
	}
	fmt.Printf("✓ Invalid tx check: %s (code: %d)\n", checkInvalidRes.Log, checkInvalidRes.Code)
	
	// Test query for non-existent key
	queryInvalidRes, err := app.Query(context.Background(), &types.RequestQuery{
		Path: "state",
		Data: []byte("nonexistent"),
	})
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("✓ Query non-existent: %s (code: %d)\n", queryInvalidRes.Log, queryInvalidRes.Code)
	
	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The ABCI application interface is working correctly!")
} 