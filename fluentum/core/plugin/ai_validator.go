package plugin

import (
	"context"
	"time"

	"github.com/fluentum-chain/fluentum/fluentum/core/types"
)

// AIValidatorPlugin interface for QMoE integration
type AIValidatorPlugin interface {
	// Initialize the model with configuration
	Initialize(config map[string]interface{}) error
	
	// Predict optimal batch composition
	PredictBatch(transactions []*types.Transaction) (*BatchPrediction, error)
	
	// Predict optimal batch composition asynchronously
	PredictBatchAsync(ctx context.Context, transactions []*types.Transaction) (*BatchPrediction, error)
	
	// Validate transaction batch using QMoE
	ValidateBatch(batch *types.Batch) (bool, float64, error)
	
	// Validate transaction batch asynchronously
	ValidateBatchAsync(ctx context.Context, batch *types.Batch) (bool, float64, error)
	
	// Predict execution pattern for a transaction
	PredictExecutionPattern(tx *types.Transaction) (string, error)
	
	// Estimate combined gas savings for a batch
	EstimateCombinedGasSavings(batch []*types.Transaction) (float64, error)
	
	// Model metrics and statistics
	GetModelMetrics() map[string]float64
	
	// Model version and info
	VersionInfo() map[string]string
	
	// Reset model metrics
	ResetMetrics()
	
	// Update model weights
	UpdateWeights(weightsPath string) error
	
	// Get model configuration
	GetConfig() map[string]interface{}
}

// BatchPrediction represents the AI model's prediction for optimal batch composition
type BatchPrediction struct {
	OptimalBatch   []*types.Transaction `json:"optimal_batch"`
	Confidence     float64              `json:"confidence"`
	EstimatedGas   uint64               `json:"estimated_gas"`
	PriorityGroups map[int][]*types.Transaction `json:"priority_groups"`
	ExecutionTime  time.Duration        `json:"execution_time"`
	GasSavings     float64              `json:"gas_savings"`
	PatternGroups  map[string][]*types.Transaction `json:"pattern_groups"`
}

// ModelConfig contains configuration for the AI validator
type ModelConfig struct {
	NumExperts                int     `json:"num_experts"`
	InputSize                 int     `json:"input_size"`
	HiddenSize                int     `json:"hidden_size"`
	OutputSize                int     `json:"output_size"`
	TopK                      int     `json:"top_k"`
	QuantizationBits          int     `json:"quantization_bits"`
	QuantizationUpdateInterval float64 `json:"quantization_update_interval"`
	WeightsPath               string  `json:"weights_path"`
	ConfidenceThreshold       float64 `json:"confidence_threshold"`
	GasSavingsThreshold       float64 `json:"gas_savings_threshold"`
	MaxBatchSize              int     `json:"max_batch_size"`
	MinBatchSize              int     `json:"min_batch_size"`
	EnableSparseActivation    bool    `json:"enable_sparse_activation"`
	EnableDynamicQuantization bool    `json:"enable_dynamic_quantization"`
}

// DefaultModelConfig returns default configuration for QMoE model
func DefaultModelConfig() ModelConfig {
	return ModelConfig{
		NumExperts:                8,
		InputSize:                 256,
		HiddenSize:                512,
		OutputSize:                128,
		TopK:                      2,
		QuantizationBits:          4,
		QuantizationUpdateInterval: 60.0,
		WeightsPath:               "./models/qmoe_fluentum.bin",
		ConfidenceThreshold:       0.7,
		GasSavingsThreshold:       0.3,
		MaxBatchSize:              100,
		MinBatchSize:              5,
		EnableSparseActivation:    true,
		EnableDynamicQuantization: true,
	}
}

// ModelMetrics tracks performance metrics for the AI model
type ModelMetrics struct {
	InferenceCount     int64         `json:"inference_count"`
	AvgInferenceTime   time.Duration `json:"avg_inference_time"`
	TotalInferenceTime time.Duration `json:"total_inference_time"`
	AccuracyHistory    []float64     `json:"accuracy_history"`
	GasSavings         float64       `json:"gas_savings"`
	TotalGasSaved      uint64        `json:"total_gas_saved"`
	PredictionAccuracy float64       `json:"prediction_accuracy"`
	ModelLoadTime      time.Duration `json:"model_load_time"`
	LastUpdate         time.Time     `json:"last_update"`
}

