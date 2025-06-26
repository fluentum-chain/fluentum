package plugin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fluentum-chain/fluentum/types"
)

func TestAIValidatorPlugin(t *testing.T) {
	// Test AI validator plugin interface compliance
	t.Run("Interface Compliance", func(t *testing.T) {
		// Create a mock AI validator plugin
		mockPlugin := &MockAIValidatorPlugin{
			initialized: false,
			metrics:     &ModelMetrics{},
		}

		// Test that it implements the interface
		var _ AIValidatorPlugin = mockPlugin
	})

	t.Run("Batch Prediction", func(t *testing.T) {
		// Create test transactions
		txs := []*types.Transaction{
			{
				Hash:      "tx1",
				From:      "0x123",
				To:        "0x456",
				Value:     1000000000000000000, // 1 ETH
				Gas:       21000,
				GasPrice:  20000000000, // 20 Gwei
				Nonce:     1,
				Type:      "transfer",
				Data:      []byte{},
			},
			{
				Hash:      "tx2",
				From:      "0x789",
				To:        "0xabc",
				Value:     500000000000000000, // 0.5 ETH
				Gas:       50000,
				GasPrice:  25000000000, // 25 Gwei
				Nonce:     2,
				Type:      "contract",
				Data:      []byte{0x60, 0x60, 0x60}, // Some contract data
			},
		}

		// Create mock plugin
		plugin := &MockAIValidatorPlugin{
			initialized: true,
			metrics:     &ModelMetrics{},
		}

		// Test batch prediction
		prediction, err := plugin.PredictBatch(txs)
		if err != nil {
			t.Fatalf("Failed to predict batch: %v", err)
		}

		// Verify prediction
		if prediction == nil {
			t.Fatal("Prediction should not be nil")
		}

		if len(prediction.OptimalBatch) == 0 {
			t.Fatal("Optimal batch should not be empty")
		}

		if prediction.Confidence <= 0 {
			t.Fatal("Confidence should be positive")
		}

		// Test priority groups
		if len(prediction.PriorityGroups) == 0 {
			t.Fatal("Priority groups should not be empty")
		}

		// Test pattern groups
		if len(prediction.PatternGroups) == 0 {
			t.Fatal("Pattern groups should not be empty")
		}
	})

	t.Run("Batch Validation", func(t *testing.T) {
		// Create test batch
		batch := &types.Batch{
			Transactions: []*types.Transaction{
				{
					Hash: "tx1",
					Gas:  21000,
				},
				{
					Hash: "tx2",
					Gas:  50000,
				},
			},
		}

		// Create mock plugin
		plugin := &MockAIValidatorPlugin{
			initialized: true,
			metrics:     &ModelMetrics{},
		}

		// Test batch validation
		valid, confidence, err := plugin.ValidateBatch(batch)
		if err != nil {
			t.Fatalf("Failed to validate batch: %v", err)
		}

		if !valid {
			t.Fatal("Batch should be valid")
		}

		if confidence <= 0 {
			t.Fatal("Confidence should be positive")
		}
	})

	t.Run("Model Metrics", func(t *testing.T) {
		plugin := &MockAIValidatorPlugin{
			initialized: true,
			metrics: &ModelMetrics{
				InferenceCount:    10,
				AvgInferenceTime:  time.Millisecond * 5,
				GasSavings:        0.4, // 40%
				PredictionAccuracy: 0.95,
			},
		}

		metrics := plugin.GetModelMetrics()
		if metrics == nil {
			t.Fatal("Metrics should not be nil")
		}

		if metrics["inference_count"] != 10 {
			t.Fatal("Inference count should be 10")
		}

		if metrics["gas_savings"] != 0.4 {
			t.Fatal("Gas savings should be 0.4")
		}
	})

	t.Run("Version Info", func(t *testing.T) {
		plugin := &MockAIValidatorPlugin{
			initialized: true,
		}

		info := plugin.VersionInfo()
		if info == nil {
			t.Fatal("Version info should not be nil")
		}

		if info["version"] == "" {
			t.Fatal("Version should not be empty")
		}

		if info["model_type"] == "" {
			t.Fatal("Model type should not be empty")
		}
	})
}

