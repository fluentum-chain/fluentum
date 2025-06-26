package validator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fluentum-chain/fluentum/types"
	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
)

// AIValidator integrates QMoE consensus into Fluentum validator nodes
type AIValidator struct {
	aiPlugin    plugin.AIValidatorPlugin
	signer      plugin.SignerPlugin
	mutex       sync.RWMutex
	config      *AIValidatorConfig
	metrics     *ValidatorMetrics
	initialized bool
	
	// Batch processing
	batchQueue  []*types.Transaction
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
	EnableAIPrediction    bool          `json:"enable_ai_prediction"`
	EnableQuantumSigning  bool          `json:"enable_quantum_signing"`
	BatchSize             int           `json:"batch_size"`
	MaxWaitTime           time.Duration `json:"max_wait_time"`
	ConfidenceThreshold   float64       `json:"confidence_threshold"`
	GasSavingsThreshold   float64       `json:"gas_savings_threshold"`
	PluginPath            string        `json:"plugin_path"`
	QuantumPluginPath     string        `json:"quantum_plugin_path"`
	ModelConfig           map[string]interface{} `json:"model_config"`
}

// ValidatorMetrics tracks validator performance metrics
type ValidatorMetrics struct {
	BlocksProcessed    int64         `json:"blocks_processed"`
	TransactionsProcessed int64      `json:"transactions_processed"`
	AvgBlockTime       time.Duration `json:"avg_block_time"`
	TotalGasSaved      uint64        `json:"total_gas_saved"`
	AvgGasSavings      float64       `json:"avg_gas_savings"`
	PredictionAccuracy float64       `json:"prediction_accuracy"`
	LastUpdate         time.Time     `json:"last_update"`
	
	// AI-specific metrics
	AIPredictions      int64         `json:"ai_predictions"`
	AvgPredictionTime  time.Duration `json:"avg_prediction_time"`
	ModelConfidence    float64       `json:"model_confidence"`
	BatchEfficiency    float64       `json:"batch_efficiency"`
}

// NewAIValidator creates a new AI-powered validator
func NewAIValidator(config *AIValidatorConfig) (*AIValidator, error) {
	v := &AIValidator{
		config:      config,
		batchQueue:  make([]*types.Transaction, 0),
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
		prediction, err = v.aiPlugin.PredictBatch(block.Transactions)
		if err != nil {
			return fmt.Errorf("AI prediction failed: %w", err)
		}
		
		// Update metrics
		v.metrics.AIPredictions++
		v.metrics.AvgPredictionTime = (v.metrics.AvgPredictionTime*time.Duration(v.metrics.AIPredictions-1) + time.Since(start)) / time.Duration(v.metrics.AIPredictions)
		v.metrics.ModelConfidence = prediction.Confidence
	}
	
	// Create optimized batch
	optimizedBatch := v.createOptimizedBatch(prediction, block.Transactions)
	
	// Validate batch with AI if available
	if v.aiPlugin != nil {
		valid, confidence, err := v.aiPlugin.ValidateBatch(&types.Batch{
			Transactions: optimizedBatch,
		})
		if err != nil {
			return fmt.Errorf("AI validation failed: %w", err)
		}
		
		if !valid {
			return fmt.Errorf("AI validation rejected batch (confidence: %.2f)", confidence)
		}
		
		// Update prediction accuracy
		v.updatePredictionAccuracy(confidence)
	}
	
	// Sign the block
	if v.signer != nil {
		blockData, err := v.serializeBlock(optimizedBatch)
		if err != nil {
			return fmt.Errorf("failed to serialize block: %w", err)
		}
		
		signature, err := v.signer.Sign(blockData)
		if err != nil {
			return fmt.Errorf("quantum signing failed: %w", err)
		}
		
		block.Signature = signature
	}
	
	// Update block metadata
	block.Confidence = prediction.Confidence
	block.GasSavings = prediction.GasSavings
	block.OptimizedBatch = optimizedBatch
	
	// Update metrics
	v.updateMetrics(start, prediction)
	
	return nil
}

