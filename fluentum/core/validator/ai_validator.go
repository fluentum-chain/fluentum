package validator

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
	"github.com/fluentum-chain/fluentum/types"
)

// TxAdapter adapts types.Tx to plugin.Transaction interface
type TxAdapter struct {
	tx types.Tx
}

func (t TxAdapter) GetData() []byte { return t.tx }
func (t TxAdapter) GetHash() []byte { return t.tx.Hash() }
func (t TxAdapter) GetSize() int    { return len(t.tx) }

// convertTxsToTransactions converts types.Txs to []plugin.Transaction
func convertTxsToTransactions(txs types.Txs) []plugin.Transaction {
	result := make([]plugin.Transaction, len(txs))
	for i, tx := range txs {
		result[i] = TxAdapter{tx: tx}
	}
	return result
}

// AIValidator integrates QMoE consensus into Fluentum validator nodes
type AIValidator struct {
	aiPlugin    plugin.AIValidatorPlugin
	signer      plugin.SignerPlugin
	mutex       sync.RWMutex
	config      *AIValidatorConfig
	metrics     *ValidatorMetrics
	initialized bool

	// Batch processing
	batchQueue  []types.Tx
	batchMutex  sync.Mutex
	batchSize   int
	maxWaitTime time.Duration

	// Performance tracking
	lastBlockTime time.Time
	blockCount    int64
	gasSavings    float64
}

// AIValidatorConfig contains configuration for AI validator
type AIValidatorConfig struct {
	EnableAIPrediction   bool                   `json:"enable_ai_prediction"`
	EnableQuantumSigning bool                   `json:"enable_quantum_signing"`
	BatchSize            int                    `json:"batch_size"`
	MaxWaitTime          time.Duration          `json:"max_wait_time"`
	ConfidenceThreshold  float64                `json:"confidence_threshold"`
	GasSavingsThreshold  float64                `json:"gas_savings_threshold"`
	PluginPath           string                 `json:"plugin_path"`
	QuantumPluginPath    string                 `json:"quantum_plugin_path"`
	ModelConfig          map[string]interface{} `json:"model_config"`
}

// ValidatorMetrics tracks validator performance metrics
type ValidatorMetrics struct {
	BlocksProcessed       int64         `json:"blocks_processed"`
	TransactionsProcessed int64         `json:"transactions_processed"`
	AvgBlockTime          time.Duration `json:"avg_block_time"`
	TotalGasSaved         uint64        `json:"total_gas_saved"`
	AvgGasSavings         float64       `json:"avg_gas_savings"`
	PredictionAccuracy    float64       `json:"prediction_accuracy"`
	LastUpdate            time.Time     `json:"last_update"`

	// AI-specific metrics
	AIPredictions     int64         `json:"ai_predictions"`
	AvgPredictionTime time.Duration `json:"avg_prediction_time"`
	ModelConfidence   float64       `json:"model_confidence"`
	BatchEfficiency   float64       `json:"batch_efficiency"`
}

// NewAIValidator creates a new AI-powered validator
func NewAIValidator(config *AIValidatorConfig) (*AIValidator, error) {
	v := &AIValidator{
		config:      config,
		batchQueue:  make([]types.Tx, 0),
		batchSize:   config.BatchSize,
		maxWaitTime: config.MaxWaitTime,
		metrics: &ValidatorMetrics{
			LastUpdate: time.Now(),
		},
	}

	// Load AI validation plugin
	if config.EnableAIPrediction {
		if err := v.loadAIPlugin(); err != nil {
			return nil, fmt.Errorf("failed to load AI plugin: %w", err)
		}
	}

	// Load quantum signing plugin
	if config.EnableQuantumSigning {
		if err := v.loadQuantumSigner(); err != nil {
			return nil, fmt.Errorf("failed to load quantum signer: %w", err)
		}
	}

	v.initialized = true
	return v, nil
}