func TestQuantumSigner(t *testing.T) {
	t.Run("Signer Creation", func(t *testing.T) {
		config := DefaultSignerConfig()
		signer := NewQuantumSigner(config)

		if signer == nil {
			t.Fatal("Signer should not be nil")
		}

		if signer.config == nil {
			t.Fatal("Config should not be nil")
		}
	})

	t.Run("Key Pair Generation", func(t *testing.T) {
		config := DefaultSignerConfig()
		signer := NewQuantumSigner(config)

		keyPair, err := signer.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		if keyPair == nil {
			t.Fatal("Key pair should not be nil")
		}

		if len(keyPair.PrivateKey) == 0 {
			t.Fatal("Private key should not be empty")
		}

		if len(keyPair.PublicKey) == 0 {
			t.Fatal("Public key should not be empty")
		}

		if keyPair.Algorithm != config.Algorithm {
			t.Fatal("Algorithm should match config")
		}

		if keyPair.ID == "" {
			t.Fatal("Key ID should not be empty")
		}
	})

	t.Run("Signing and Verification", func(t *testing.T) {
		config := DefaultSignerConfig()
		signer := NewQuantumSigner(config)

		// Generate key pair
		keyPair, err := signer.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		// Test data
		data := []byte("Hello, Fluentum!")

		// Sign data
		signature, err := signer.Sign(data)
		if err != nil {
			t.Fatalf("Failed to sign data: %v", err)
		}

		if len(signature) == 0 {
			t.Fatal("Signature should not be empty")
		}

		// Verify signature
		valid, err := signer.Verify(data, signature, keyPair.PublicKey)
		if err != nil {
			t.Fatalf("Failed to verify signature: %v", err)
		}

		if !valid {
			t.Fatal("Signature should be valid")
		}

		// Test with wrong data
		wrongData := []byte("Wrong data")
		valid, err = signer.Verify(wrongData, signature, keyPair.PublicKey)
		if err != nil {
			t.Fatalf("Failed to verify signature: %v", err)
		}

		if valid {
			t.Fatal("Signature should be invalid for wrong data")
		}
	})

	t.Run("Supported Algorithms", func(t *testing.T) {
		config := DefaultSignerConfig()
		signer := NewQuantumSigner(config)

		algorithms := signer.GetSupportedAlgorithms()
		if len(algorithms) == 0 {
			t.Fatal("Should have supported algorithms")
		}

		// Check for expected algorithms
		expectedAlgorithms := []string{"dilithium3", "rsa"}
		for _, expected := range expectedAlgorithms {
			found := false
			for _, actual := range algorithms {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Expected algorithm %s not found", expected)
			}
		}
	})

	t.Run("Signer Info", func(t *testing.T) {
		config := DefaultSignerConfig()
		signer := NewQuantumSigner(config)

		info := signer.GetSignerInfo()
		if info == nil {
			t.Fatal("Signer info should not be nil")
		}

		if info["algorithm"] != config.Algorithm {
			t.Fatal("Algorithm should match config")
		}

		if info["quantum_resistant"] != "true" {
			t.Fatal("Should be quantum resistant")
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		config := DefaultSignerConfig()
		signer := NewQuantumSigner(config)

		// Perform some operations
		data := []byte("Test data")
		signer.Sign(data)
		signer.Sign(data)

		metrics := signer.GetMetrics()
		if metrics == nil {
			t.Fatal("Metrics should not be nil")
		}

		if metrics["sign_count"] != 2 {
			t.Fatal("Sign count should be 2")
		}

		if metrics["error_count"] != 0 {
			t.Fatal("Error count should be 0")
		}
	})
}

func TestPluginManager(t *testing.T) {
	t.Run("Manager Creation", func(t *testing.T) {
		manager := Instance()
		if manager == nil {
			t.Fatal("Plugin manager should not be nil")
		}
	})

	t.Run("Configuration", func(t *testing.T) {
		manager := Instance()
		config := DefaultPluginManagerConfig()

		err := manager.Initialize(config)
		if err != nil {
			t.Fatalf("Failed to initialize plugin manager: %v", err)
		}

		managerConfig := manager.GetConfig()
		if managerConfig == nil {
			t.Fatal("Manager config should not be nil")
		}

		if managerConfig.PluginDirectory != config.PluginDirectory {
			t.Fatal("Plugin directory should match")
		}
	})

	t.Run("Plugin Registration", func(t *testing.T) {
		manager := Instance()

		// Create mock plugin
		mockPlugin := &MockAIValidatorPlugin{
			initialized: true,
		}

		// Register plugin
		err := manager.RegisterPlugin("test_plugin", mockPlugin)
		if err != nil {
			t.Fatalf("Failed to register plugin: %v", err)
		}

		// Get plugin
		plugin, exists := manager.GetPlugin("test_plugin")
		if !exists {
			t.Fatal("Plugin should exist")
		}

		if plugin == nil {
			t.Fatal("Plugin should not be nil")
		}

		// List plugins
		plugins := manager.ListPlugins()
		if len(plugins) == 0 {
			t.Fatal("Should have registered plugins")
		}

		found := false
		for _, p := range plugins {
			if p == "test_plugin" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("Should find test_plugin in list")
		}
	})

	t.Run("Plugin Stats", func(t *testing.T) {
		manager := Instance()

		stats := manager.GetPluginStats()
		if stats == nil {
			t.Fatal("Stats should not be nil")
		}

		if stats.TotalPlugins < 0 {
			t.Fatal("Total plugins should be non-negative")
		}

		if stats.MaxPlugins <= 0 {
			t.Fatal("Max plugins should be positive")
		}
	})
}

func TestBatchPrediction(t *testing.T) {
	t.Run("Prediction Creation", func(t *testing.T) {
		prediction := NewBatchPrediction()
		if prediction == nil {
			t.Fatal("Prediction should not be nil")
		}

		if prediction.PriorityGroups == nil {
			t.Fatal("Priority groups should be initialized")
		}

		if prediction.PatternGroups == nil {
			t.Fatal("Pattern groups should be initialized")
		}
	})

	t.Run("Transaction Addition", func(t *testing.T) {
		prediction := NewBatchPrediction()

		tx := &types.Transaction{
			Hash: "test_tx",
			Gas:  21000,
		}

		prediction.AddTransaction(tx, 5)
		if len(prediction.OptimalBatch) != 1 {
			t.Fatal("Should have one transaction in optimal batch")
		}

		if len(prediction.PriorityGroups[5]) != 1 {
			t.Fatal("Should have one transaction in priority group 5")
		}
	})

	t.Run("Pattern Group Addition", func(t *testing.T) {
		prediction := NewBatchPrediction()

		tx := &types.Transaction{
			Hash: "test_tx",
			Type: "transfer",
		}

		prediction.AddPatternGroup("value_transfer", tx)
		if len(prediction.PatternGroups["value_transfer"]) != 1 {
			t.Fatal("Should have one transaction in pattern group")
		}
	})

	t.Run("Gas Savings Calculation", func(t *testing.T) {
		prediction := &BatchPrediction{
			EstimatedGas: 80000,
		}

		originalGas := uint64(100000)
		savings := prediction.CalculateGasSavings(originalGas)

		expectedSavings := float64(originalGas-prediction.EstimatedGas) / float64(originalGas) * 100.0
		if savings != expectedSavings {
			t.Fatalf("Gas savings calculation incorrect. Expected: %f, Got: %f", expectedSavings, savings)
		}
	})
}

func TestModelConfig(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		config := DefaultModelConfig()

		if config.NumExperts != 8 {
			t.Fatal("Default num experts should be 8")
		}

		if config.InputSize != 256 {
			t.Fatal("Default input size should be 256")
		}

		if config.ConfidenceThreshold != 0.7 {
			t.Fatal("Default confidence threshold should be 0.7")
		}

		if !config.EnableSparseActivation {
			t.Fatal("Default sparse activation should be enabled")
		}
	})

	t.Run("Config Serialization", func(t *testing.T) {
		config := DefaultModelConfig()

		// Test JSON serialization
		data, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		var decoded ModelConfig
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		if decoded.NumExperts != config.NumExperts {
			t.Fatal("Config should be preserved after serialization")
		}
	})
}

