// AI Validator demo temporarily disabled.

package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fluentum-chain/fluentum/core/plugin"
	"github.com/fluentum-chain/fluentum/core/validator"
	"github.com/fluentum-chain/fluentum/types"
)

func main() {
	fmt.Println("🚀 Fluentum AI-Validation Core Demo")
	fmt.Println("=====================================")

	if err := runAIDemo(); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Demo completed successfully!")
}

func runAIDemo() error {
	// Initialize AI validator
	fmt.Println("\n🔧 Initializing AI Validator...")

	config := &validator.AIValidatorConfig{
		EnableAIPrediction:   true,
		EnableQuantumSigning: true,
		BatchSize:            50,
		MaxWaitTime:          5 * time.Second,
		ConfidenceThreshold:  0.7,
		GasSavingsThreshold:  0.3,
		PluginPath:           "./plugins/qmoe_validator.so",
		QuantumPluginPath:    "./plugins/quantum_signer.so",
		ModelConfig: map[string]interface{}{
			"num_experts":                  8,
			"input_size":                   256,
			"hidden_size":                  512,
			"output_size":                  128,
			"top_k":                        2,
			"quantization_bits":            4,
			"quantization_update_interval": 60.0,
			"weights_path":                 "./models/qmoe_fluentum.bin",
			"confidence_threshold":         0.7,
			"gas_savings_threshold":        0.3,
			"max_batch_size":               100,
			"min_batch_size":               5,
			"enable_sparse_activation":     true,
			"enable_dynamic_quantization":  true,
		},
	}

	aiValidator, err := validator.NewAIValidator(config)
	if err != nil {
		return fmt.Errorf("failed to create AI validator: %w", err)
	}

	fmt.Println("   ✅ AI Validator initialized successfully")

	// Generate sample transactions
	fmt.Println("\n📝 Generating sample transactions...")
	transactions := generateSampleTransactions(100)
	fmt.Printf("   ✅ Generated %d sample transactions\n", len(transactions))

	// Demonstrate AI prediction
	fmt.Println("\n🧠 AI Prediction Demo")
	if err := demonstrateAIPrediction(aiValidator, transactions); err != nil {
		return fmt.Errorf("AI prediction demo failed: %w", err)
	}

	// Demonstrate batch processing
	fmt.Println("\n⚙️  Batch Processing Demo")
	if err := demonstrateBatchProcessing(aiValidator, transactions); err != nil {
		return fmt.Errorf("batch processing demo failed: %w", err)
	}

	// Demonstrate quantum signing
	fmt.Println("\n🔐 Quantum Signing Demo")
	if err := demonstrateQuantumSigning(aiValidator); err != nil {
		return fmt.Errorf("quantum signing demo failed: %w", err)
	}

	// Demonstrate metrics
	fmt.Println("\n📊 Metrics Demo")
	if err := demonstrateMetrics(aiValidator); err != nil {
		return fmt.Errorf("metrics demo failed: %w", err)
	}

	// Demonstrate configuration
	demonstrateConfiguration()

	// Demonstrate error handling
	demonstrateErrorHandling()

	// Demonstrate performance
	demonstratePerformance()

	// Demonstrate advanced features
	demonstrateAdvancedFeatures()

	return nil
}

func generateSampleTransactions(count int) []types.Tx {
	transactions := make([]types.Tx, count)

	for i := 0; i < count; i++ {
		// Create a simple transaction with random data
		txData := make([]byte, 32)
		rand.Read(txData)

		// Add some metadata to make it more realistic
		txData = append(txData, byte(i%10)) // Transaction type
		txData = append(txData, byte(i%5))  // Priority

		transactions[i] = types.Tx(txData)
	}

	return transactions
}

