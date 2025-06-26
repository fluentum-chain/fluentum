package plugin

import (
	"context"
	"time"

	"github.com/fluentum-chain/fluentum/types"
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
	TotalGasSaved      uint64        `