// MockAIValidatorPlugin implements AIValidatorPlugin for testing
type MockAIValidatorPlugin struct {
	initialized bool
	metrics     *ModelMetrics
}

func (m *MockAIValidatorPlugin) Initialize(config map[string]interface{}) error {
	m.initialized = true
	return nil
}

func (m *MockAIValidatorPlugin) PredictBatch(transactions []*types.Transaction) (*BatchPrediction, error) {
	prediction := NewBatchPrediction()

	// Add all transactions to optimal batch
	prediction.OptimalBatch = transactions
	prediction.Confidence = 0.85
	prediction.EstimatedGas = 70000
	prediction.GasSavings = 0.4

	// Create priority groups
	for i, tx := range transactions {
		priority := (i % 10) + 1
		prediction.AddTransaction(tx, priority)
	}

	// Create pattern groups
	for _, tx := range transactions {
		pattern := "standard_transaction"
		if tx.Type == "transfer" {
			pattern = "value_transfer"
		} else if tx.Type == "contract" {
			pattern = "contract_execution"
		}
		prediction.AddPatternGroup(pattern, tx)
	}

	return prediction, nil
}

func (m *MockAIValidatorPlugin) PredictBatchAsync(ctx context.Context, transactions []*types.Transaction) (*BatchPrediction, error) {
	return m.PredictBatch(transactions)
}