func demonstrateAIPrediction(aiValidator *validator.AIValidator, transactions []types.Tx) error {
	// Take a subset for prediction
	predictionTxs := transactions[:20]

	// Get AI prediction
	prediction, err := aiValidator.PredictOptimalBatch(predictionTxs)
	if err != nil {
		return fmt.Errorf("failed to get AI prediction: %w", err)
	}

	fmt.Printf("   📈 Original transactions: %d\n", len(predictionTxs))
	fmt.Printf("   🎯 Optimized batch: %d transactions\n", len(prediction))

	// Calculate gas savings (simplified - in real implementation this would be more complex)
	originalGas := uint64(len(predictionTxs) * 21000) // Assume 21k gas per tx
	optimizedGas := uint64(len(prediction) * 21000)

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

func demonstrateBatchProcessing(aiValidator *validator.AIValidator, transactions []types.Tx) error {
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

	// Create a block with the sample transactions
	block := &types.Block{
		Header: types.Header{
			Height: 1,
			Time:   time.Now(),
		},
		Data: types.Data{
			Txs: sampleBatch,
		},
	}

	if err := aiValidator.ProcessBlock(block); err != nil {
		return fmt.Errorf("failed to process block: %w", err)
	}

	fmt.Printf("   ✅ Block processed successfully\n")

	return nil
}

func demonstrateQuantumSigning(aiValidator *validator.AIValidator) error {
	fmt.Println("\n🔐 Quantum Signing Demo")

	// Get plugin manager
	pm := plugin.Instance()
	if pm.GetPluginCount() == 0 {
		fmt.Println("   ⚠️  No plugins loaded, skipping quantum signing demo")
		return nil
	}

	// Load quantum signer
	signer, err := pm.GetSigner()
	if err != nil {
		return fmt.Errorf("failed to get quantum signer: %w", err)
	}

	// Get signer information
	fmt.Printf("   🔐 Signer algorithm: %s\n", signer.AlgorithmName())
	fmt.Printf("   🛡️  Quantum resistant: %t\n", signer.IsQuantumResistant())
	fmt.Printf("   🔑 Key size: %d bytes\n", signer.PublicKeySize())

	// Get performance metrics
	metrics := signer.PerformanceMetrics()
	fmt.Printf("   📊 Performance metrics: %v\n", metrics)

	// Generate key pair for testing
	privateKey, publicKey, err := signer.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Test signing
	testData := []byte("Fluentum AI-Validation Core Test")
	signature, err := signer.Sign(privateKey, testData)
	if err != nil {
		return fmt.Errorf("failed to sign test data: %w", err)
	}

	fmt.Printf("   ✍️  Signature length: %d bytes\n", len(signature))

	// Verify signature
	valid, err := signer.Verify(publicKey, testData, signature)
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
	if versionInfo != nil {
		fmt.Printf("   📦 Validator version: %s\n", versionInfo["version"])
		fmt.Printf("   🎯 Model type: %s\n", versionInfo["model_type"])
	}

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

	// Start batch processor
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config := &validator.AIValidatorConfig{
		EnableAIPrediction: false, // Disable for performance test
		BatchSize:          100,
		MaxWaitTime:        1 * time.Second,
	}

	aiValidator, err := validator.NewAIValidator(config)
	if err != nil {
		fmt.Printf("   ❌ Failed to create validator: %v\n", err)
		return
	}

	// Generate test transactions
	transactions := generateSampleTransactions(1000)

	// Start batch processor
	aiValidator.StartBatchProcessor(ctx)

	// Add transactions
	start := time.Now()
	for _, tx := range transactions {
		aiValidator.AddTransaction(tx)
	}

	elapsed := time.Since(start)
	fmt.Printf("   ⏱️  Added %d transactions in %v\n", len(transactions), elapsed)
	fmt.Printf("   📊 Average time per transaction: %v\n", elapsed/time.Duration(len(transactions)))

	// Get final metrics
	metrics := aiValidator.GetMetrics()
	fmt.Printf("   📝 Total transactions processed: %d\n", metrics.TransactionsProcessed)
}

func demonstrateAdvancedFeatures() {
	fmt.Println("\n🚀 Advanced Features Demo")

	// Test transaction validation
	config := &validator.AIValidatorConfig{
		EnableAIPrediction: false,
	}

	aiValidator, err := validator.NewAIValidator(config)
	if err != nil {
		fmt.Printf("   ❌ Failed to create validator: %v\n", err)
		return
	}

	// Test transaction validation
	testTx := types.Tx([]byte("test transaction"))
	valid, err := aiValidator.ValidateTransaction(testTx)
	if err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Transaction validation: %t\n", valid)
	}

	// Test batch queue management
	queueSize := aiValidator.GetBatchQueueSize()
	fmt.Printf("   📋 Initial queue size: %d\n", queueSize)

	// Test configuration update
	newConfig := &validator.AIValidatorConfig{
		EnableAIPrediction: false,
		BatchSize:          200,
		MaxWaitTime:        2 * time.Second,
	}

	if err := aiValidator.UpdateConfig(newConfig); err != nil {
		fmt.Printf("   ❌ Config update failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Configuration updated successfully")
	}

	// Test metrics reset
	aiValidator.ResetMetrics()
	fmt.Println("   ✅ Metrics reset successfully")
}

func totalValue(transactions []types.Tx) uint64 {
	// Simplified calculation - in real implementation this would parse transaction data
	return uint64(len(transactions) * 1000)
}
