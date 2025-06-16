package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/kellyadamtan/tendermint/types"
	"github.com/gorilla/websocket"
)

type BenchmarkConfig struct {
	Endpoints []string
	Duration  int
	Size      int
	Rate      int
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	// Parse flags
	config := &BenchmarkConfig{}
	flag.StringVar(&config.Endpoints[0], "endpoints", "ws://localhost:26657/websocket", "WebSocket endpoints")
	flag.IntVar(&config.Duration, "duration", 60, "Benchmark duration in seconds")
	flag.IntVar(&config.Size, "size", 250, "Transaction size in bytes")
	flag.IntVar(&config.Rate, "rate", 10000, "Transactions per second")
	flag.Parse()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Duration)*time.Second)
	defer cancel()

	// Initialize metrics
	metrics := &BenchmarkMetrics{
		StartTime: time.Now(),
		Mutex:     &sync.Mutex{},
	}

	// Connect to nodes
	conns := make([]*websocket.Conn, len(config.Endpoints))
	for i, endpoint := range config.Endpoints {
		conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", endpoint, err)
		}
		conns[i] = conn
		defer conn.Close()
	}

	// Start transaction generator
	txChan := make(chan *types.Tx, config.Rate*2)
	go generateTransactions(ctx, txChan, config.Size, config.Rate)

	// Start workers
	var wg sync.WaitGroup
	for _, conn := range conns {
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			worker(ctx, c, txChan, metrics)
		}(conn)
	}

	// Wait for benchmark to complete
	<-ctx.Done()
	wg.Wait()

	// Print results
	printResults(metrics, config)
	return nil
}

type BenchmarkMetrics struct {
	StartTime    time.Time
	Mutex        *sync.Mutex
	TotalTxs     int64
	ConfirmedTxs int64
	Latencies    []time.Duration
}

func generateTransactions(ctx context.Context, txChan chan<- *types.Tx, size, rate int) {
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tx := &types.Tx{
				Type:    types.TxTypeTransfer,
				From:    fmt.Sprintf("0x%x", rand.Int63()),
				To:      fmt.Sprintf("0x%x", rand.Int63()),
				Amount:  rand.Int63(),
				Nonce:   rand.Int63(),
				Gas:     21000,
				GasPrice: 20000000000,
				Data:    make([]byte, size),
			}
			rand.Read(tx.Data)
			txChan <- tx
		}
	}
}

func worker(ctx context.Context, conn *websocket.Conn, txChan <-chan *types.Tx, metrics *BenchmarkMetrics) {
	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-txChan:
			start := time.Now()

			// Send transaction
			err := conn.WriteJSON(map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "broadcast_tx_async",
				"params":  []interface{}{tx},
				"id":      rand.Int63(),
			})
			if err != nil {
				log.Printf("Failed to send transaction: %v", err)
				continue
			}

			// Wait for confirmation
			var response map[string]interface{}
			err = conn.ReadJSON(&response)
			if err != nil {
				log.Printf("Failed to read response: %v", err)
				continue
			}

			// Update metrics
			metrics.Mutex.Lock()
			metrics.TotalTxs++
			if response["result"] != nil {
				metrics.ConfirmedTxs++
				metrics.Latencies = append(metrics.Latencies, time.Since(start))
			}
			metrics.Mutex.Unlock()
		}
	}
}

func printResults(metrics *BenchmarkMetrics, config *BenchmarkConfig) {
	duration := time.Since(metrics.StartTime).Seconds()
	tps := float64(metrics.ConfirmedTxs) / duration

	// Calculate latency statistics
	var totalLatency time.Duration
	for _, lat := range metrics.Latencies {
		totalLatency += lat
	}
	avgLatency := totalLatency / time.Duration(len(metrics.Latencies))

	fmt.Printf("\nBenchmark Results:\n")
	fmt.Printf("Duration: %d seconds\n", config.Duration)
	fmt.Printf("Transaction Size: %d bytes\n", config.Size)
	fmt.Printf("Target Rate: %d TPS\n", config.Rate)
	fmt.Printf("Total Transactions: %d\n", metrics.TotalTxs)
	fmt.Printf("Confirmed Transactions: %d\n", metrics.ConfirmedTxs)
	fmt.Printf("TPS: %.0f\n", tps)
	fmt.Printf("Average Latency: %v\n", avgLatency)
	fmt.Printf("Success Rate: %.2f%%\n", float64(metrics.ConfirmedTxs)/float64(metrics.TotalTxs)*100)
} 