// LoadAIPlugin loads the AI validation plugin
func (v *AIValidator) loadAIPlugin() error {
	pm := plugin.Instance()

	// Load AI plugin
	aiPlugin, err := pm.LoadPlugin(v.config.PluginPath, "AIValidatorPlugin")
	if err != nil {
		return err
	}

	// Cast to AIValidatorPlugin interface
	if aiValidator, ok := aiPlugin.(plugin.AIValidatorPlugin); ok {
		v.aiPlugin = aiValidator

		// Initialize the model
		if err := aiValidator.Initialize(v.config.ModelConfig); err != nil {
			return fmt.Errorf("failed to initialize AI model: %w", err)
		}

		return nil
	}

	return fmt.Errorf("plugin does not implement AIValidatorPlugin interface")
}

// LoadQuantumSigner loads the quantum signing plugin
func (v *AIValidator) loadQuantumSigner() error {
	pm := plugin.Instance()

	// Load quantum signer plugin
	signerPlugin, err := pm.LoadPlugin(v.config.QuantumPluginPath, "SignerPlugin")
	if err != nil {
		return err
	}

	// Cast to SignerPlugin interface
	if signer, ok := signerPlugin.(plugin.SignerPlugin); ok {
		v.signer = signer
		return nil
	}

	return fmt.Errorf("plugin does not implement SignerPlugin interface")
}

// ProcessBlock processes a block using AI prediction and quantum signing
func (v *AIValidator) ProcessBlock(block *types.Block) error {
	if !v.initialized {
		return fmt.Errorf("AI validator not initialized")
	}

	v.mutex.Lock()
	defer v.mutex.Unlock()

	start := time.Now()

	// Get AI batch prediction
	var prediction *plugin.BatchPrediction
	var err error

	if v.aiPlugin != nil {
		transactions := convertTxsToTransactions(block.Data.Txs)
		prediction, err = v.aiPlugin.PredictBatch(transactions)
		if err != nil {
			return fmt.Errorf("AI prediction failed: %w", err)
		}

		// Update metrics
		v.metrics.AIPredictions++
		v.metrics.AvgPredictionTime = (v.metrics.AvgPredictionTime*time.Duration(v.metrics.AIPredictions-1) + time.Since(start)) / time.Duration(v.metrics.AIPredictions)
		v.metrics.ModelConfidence = prediction.Confidence
	}

	// Create optimized batch
	optimizedBatch := v.createOptimizedBatch(prediction, block.Data.Txs)

	// Validate batch with AI if available
	if v.aiPlugin != nil {
		transactions := convertTxsToTransactions(optimizedBatch)
		batch := &plugin.Batch{
			Transactions: transactions,
			Hash:         optimizedBatch.Hash(),
			Size:         len(optimizedBatch),
		}

		valid, confidence, err := v.aiPlugin.ValidateBatch(batch)
		if err != nil {
			return fmt.Errorf("AI validation failed: %w", err)
		}

		if !valid {
			return fmt.Errorf("AI validation rejected batch (confidence: %.2f)", confidence)
		}

		// Update prediction accuracy
		v.updatePredictionAccuracy(confidence)
	}

	// Sign the block (note: we don't modify the block directly as it doesn't have signature fields)
	if v.signer != nil {
		blockData, err := v.serializeBlock(optimizedBatch)
		if err != nil {
			return fmt.Errorf("failed to serialize block: %w", err)
		}

		// Generate a private key for signing (in real implementation, this would come from validator)
		privateKey := make([]byte, v.signer.PrivateKeySize())
		signature, err := v.signer.Sign(privateKey, blockData)
		if err != nil {
			return fmt.Errorf("quantum signing failed: %w", err)
		}

		// Store signature in metrics or separate storage
		_ = signature // Use signature as needed
	}

	// Update block metadata
	v.updateMetrics(start, prediction)

	return nil
}

