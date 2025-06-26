package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
	"github.com/fluentum-chain/fluentum/fluentum/core/types"
	"github.com/fluentum-chain/fluentum/fluentum/core/validator"
)

func main() {
	fmt.Println("🚀 Fluentum AI-Validation Core Demo")
	fmt.Println("=====================================")

	// Initialize AI-Validation Core
	if err := runAIDemo(); err != nil {
		log.Fatalf("AI demo failed: %v", err)
	}

	fmt.Println("\n✅ AI-Validation Core demo completed successfully!")
}

func runAIDemo() error {
	// Step 1: Initialize Plugin Manager
	fmt.Println("\n📦 Step 1: Initializing Plugin Manager")
	pm := plugin.Instance()
	
	config := plugin.DefaultPluginManagerConfig()
	config.PluginDirectory = "./plugins"
	config.AutoLoad = true
	
	if err := pm.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}
	fmt.Println("✅ Plugin manager initialized")

	// Step 2: Create AI Validator Configuration
	fmt.Println("\n🤖 Step 2: Configuring AI Validator")
	aiConfig := &validator.AIValidatorConfig{
		EnableAIPrediction:   true,
		EnableQuantumSigning: true,
		BatchSize:            50,
		MaxWaitTime:          5 * time.Second,
		ConfidenceThreshold:  0.7,
		GasSavingsThreshold:  0.3,
		PluginPath:           "./plugins/qmoe_validator.so",
		QuantumPluginPath:    "./plugins/quantum_signer.so",
		ModelConfig: map[string]interface{}{
			"num_experts":                8,
			"input_size":                 256,
			"hidden_size":                512,
			"output_size":                128,
			"top_k":                      2,
			"quantization_bits":          4,
			"quantization_update_interval": 60.0,
			"weights_path":               "./models/qmoe_fluentum.bin",
			"confidence_threshold":       0.7,
			"gas_savings_threshold":      0.3,
			"max_batch_size":             100,
			"min_batch_size":             5,
			"enable_sparse_activation":    true,
			"enable_dynamic_quantization": true,
		},
	}

	// Step 3: Create AI Validator
	fmt.Println("\n🔧 Step 3: Creating AI Validator")
	aiValidator, err := validator.NewAIValidator(aiConfig)
	if err != nil {
		return fmt.Errorf("failed to create AI validator: %w", err)
	}
	fmt.Println("✅ AI validator created")

	// Step 4: Generate Sample Transactions
	fmt.Println("\n📝 Step 4: Generating Sample Transactions")
	transactions := generateSampleTransactions(100)
	fmt.Printf("✅ Generated %d sample transactions\n", len(transactions))

	// Step 5: Demonstrate AI Prediction
	fmt.Println("\n🧠 Step 5: AI Batch Prediction")
	if err := demonstrateAIPrediction(aiValidator, transactions); err != nil {
		return fmt.Errorf("AI prediction demo failed: %w", err)
	}

	// Step 6: Demonstrate Batch Processing
	fmt.Println("\n⚡ Step 6: Batch Processing")
	if err := demonstrateBatchProcessing(aiValidator, transactions); err != nil {
		return fmt.Errorf("batch processing demo failed: %w", err)
	}

	// Step 7: Demonstrate Quantum Signing
	fmt.Println("\n🔐 Step 7: Quantum Signing")
	if err := demonstrateQuantumSigning(aiValidator); err != nil {
		return fmt.Errorf("quantum signing demo failed: %w", err)
	}

	// Step 8: Performance Metrics
	fmt.Println("\n📊 Step 8: Performance Metrics")
	if err := demonstrateMetrics(aiValidator); err != nil {
		return fmt.Errorf("metrics demo failed: %w", err)
	}

	return nil
}