func (m *MockAIValidatorPlugin) ValidateBatch(batch *types.Batch) (bool, float64, error) {
	return true, 0.9, nil
}

func (m *MockAIValidatorPlugin) ValidateBatchAsync(ctx context.Context, batch *types.Batch) (bool, float64, error) {
	return m.ValidateBatch(batch)
}

func (m *MockAIValidatorPlugin) PredictExecutionPattern(tx *types.Transaction) (string, error) {
	if tx.Type == "transfer" {
		return "value_transfer", nil
	} else if tx.Type == "contract" {
		return "contract_execution", nil
	}
	return "standard_transaction", nil
}

func (m *MockAIValidatorPlugin) EstimateCombinedGasSavings(batch []*types.Transaction) (float64, error) {
	return 0.4, nil // 40% savings
}

func (m *MockAIValidatorPlugin) GetModelMetrics() map[string]float64 {
	if m.metrics == nil {
		return map[string]float64{}
	}

	return map[string]float64{
		"inference_count":     float64(m.metrics.InferenceCount),
		"avg_inference_time":  m.metrics.AvgInferenceTime.Seconds(),
		"gas_savings":         m.metrics.GasSavings,
		"prediction_accuracy": m.metrics.PredictionAccuracy,
	}
}

func (m *MockAIValidatorPlugin) VersionInfo() map[string]string {
	return map[string]string{
		"version":        "1.0.0",
		"model_type":     "QMoE",
		"consensus_version": "QMoE v1.0",
		"quantization":   "4-bit",
		"experts":        "8",
		"top_k":          "2",
	}
}

func (m *MockAIValidatorPlugin) ResetMetrics() {
	m.metrics = &ModelMetrics{}
}

func (m *MockAIValidatorPlugin) UpdateWeights(weightsPath string) error {
	return nil
}

func (m *MockAIValidatorPlugin) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"num_experts": 8,
		"input_size":  256,
		"top_k":       2,
	}
}

// Import context package for async methods
import "context" 