// CreateOptimizedBatch creates an optimized transaction batch
func (v *AIValidator) createOptimizedBatch(prediction *plugin.BatchPrediction, originalTxs types.Txs) types.Txs {
	if prediction == nil {
		// No AI prediction available, return original batch
		return originalTxs
	}

	// Create optimized batch based on AI prediction
	optimizedBatch := make(types.Txs, 0, len(prediction.OptimalBatch))

	// Convert plugin.Transaction back to types.Tx
	for _, tx := range prediction.OptimalBatch {
		if txAdapter, ok := tx.(TxAdapter); ok {
			optimizedBatch = append(optimizedBatch, txAdapter.tx)
		}
	}

	// Add any remaining transactions that weren't in the prediction
	for _, tx := range originalTxs {
		if !v.containsTransaction(optimizedBatch, tx) {
			optimizedBatch = append(optimizedBatch, tx)
		}
	}

	return optimizedBatch
}

// ContainsTransaction checks if a transaction is already in the batch
func (v *AIValidator) containsTransaction(batch types.Txs, tx types.Tx) bool {
	for _, batchTx := range batch {
		if bytes.Equal(batchTx, tx) {
			return true
		}
	}
	return false
}

// SerializeBlock serializes block data for signing
func (v *AIValidator) serializeBlock(transactions types.Txs) ([]byte, error) {
	// Simple serialization - concatenate all transaction hashes
	var data []byte
	for _, tx := range transactions {
		data = append(data, tx.Hash()...)
	}
	return data, nil
}

// UpdateMetrics updates validator metrics
func (v *AIValidator) updateMetrics(start time.Time, prediction *plugin.BatchPrediction) {
	v.metrics.BlocksProcessed++
	v.metrics.TransactionsProcessed += int64(len(v.batchQueue))

	blockTime := time.Since(start)
	v.metrics.AvgBlockTime = (v.metrics.AvgBlockTime*time.Duration(v.metrics.BlocksProcessed-1) + blockTime) / time.Duration(v.metrics.BlocksProcessed)

	if prediction != nil {
		v.metrics.TotalGasSaved += uint64(prediction.GasSavings)
		v.metrics.AvgGasSavings = float64(v.metrics.TotalGasSaved) / float64(v.metrics.BlocksProcessed)
	}

	v.metrics.LastUpdate = time.Now()
}

// UpdatePredictionAccuracy updates prediction accuracy metrics
func (v *AIValidator) updatePredictionAccuracy(confidence float64) {
	v.metrics.PredictionAccuracy = (v.metrics.PredictionAccuracy*float64(v.metrics.AIPredictions-1) + confidence) / float64(v.metrics.AIPredictions)
}

// GetMetrics returns current validator metrics
func (v *AIValidator) GetMetrics() *ValidatorMetrics {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *v.metrics
	return &metrics
}

// GetAIMetrics returns AI-specific metrics
func (v *AIValidator) GetAIMetrics() map[string]float64 {
	if v.aiPlugin == nil {
		return nil
	}
	return v.aiPlugin.GetModelMetrics()
}

// GetVersionInfo returns version information
func (v *AIValidator) GetVersionInfo() map[string]string {
	if v.aiPlugin == nil {
		return nil
	}
	return v.aiPlugin.VersionInfo()
}

// AddTransaction adds a transaction to the batch queue
func (v *AIValidator) AddTransaction(tx types.Tx) error {
	if !v.initialized {
		return fmt.Errorf("AI validator not initialized")
	}

	v.batchMutex.Lock()
	defer v.batchMutex.Unlock()

	// Add transaction to queue
	v.batchQueue = append(v.batchQueue, tx)

	// Process batch if size threshold reached
	if len(v.batchQueue) >= v.batchSize {
		return v.processBatch()
	}

	return nil
}

// ProcessBatch processes the current batch queue
func (v *AIValidator) processBatch() error {
	if len(v.batchQueue) == 0 {
		return nil
	}

	// Get AI prediction for batch
	var prediction *plugin.BatchPrediction
	var err error

	if v.aiPlugin != nil {
		transactions := convertTxsToTransactions(v.batchQueue)
		prediction, err = v.aiPlugin.PredictBatch(transactions)
		if err != nil {
			return fmt.Errorf("AI batch prediction failed: %w", err)
		}
	}

	// Create optimized batch
	optimizedBatch := v.createOptimizedBatch(prediction, v.batchQueue)

	// Clear the queue
	v.batchQueue = v.batchQueue[:0]

	// Update metrics
	v.metrics.TransactionsProcessed += int64(len(optimizedBatch))
	if prediction != nil {
		v.metrics.TotalGasSaved += uint64(prediction.GasSavings)
	}

	return nil
}