func generateSampleTransactions(count int) []*types.Transaction {
	transactions := make([]*types.Transaction, count)
	
	for i := 0; i < count; i++ {
		txType := "transfer"
		gas := uint64(21000)
		value := uint64(1000000000000000000) // 1 ETH
		
		// Vary transaction types
		if i%3 == 0 {
			txType = "contract"
			gas = uint64(50000 + (i*100))
			value = uint64(500000000000000000) // 0.5 ETH
		} else if i%5 == 0 {
			txType = "delegate"
			gas = uint64(30000 + (i*50))
			value = uint64(2000000000000000000) // 2 ETH
		}
		
		transactions[i] = &types.Transaction{
			Hash:      fmt.Sprintf("tx_%d", i),
			From:      fmt.Sprintf("0x%040d", i),
			To:        fmt.Sprintf("0x%040d", i+1),
			Value:     value,
			Gas:       gas,
			GasPrice:  uint64(20000000000 + (i*1000000000)), // 20-30 Gwei
			Nonce:     uint64(i),
			Type:      txType,
			Data:      []byte(fmt.Sprintf("data_%d", i)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
	}
	
	return transactions
}

func demonstrateAIPrediction(aiValidator *validator.AIValidator, transactions []*types.Transaction) error {
	// Take a subset for prediction
	predictionTxs := transactions[:20]
	
	// Get AI prediction
	prediction, err := aiValidator.PredictOptimalBatch(predictionTxs)
	if err != nil {
		return fmt.Errorf("failed to get AI prediction: %w", err)
	}
	
	fmt.Printf("   📈 Original transactions: %d\n", len(predictionTxs))
	fmt.Printf("   🎯 Optimized batch: %d transactions\n", len(prediction))
	
	// Calculate gas savings
	originalGas := uint64(0)
	for _, tx := range predictionTxs {
		originalGas += tx.Gas
	}
	
	optimizedGas := uint64(0)
	for _, tx := range prediction {
		optimizedGas += tx.Gas
	}
	
	savings := float64(originalGas-optimizedGas) / float64(originalGas) * 100.0
	fmt.Printf("   💰 Gas savings: %.2f%%\n", savings)
	fmt.Printf("   🔥 Original gas: %d\n", originalGas)
	fmt.Printf("   ⚡ Optimized gas: %d\n", optimizedGas)
	
	// Estimate gas savings
	estimatedSavings, err := aiValidator.EstimateGasSavings(predictionTxs)
	if err != nil {
		return fmt.Errorf("failed to estimate gas savings: %w", err)
	}
	
	fmt.Printf("   🧠 AI estimated savings: %.2f%%\n", estimatedSavings*100)
	
	return nil
}

func demonstrateBatchProcessing(aiValidator *validator.AIValidator, transactions []*types.Transaction) error {
	// Add transactions to batch queue
	fmt.Println("   📥 Adding transactions to batch queue...")
	
	for i, tx := range transactions {
		if err := aiValidator.AddTransaction(tx); err != nil {
			return fmt.Errorf("failed to add transaction %d: %w", i, err)
		}
		
		// Show progress every 10 transactions
		if (i+1)%10 == 0 {
			fmt.Printf("   📊 Added %d/%d transactions\n", i+1, len(transactions))
		}
	}
	
	// Get queue size
	queueSize := aiValidator.GetBatchQueueSize()
	fmt.Printf("   📋 Batch queue size: %d\n", queueSize)
	
	// Process a sample batch
	fmt.Println("   ⚙️  Processing sample batch...")
	sampleBatch := transactions[:10]
	
	block := &types.Block{
		Transactions: sampleBatch,
		Timestamp:    time.Now(),
		Validator:    "ai_validator_demo",
	}
	
	if err := aiValidator.ProcessBlock(block); err != nil {
		return fmt.Errorf("failed to process block: %w", err)
	}
	
	fmt.Printf("   ✅ Block processed successfully\n")
	fmt.Printf("   🎯 Block confidence: %.2f\n", block.Confidence)
	fmt.Printf("   💰 Block gas savings: %.2f%%\n", block.GasSavings*100)
	
	return nil
}

func demonstrateQuantumSigning(aiValidator *validator.AIValidator) error {
	// Get signer info
	pm := plugin.Instance()
	signer, err := pm.GetSigner()
	if err != nil {
		fmt.Printf("   ⚠️  No quantum signer available: %v\n", err)
		return nil
	}
	
	// Get signer information
	info := signer.GetSignerInfo()
	fmt.Printf("   🔐 Signer algorithm: %s\n", info["algorithm"])
	fmt.Printf("   🛡️  Quantum resistant: %s\n", info["quantum_resistant"])
	fmt.Printf("   🔑 Key size: %s\n", info["key_size"])
	
	// Get supported algorithms
	algorithms := signer.GetSupportedAlgorithms()
	fmt.Printf("   📋 Supported algorithms: %v\n", algorithms)
	
	// Test signing
	testData := []byte("Fluentum AI-Validation Core Test")
	signature, err := signer.Sign(testData)
	if err != nil {
		return fmt.Errorf("failed to sign test data: %w", err)
	}
	
	fmt.Printf("   ✍️  Signature length: %d bytes\n", len(signature))
	
	// Get public key for verification
	publicKey, err := signer.GetPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	
	// Verify signature
	valid, err := signer.Verify(testData, signature, publicKey)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}
	
	if valid {
		fmt.Println("   ✅ Signature verification successful")
	} else {
		fmt.Println("   ❌ Signature verification failed")
	}
	
	return nil
}

func demonstrateMetrics(aiValidator *validator.AIValidator) error {
	// Get validator metrics
	metrics := aiValidator.GetMetrics()
	fmt.Printf("   📊 Blocks processed: %d\n", metrics.BlocksProcessed)
	fmt.Printf("   📝 Transactions processed: %d\n", metrics.TransactionsProcessed)
	fmt.Printf("   ⏱️  Average block time: %v\n", metrics.AvgBlockTime)
	fmt.Printf("   💰 Average gas savings: %.2f%%\n", metrics.AvgGasSavings*100)
	fmt.Printf("   🧠 Prediction accuracy: %.2f%%\n", metrics.PredictionAccuracy*100)
	
	// Get AI-specific metrics
	aiMetrics := aiValidator.GetAIMetrics()
	if aiMetrics != nil {
		fmt.Printf("   🤖 AI predictions: %d\n", int(aiMetrics["inference_count"]))
		fmt.Printf("   ⚡ Avg inference time: %.2f ms\n", aiMetrics["avg_inference_time"])
		fmt.Printf("   💾 Model confidence: %.2f\n", aiMetrics["model_confidence"])
	}
	
	// Get version info
	versionInfo := aiValidator.GetVersionInfo()
	fmt.Printf("   📦 Validator version: %s\n", versionInfo["validator_version"])
	fmt.Printf("   🎯 Consensus type: %s\n", versionInfo["consensus_type"])
	fmt.Printf("   🤖 AI enabled: %s\n", versionInfo["ai_enabled"])
	fmt.Printf("   🔐 Quantum enabled: %s\n", versionInfo["quantum_enabled"])
	
	// Get plugin stats
	pm := plugin.Instance()
	stats := pm.GetPluginStats()
	fmt.Printf("   📦 Total plugins: %d\n", stats.TotalPlugins)
	fmt.Printf("   🔝 Max plugins: %d\n", stats.MaxPlugins)
	fmt.Printf("   🏷️  Plugin types: %v\n", stats.PluginTypes)
	
	return nil
}

// Additional utility functions for demonstration

func demonstrateConfiguration() {
	fmt.Println("\n🔧 Configuration Demo")
	
	// Model configuration
	modelConfig := plugin.DefaultModelConfig()
	configJSON, _ := json.MarshalIndent(modelConfig, "", "  ")
	fmt.Printf("Model Configuration:\n%s\n", string(configJSON))
	
	// Signer configuration
	signerConfig := plugin.DefaultSignerConfig()
	signerJSON, _ := json.MarshalIndent(signerConfig, "", "  ")
	fmt.Printf("Signer Configuration:\n%s\n", string(signerJSON))
}

func demonstrateErrorHandling() {
	fmt.Println("\n⚠️  Error Handling Demo")
	
	// Test with invalid configuration
	invalidConfig := &validator.AIValidatorConfig{
		EnableAIPrediction: true,
		PluginPath:         "./nonexistent_plugin.so",
	}
	
	_, err := validator.NewAIValidator(invalidConfig)
	if err != nil {
		fmt.Printf("   ✅ Properly handled invalid plugin path: %v\n", err)
	} else {
		fmt.Println("   ❌ Should have failed with invalid plugin path")
	}
}

func demonstratePerformance() {
	fmt.Println("\n⚡ Performance Demo")
	
	// Generate large transaction set
	largeTxs := generateSampleTransactions(1000)
	fmt.Printf("   📊 Generated %d transactions for performance test\n", len(largeTxs))
	
	// Time the prediction
	start := time.Now()
	
	// Simulate prediction (in real implementation, this would use the AI model)
	_ = largeTxs[:100] // Take subset for demo
	
	elapsed := time.Since(start)
	fmt.Printf("   ⏱️  Prediction time: %v\n", elapsed)
	fmt.Printf("   🚀 Throughput: %.0f tx/s\n", float64(100)/elapsed.Seconds())
}

func demonstrateAdvancedFeatures() {
	fmt.Println("\n🚀 Advanced Features Demo")
	
	// Demonstrate batch prediction with different transaction types
	transactions := []*types.Transaction{
		{Hash: "tx1", Type: "transfer", Gas: 21000, Value: 1000000000000000000},
		{Hash: "tx2", Type: "contract", Gas: 50000, Value: 500000000000000000},
		{Hash: "tx3", Type: "delegate", Gas: 30000, Value: 2000000000000000000},
		{Hash: "tx4", Type: "transfer", Gas: 21000, Value: 1000000000000000000},
		{Hash: "tx5", Type: "contract", Gas: 60000, Value: 300000000000000000},
	}
	
	fmt.Printf("   📝 Sample transactions: %d\n", len(transactions))
	
	// Group by type
	typeGroups := make(map[string][]*types.Transaction)
	for _, tx := range transactions {
		typeGroups[tx.Type] = append(typeGroups[tx.Type], tx)
	}
	
	for txType, txs := range typeGroups {
		fmt.Printf("   📋 %s transactions: %d\n", txType, len(txs))
	}
	
	// Calculate gas efficiency
	totalGas := uint64(0)
	for _, tx := range transactions {
		totalGas += tx.Gas
	}
	
	avgGas := float64(totalGas) / float64(len(transactions))
	fmt.Printf("   ⛽ Average gas per transaction: %.0f\n", avgGas)
	fmt.Printf("   💰 Total value: %.2f ETH\n", float64(totalValue(transactions))/1e18)
}

func totalValue(transactions []*types.Transaction) uint64 {
	total := uint64(0)
	for _, tx := range transactions {
		total += tx.Value
	}
	return total
} 