// CreateOptimizedBatch creates an optimized transaction batch
func (v *AIValidator) createOptimizedBatch(prediction *plugin.BatchPrediction, originalTxs []*types.Transaction) []*types.Transaction {
	if prediction == nil || !prediction.IsValid() {
		// Fallback to original transactions
		return originalTxs
	}
	
	optimizedBatch := make([]*types.Transaction, 0)
	
	// Process priority groups in order (highest priority first)
	for priority := 10; priority >= 1; priority-- {
		if txs, ok := prediction.PriorityGroups[priority]; ok {
			optimizedBatch = append(optimizedBatch, txs...)
			
			// Early exit if we've reached optimal batch size
			if len(optimizedBatch) >= v.config.BatchSize {
				break
			}
		}
	}
	
	// If we don't have enough transactions, add from pattern groups
	if len(optimizedBatch) < v.config.MinBatchSize() {
		for pattern, txs := range prediction.PatternGroups {
			if len(optimizedBatch) >= v.config.BatchSize {
				break
			}
			
			// Add transactions from this pattern
			for _, tx := range txs {
				if !v.containsTransaction(optimizedBatch, tx) {
					optimizedBatch = append(optimizedBatch, tx)
					if len(optimizedBatch) >= v.config.BatchSize {
						break
					}
				}
			}
		}
	}
	
	// If still not enough, add remaining transactions
	if len(optimizedBatch) < v.config.MinBatchSize() {
		for _, tx := range originalTxs {
			if !v.containsTransaction(optimizedBatch, tx) {
				optimizedBatch = append(optimizedBatch, tx)
				if len(optimizedBatch) >= v.config.BatchSize {
					break
				}
			}
		}
	}
	
	return optimizedBatch
}

// ContainsTransaction checks if a transaction is already in the batch
func (v *AIValidator) containsTransaction(batch []*types.Transaction, tx *types.Transaction) bool {
	for _, batchTx := range batch {
		if batchTx.Hash == tx.Hash {
			return true
		}
	}
	return false
}

// SerializeBlock serializes block data for signing
func (v *AIValidator) serializeBlock(transactions []*types.Transaction) ([]byte, error) {
	// Create block data structure
	blockData := &types.BlockData{
		Transactions: transactions,
		Timestamp:    time.Now().Unix(),
		Version:      "1.0",
	}
	
	// Serialize to bytes
	data, err := blockData.Serialize()
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

// UpdateMetrics updates validator metrics
func (v *AIValidator) updateMetrics(start time.Time, prediction *plugin.BatchPrediction) {
	elapsed := time.Since(start)
	
	v.metrics.BlocksProcessed++
	v.metrics.AvgBlockTime = (v.metrics.AvgBlockTime*time.Duration(v.metrics.BlocksProcessed-1) + elapsed) / time.Duration(v.metrics.BlocksProcessed)
	
	if prediction != nil {
		v.metrics.TotalGasSaved += uint64(prediction.GasSavings * 100) // Convert percentage to integer
		v.metrics.AvgGasSavings = float64(v.metrics.TotalGasSaved) / float64(v.metrics.BlocksProcessed)
		v.metrics.BatchEfficiency = float64(len(prediction.OptimalBatch)) / float64(len(prediction.OptimalBatch)+len(prediction.PriorityGroups))
	}
	
	v.metrics.LastUpdate = time.Now()
}

// UpdatePredictionAccuracy updates prediction accuracy metrics
func (v *AIValidator) updatePredictionAccuracy(confidence float64) {
	// Simple exponential moving average for prediction accuracy
	alpha := 0.1
	v.metrics.PredictionAccuracy = alpha*confidence + (1-alpha)*v.metrics.PredictionAccuracy
}

// GetMetrics returns current validator metrics
func (v *AIValidator) GetMetrics() *ValidatorMetrics {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
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
	info := map[string]string{
		"validator_version": "1.0.0",
		"consensus_type":    "QMoE + DPoS",
		"ai_enabled":        fmt.Sprintf("%t", v.aiPlugin != nil),
		"quantum_enabled":   fmt.Sprintf("%t", v.signer != nil),
	}
	
	if v.aiPlugin != nil {
		aiInfo := v.aiPlugin.VersionInfo()
		for k, v := range aiInfo {
			info["ai_"+k] = v
		}
	}
	
	return info
}

// AddTransaction adds a transaction to the batch queue
func (v *AIValidator) AddTransaction(tx *types.Transaction) error {
	v.batchMutex.Lock()
	defer v.batchMutex.Unlock()
	
	// Add transaction to queue
	v.batchQueue = append(v.batchQueue, tx)
	
	// Process batch if full or timeout reached
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
		prediction, err = v.aiPlugin.PredictBatch(v.batchQueue)
		if err != nil {
			return fmt.Errorf("batch prediction failed: %w", err)
		}
	}
	
	// Create optimized batch
	optimizedBatch := v.createOptimizedBatch(prediction, v.batchQueue)
	
	// Create block
	block := &types.Block{
		Transactions: optimizedBatch,
		Timestamp:    time.Now(),
		Validator:    v.config.ValidatorAddress,
	}
	
	// Process block
	if err := v.ProcessBlock(block); err != nil {
		return fmt.Errorf("block processing failed: %w", err)
	}
	
	// Clear batch queue
	v.batchQueue = v.batchQueue[:0]
	
	return nil
}