// StartBatchProcessor starts the batch processing goroutine
func (v *AIValidator) StartBatchProcessor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(v.maxWaitTime)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				v.batchMutex.Lock()
				if len(v.batchQueue) > 0 {
					v.processBatch()
				}
				v.batchMutex.Unlock()
			}
		}
	}()
}

// PredictOptimalBatch predicts the optimal batch composition for given transactions
func (v *AIValidator) PredictOptimalBatch(txs types.Txs) (types.Txs, error) {
	if v.aiPlugin == nil {
		return txs, fmt.Errorf("AI plugin not available")
	}

	transactions := convertTxsToTransactions(txs)
	prediction, err := v.aiPlugin.PredictBatch(transactions)
	if err != nil {
		return nil, fmt.Errorf("AI prediction failed: %w", err)
	}

	return v.createOptimizedBatch(prediction, txs), nil
}

// EstimateGasSavings estimates gas savings for a batch of transactions
func (v *AIValidator) EstimateGasSavings(txs types.Txs) (float64, error) {
	if v.aiPlugin == nil {
		return 0, fmt.Errorf("AI plugin not available")
	}

	transactions := convertTxsToTransactions(txs)
	return v.aiPlugin.EstimateCombinedGasSavings(transactions)
}

// ValidateTransaction validates a single transaction using AI
func (v *AIValidator) ValidateTransaction(tx types.Tx) (bool, error) {
	if v.aiPlugin == nil {
		return true, nil // No AI validation available
	}

	// Create a single-transaction batch for validation
	transactions := []plugin.Transaction{TxAdapter{tx: tx}}
	batch := &plugin.Batch{
		Transactions: transactions,
		Hash:         tx.Hash(),
		Size:         1,
	}

	valid, _, err := v.aiPlugin.ValidateBatch(batch)
	return valid, err
}

// GetBatchQueueSize returns the current size of the batch queue
func (v *AIValidator) GetBatchQueueSize() int {
	v.batchMutex.Lock()
	defer v.batchMutex.Unlock()
	return len(v.batchQueue)
}

// ResetMetrics resets all metrics to zero
func (v *AIValidator) ResetMetrics() {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.metrics = &ValidatorMetrics{
		LastUpdate: time.Now(),
	}

	if v.aiPlugin != nil {
		v.aiPlugin.ResetMetrics()
	}
}

// UpdateConfig updates the validator configuration
func (v *AIValidator) UpdateConfig(config *AIValidatorConfig) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	// Update configuration
	v.config = config
	v.batchSize = config.BatchSize
	v.maxWaitTime = config.MaxWaitTime

	// Reload plugins if paths changed
	if config.EnableAIPrediction && v.aiPlugin == nil {
		if err := v.loadAIPlugin(); err != nil {
			return fmt.Errorf("failed to reload AI plugin: %w", err)
		}
	}

	if config.EnableQuantumSigning && v.signer == nil {
		if err := v.loadQuantumSigner(); err != nil {
			return fmt.Errorf("failed to reload quantum signer: %w", err)
		}
	}

	return nil
}

// IsInitialized returns whether the validator is initialized
func (v *AIValidator) IsInitialized() bool {
	return v.initialized
}

// GetConfig returns the current configuration
func (v *AIValidator) GetConfig() *AIValidatorConfig {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	// Return a copy to avoid race conditions
	config := *v.config
	return &config
}

// MinBatchSize returns the minimum batch size for processing
func (c *AIValidatorConfig) MinBatchSize() int {
	if c.BatchSize < 1 {
		return 1
	}
	return c.BatchSize
}