// ExecutionPattern represents a predicted execution pattern for transactions
type ExecutionPattern struct {
	PatternID   string  `json:"pattern_id"`
	Description string  `json:"description"`
	GasCost     uint64  `json:"gas_cost"`
	Confidence  float64 `json:"confidence"`
	Frequency   float64 `json:"frequency"`
}

// AIValidatorError represents AI validator specific errors
type AIValidatorError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AIValidatorError) Error() string {
	return e.Message
}

// Common AI validator error codes
const (
	ErrCodeModelNotInitialized = "MODEL_NOT_INITIALIZED"
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeInferenceFailed     = "INFERENCE_FAILED"
	ErrCodeModelLoadFailed     = "MODEL_LOAD_FAILED"
	ErrCodeQuantizationFailed  = "QUANTIZATION_FAILED"
	ErrCodeLowConfidence       = "LOW_CONFIDENCE"
	ErrCodeInsufficientGas     = "INSUFFICIENT_GAS"
)

// AIValidatorStats contains comprehensive statistics
type AIValidatorStats struct {
	ModelMetrics    *ModelMetrics            `json:"model_metrics"`
	Config          *ModelConfig             `json:"config"`
	VersionInfo     map[string]string        `json:"version_info"`
	Performance     map[string]float64       `json:"performance"`
	GasOptimization map[string]interface{}   `json:"gas_optimization"`
	LastReset       time.Time                `json:"last_reset"`
}

// NewBatchPrediction creates a new batch prediction
func NewBatchPrediction() *BatchPrediction {
	return &BatchPrediction{
		PriorityGroups: make(map[int][]*types.Transaction),
		PatternGroups:  make(map[string][]*types.Transaction),
		Confidence:     0.0,
		EstimatedGas:   0,
		GasSavings:     0.0,
	}
}

// AddTransaction adds a transaction to the optimal batch
func (bp *BatchPrediction) AddTransaction(tx *types.Transaction, priority int) {
	bp.OptimalBatch = append(bp.OptimalBatch, tx)
	
	if bp.PriorityGroups[priority] == nil {
		bp.PriorityGroups[priority] = make([]*types.Transaction, 0)
	}
	bp.PriorityGroups[priority] = append(bp.PriorityGroups[priority], tx)
}

// AddPatternGroup adds a transaction to a pattern group
func (bp *BatchPrediction) AddPatternGroup(pattern string, tx *types.Transaction) {
	if bp.PatternGroups[pattern] == nil {
		bp.PatternGroups[pattern] = make([]*types.Transaction, 0)
	}
	bp.PatternGroups[pattern] = append(bp.PatternGroups[pattern], tx)
}

// GetTotalTransactions returns the total number of transactions
func (bp *BatchPrediction) GetTotalTransactions() int {
	return len(bp.OptimalBatch)
}

// GetPriorityGroup returns transactions for a specific priority
func (bp *BatchPrediction) GetPriorityGroup(priority int) []*types.Transaction {
	return bp.PriorityGroups[priority]
}

// GetPatternGroup returns transactions for a specific pattern
func (bp *BatchPrediction) GetPatternGroup(pattern string) []*types.Transaction {
	return bp.PatternGroups[pattern]
}

// IsValid checks if the batch prediction is valid
func (bp *BatchPrediction) IsValid() bool {
	return len(bp.OptimalBatch) > 0 && bp.Confidence > 0.0
}

// CalculateGasSavings calculates the gas savings percentage
func (bp *BatchPrediction) CalculateGasSavings(originalGas uint64) float64 {
	if originalGas == 0 {
		return 0.0
	}
	return float64(originalGas-bp.EstimatedGas) / float64(originalGas) * 100.0
} 