// StartBatchProcessor starts the background batch processor
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

// PredictOptimalBatch predicts optimal batch composition
func (v *AIValidator) PredictOptimalBatch(txs []*types.Transaction) ([]*types.Transaction, error) {
	if v.aiPlugin == nil {
		return txs, nil // Fallback to original transactions
	}
	
	prediction, err := v.aiPlugin.PredictBatch(txs)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}
	
	return prediction.OptimalBatch, nil
}

// EstimateGasSavings estimates gas savings for a batch
func (v *AIValidator) EstimateGasSavings(txs []*types.Transaction) (float64, error) {
	if v.aiPlugin == nil {
		return 0.0, nil
	}
	
	return v.aiPlugin.EstimateCombinedGasSavings(txs)
}

// ValidateTransaction validates a single transaction
func (v *AIValidator) ValidateTransaction(tx *types.Transaction) (bool, error) {
	// Basic validation
	if tx == nil {
		return false, fmt.Errorf("transaction is nil")
	}
	
	if tx.Gas == 0 {
		return false, fmt.Errorf("transaction has zero gas")
	}
	
	if len(tx.From) == 0 {
		return false, fmt.Errorf("transaction has no sender")
	}
	
	// AI-based validation if available
	if v.aiPlugin != nil {
		pattern, err := v.aiPlugin.PredictExecutionPattern(tx)
		if err != nil {
			return false, fmt.Errorf("pattern prediction failed: %w", err)
		}
		
		// Check if pattern is suspicious
		if pattern == "suspicious" {
			return false, fmt.Errorf("transaction pattern flagged as suspicious")
		}
	}
	
	return true, nil
}

// GetBatchQueueSize returns the current batch queue size
func (v *AIValidator) GetBatchQueueSize() int {
	v.batchMutex.Lock()
	defer v.batchMutex.Unlock()
	return len(v.batchQueue)
}

// ResetMetrics resets all metrics
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

// UpdateConfig updates validator configuration
func (v *AIValidator) UpdateConfig(config *AIValidatorConfig) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Update configuration
	v.config = config
	v.batchSize = config.BatchSize
	v.maxWaitTime = config.MaxWaitTime
	
	// Reinitialize AI plugin if needed
	if config.EnableAIPrediction && v.aiPlugin == nil {
		if err := v.loadAIPlugin(); err != nil {
			return fmt.Errorf("failed to load AI plugin: %w", err)
		}
	}
	
	// Reinitialize quantum signer if needed
	if config.EnableQuantumSigning && v.signer == nil {
		if err := v.loadQuantumSigner(); err != nil {
			return fmt.Errorf("failed to load quantum signer: %w", err)
		}
	}
	
	return nil
}

// IsInitialized returns whether the validator is initialized
func (v *AIValidator) IsInitialized() bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.initialized
}

// GetConfig returns current configuration
func (v *AIValidator) GetConfig() *AIValidatorConfig {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.config
}

// MinBatchSize returns the minimum batch size
func (c *AIValidatorConfig) MinBatchSize() int {
	if c.BatchSize < 5 {
		return 5
	}
	return c.BatchSize / 2
} 
