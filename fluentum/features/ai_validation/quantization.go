// AI Validation temporarily disabled.

package ai_validation

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// DynamicQuantizer implements adaptive quantization for QMoE models
type DynamicQuantizer struct {
	bits               int
	threshold          float64
	updateInterval     float64
	lastUpdate         time.Time
	observationWindow  []float64
	maxObservationSize int
	mutex              sync.RWMutex

	// Quantization parameters
	scale     float64
	zeroPoint float64
	minValue  float64
	maxValue  float64

	// Adaptive parameters
	learningRate float64
	momentum     float64
	emaAlpha     float64
	emaValue     float64

	// Statistics
	quantizationError float64
	compressionRatio  float64
	updateCount       int
}

// QuantizationConfig holds configuration for the quantizer
type QuantizationConfig struct {
	Bits               int     `json:"bits"`
	UpdateInterval     float64 `json:"update_interval"`
	MaxObservationSize int     `json:"max_observation_size"`
	AdaptiveThreshold  bool    `json:"adaptive_threshold"`
	HistogramBins      int     `json:"histogram_bins"`
	InitialThreshold   float64 `json:"initial_threshold"`
}

// DefaultQuantizationConfig returns default quantization settings
func DefaultQuantizationConfig() QuantizationConfig {
	return QuantizationConfig{
		Bits:               4,
		UpdateInterval:     60.0,
		MaxObservationSize: 1000,
		AdaptiveThreshold:  true,
		HistogramBins:      256,
		InitialThreshold:   0.7,
	}
}

// NewDynamicQuantizer creates a new dynamic quantizer
func NewDynamicQuantizer(bits int, updateInterval float64) *DynamicQuantizer {
	scale := math.Pow(2, float64(bits-1)) - 1

	return &DynamicQuantizer{
		bits:               bits,
		threshold:          0.7, // Initial threshold
		updateInterval:     updateInterval,
		maxObservationSize: 1000,
		scale:              scale,
		zeroPoint:          0.0,
		minValue:           -1.0,
		maxValue:           1.0,
		learningRate:       0.001,
		momentum:           0.9,
		emaAlpha:           0.95,
		emaValue:           0.0,
		observationWindow:  make([]float64, 0),
		lastUpdate:         time.Now(),
	}
}

// Quantize applies dynamic quantization to model outputs
func (q *DynamicQuantizer) Quantize(outputs [][]float64) [][]float64 {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	now := time.Now()
	if now.Sub(q.lastUpdate).Seconds() > q.updateInterval {
		q.updateQuantizationParameters()
		q.lastUpdate = now
	}

	quantized := make([][]float64, len(outputs))

	for i, output := range outputs {
		quantized[i] = make([]float64, len(output))

		for j, value := range output {
			// Apply quantization
			quantizedValue := q.quantizeValue(value)
			quantized[i][j] = quantizedValue

			// Record for threshold adjustment
			if len(q.observationWindow) < q.maxObservationSize {
				q.observationWindow = append(q.observationWindow, value)
			}

			// Update exponential moving average
			q.emaValue = q.emaAlpha*q.emaValue + (1-q.emaAlpha)*value
		}
	}

	return quantized
}

// QuantizeValue quantizes a single value
func (q *DynamicQuantizer) quantizeValue(value float64) float64 {
	// Clamp value to range
	clamped := math.Max(q.minValue, math.Min(q.maxValue, value))

	// Apply quantization
	quantized := math.Round((clamped-q.zeroPoint)*q.scale)/q.scale + q.zeroPoint

	// Calculate quantization error
	error := math.Abs(value - quantized)
	q.quantizationError = q.emaAlpha*q.quantizationError + (1-q.emaAlpha)*error

	return quantized
}

// UpdateQuantizationParameters updates quantization parameters based on observations
func (q *DynamicQuantizer) updateQuantizationParameters() {
	if len(q.observationWindow) == 0 {
		return
	}

	// Calculate statistics
	stats := q.calculateStatistics(q.observationWindow)

	// Update threshold based on distribution
	q.updateThreshold(stats)

	// Update quantization range
	q.updateQuantizationRange(stats)

	// Update scale and zero point
	q.updateScaleAndZeroPoint(stats)

	// Calculate compression ratio
	q.calculateCompressionRatio()

	// Reset observation window
	q.observationWindow = nil
	q.updateCount++
}

// CalculateStatistics calculates statistical measures from observations
func (q *DynamicQuantizer) calculateStatistics(values []float64) *QuantizationStats {
	stats := &QuantizationStats{
		Count: len(values),
	}

	if len(values) == 0 {
		return stats
	}

	// Calculate basic statistics
	var sum, sumSq float64
	minVal := values[0]
	maxVal := values[0]

	for _, v := range values {
		sum += v
		sumSq += v * v
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	stats.Mean = sum / float64(len(values))
	stats.Variance = (sumSq / float64(len(values))) - (stats.Mean * stats.Mean)
	stats.StdDev = math.Sqrt(stats.Variance)
	stats.Min = minVal
	stats.Max = maxVal
	stats.Range = maxVal - minVal

	// Calculate percentiles
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	stats.P25 = q.percentile(sorted, 0.25)
	stats.P50 = q.percentile(sorted, 0.50)
	stats.P75 = q.percentile(sorted, 0.75)
	stats.P95 = q.percentile(sorted, 0.95)
	stats.P99 = q.percentile(sorted, 0.99)

	return stats
}

// Percentile calculates the nth percentile from sorted values
func (q *DynamicQuantizer) percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0.0
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// UpdateThreshold updates the confidence threshold based on statistics
func (q *DynamicQuantizer) updateThreshold(stats *QuantizationStats) {
	// Adaptive threshold based on distribution characteristics
	meanConfidence := stats.Mean
	stdDev := stats.StdDev

	// Adjust threshold based on mean confidence and variability
	newThreshold := meanConfidence + 0.5*stdDev

	// Apply momentum update
	q.threshold = q.momentum*q.threshold + (1-q.momentum)*newThreshold

	// Clamp to reasonable range
	q.threshold = math.Max(0.5, math.Min(0.95, q.threshold))
}

// UpdateQuantizationRange updates the quantization range based on statistics
func (q *DynamicQuantizer) updateQuantizationRange(stats *QuantizationStats) {
	// Use percentile-based range to handle outliers
	range95 := stats.P95 - stats.P5

	// Choose range based on desired precision vs. compression
	chosenRange := range95 // Use 95th percentile range for better compression

	// Add small margin for stability
	margin := chosenRange * 0.1
	newMin := stats.P5 - margin
	newMax := stats.P95 + margin

	// Apply momentum update
	q.minValue = q.momentum*q.minValue + (1-q.momentum)*newMin
	q.maxValue = q.momentum*q.maxValue + (1-q.momentum)*newMax
}

// UpdateScaleAndZeroPoint updates scale and zero point for optimal quantization
func (q *DynamicQuantizer) updateScaleAndZeroPoint(stats *QuantizationStats) {
	// Calculate optimal zero point for symmetric quantization
	optimalZeroPoint := (q.maxValue + q.minValue) / 2.0

	// Calculate scale based on range
	valueRange := q.maxValue - q.minValue
	optimalScale := q.scale / valueRange

	// Apply momentum update
	q.zeroPoint = q.momentum*q.zeroPoint + (1-q.momentum)*optimalZeroPoint
	q.scale = q.momentum*q.scale + (1-q.momentum)*optimalScale
}

// CalculateCompressionRatio calculates the compression ratio achieved
func (q *DynamicQuantizer) calculateCompressionRatio() {
	// Calculate theoretical compression ratio
	originalBits := 32.0 // Assuming float32
	compressedBits := float64(q.bits)

	q.compressionRatio = originalBits / compressedBits
}

// Threshold returns the current confidence threshold
func (q *DynamicQuantizer) Threshold() float64 {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.threshold
}

// GetQuantizationStats returns current quantization statistics
func (q *DynamicQuantizer) GetQuantizationStats() map[string]float64 {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return map[string]float64{
		"threshold":          q.threshold,
		"quantization_error": q.quantizationError,
		"compression_ratio":  q.compressionRatio,
		"scale":              q.scale,
		"zero_point":         q.zeroPoint,
		"min_value":          q.minValue,
		"max_value":          q.maxValue,
		"ema_value":          q.emaValue,
		"update_count":       float64(q.updateCount),
		"bits":               float64(q.bits),
	}
}

// SetQuantizationBits updates the number of quantization bits
func (q *DynamicQuantizer) SetQuantizationBits(bits int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.bits = bits
	q.scale = math.Pow(2, float64(bits-1)) - 1
}

// SetUpdateInterval updates the quantization update interval
func (q *DynamicQuantizer) SetUpdateInterval(interval float64) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.updateInterval = interval
}

// Reset resets the quantizer to initial state
func (q *DynamicQuantizer) Reset() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.threshold = 0.7
	q.observationWindow = nil
	q.quantizationError = 0.0
	q.compressionRatio = 0.0
	q.updateCount = 0
	q.emaValue = 0.0
	q.lastUpdate = time.Now()
}

// QuantizationStats holds statistical information for quantization
type QuantizationStats struct {
	Count    int
	Mean     float64
	Variance float64
	StdDev   float64
	Min      float64
	Max      float64
	Range    float64
	P1       float64
	P5       float64
	P25      float64
	P50      float64
	P75      float64
	P95      float64
	P99      float64
}

// AdaptiveQuantizer provides advanced adaptive quantization
type AdaptiveQuantizer struct {
	*DynamicQuantizer
	clusters    []*QuantizationCluster
	numClusters int
	clusterSize int
}

// QuantizationCluster represents a cluster of similar values
type QuantizationCluster struct {
	ID      int
	Center  float64
	Radius  float64
	Weight  float64
	Count   int
	Values  []float64
	Updated time.Time
}

// NewAdaptiveQuantizer creates a new adaptive quantizer with clustering
func NewAdaptiveQuantizer(bits int, updateInterval float64, numClusters int) *AdaptiveQuantizer {
	return &AdaptiveQuantizer{
		DynamicQuantizer: NewDynamicQuantizer(bits, updateInterval),
		numClusters:      numClusters,
		clusterSize:      100,
		clusters:         make([]*QuantizationCluster, 0),
	}
}

// QuantizeWithClustering applies clustering-based quantization
func (aq *AdaptiveQuantizer) QuantizeWithClustering(outputs [][]float64) [][]float64 {
	// Update clusters with new observations
	aq.updateClusters(outputs)

	// Apply cluster-based quantization
	quantized := make([][]float64, len(outputs))

	for i, output := range outputs {
		quantized[i] = make([]float64, len(output))

		for j, value := range output {
			// Find best cluster
			cluster := aq.findBestCluster(value)

			// Quantize using cluster-specific parameters
			quantized[i][j] = aq.quantizeWithCluster(value, cluster)
		}
	}

	return quantized
}

// UpdateClusters updates quantization clusters
func (aq *AdaptiveQuantizer) updateClusters(outputs [][]float64) {
	// Collect all values
	var allValues []float64
	for _, output := range outputs {
		allValues = append(allValues, output...)
	}

	if len(allValues) == 0 {
		return
	}

	// Initialize clusters if needed
	if len(aq.clusters) == 0 {
		aq.initializeClusters(allValues)
	}

	// Update existing clusters
	aq.updateExistingClusters(allValues)

	// Merge similar clusters
	aq.mergeSimilarClusters()

	// Split large clusters
	aq.splitLargeClusters()
}

// InitializeClusters initializes quantization clusters
func (aq *AdaptiveQuantizer) initializeClusters(values []float64) {
	// Use k-means++ initialization
	centers := aq.kMeansPlusPlus(values, aq.numClusters)

	aq.clusters = make([]*QuantizationCluster, len(centers))
	for i, center := range centers {
		aq.clusters[i] = &QuantizationCluster{
			ID:      i,
			Center:  center,
			Radius:  0.1,
			Weight:  1.0,
			Count:   0,
			Values:  make([]float64, 0),
			Updated: time.Now(),
		}
	}
}

// KMeansPlusPlus implements k-means++ initialization
func (aq *AdaptiveQuantizer) kMeansPlusPlus(values []float64, k int) []float64 {
	if len(values) == 0 || k <= 0 {
		return nil
	}

	centers := make([]float64, k)

	// Choose first center randomly
	centers[0] = values[0]

	// Choose remaining centers
	for i := 1; i < k; i++ {
		// Calculate distances to existing centers
		distances := make([]float64, len(values))
		for j, value := range values {
			minDist := math.Inf(1)
			for _, center := range centers[:i] {
				dist := math.Abs(value - center)
				if dist < minDist {
					minDist = dist
				}
			}
			distances[j] = minDist * minDist
		}

		// Choose next center with probability proportional to distance squared
		totalDist := 0.0
		for _, dist := range distances {
			totalDist += dist
		}

		r := rand.Float64() * totalDist
		cumDist := 0.0
		for j, dist := range distances {
			cumDist += dist
			if cumDist >= r {
				centers[i] = values[j]
				break
			}
		}
	}

	return centers
}

// FindBestCluster finds the best cluster for a value
func (aq *AdaptiveQuantizer) findBestCluster(value float64) *QuantizationCluster {
	if len(aq.clusters) == 0 {
		return nil
	}

	var bestCluster *QuantizationCluster
	minDistance := math.Inf(1)

	for _, cluster := range aq.clusters {
		distance := math.Abs(value - cluster.Center)
		if distance < minDistance {
			minDistance = distance
			bestCluster = cluster
		}
	}

	return bestCluster
}

// QuantizeWithCluster quantizes a value using cluster-specific parameters
func (aq *AdaptiveQuantizer) quantizeWithCluster(value float64, cluster *QuantizationCluster) float64 {
	if cluster == nil {
		return aq.quantizeValue(value)
	}

	// Use cluster-specific quantization
	clusterScale := cluster.Radius * aq.scale
	clusterZeroPoint := cluster.Center

	// Apply quantization
	clamped := math.Max(clusterZeroPoint-cluster.Radius, math.Min(clusterZeroPoint+cluster.Radius, value))
	quantized := math.Round((clamped-clusterZeroPoint)*clusterScale)/clusterScale + clusterZeroPoint

	return quantized
}

// UpdateExistingClusters updates existing clusters with new values
func (aq *AdaptiveQuantizer) updateExistingClusters(values []float64) {
	for _, value := range values {
		cluster := aq.findBestCluster(value)
		if cluster != nil {
			// Update cluster statistics
			cluster.Count++
			cluster.Values = append(cluster.Values, value)

			// Update center using exponential moving average
			alpha := 0.1
			cluster.Center = alpha*value + (1-alpha)*cluster.Center

			// Update radius
			if len(cluster.Values) > 1 {
				var sumSq float64
				for _, v := range cluster.Values {
					sumSq += (v - cluster.Center) * (v - cluster.Center)
				}
				cluster.Radius = math.Sqrt(sumSq / float64(len(cluster.Values)))
			}

			// Limit cluster size
			if len(cluster.Values) > aq.clusterSize {
				cluster.Values = cluster.Values[len(cluster.Values)-aq.clusterSize:]
			}
		}
	}
}

// MergeSimilarClusters merges clusters that are too similar
func (aq *AdaptiveQuantizer) mergeSimilarClusters() {
	if len(aq.clusters) < 2 {
		return
	}

	merged := make([]bool, len(aq.clusters))

	for i := 0; i < len(aq.clusters); i++ {
		if merged[i] {
			continue
		}

		for j := i + 1; j < len(aq.clusters); j++ {
			if merged[j] {
				continue
			}

			// Check if clusters are similar
			distance := math.Abs(aq.clusters[i].Center - aq.clusters[j].Center)
			combinedRadius := aq.clusters[i].Radius + aq.clusters[j].Radius

			if distance < combinedRadius*0.5 {
				// Merge clusters
				aq.mergeClusters(i, j)
				merged[j] = true
			}
		}
	}

	// Remove merged clusters
	newClusters := make([]*QuantizationCluster, 0)
	for i, cluster := range aq.clusters {
		if !merged[i] {
			newClusters = append(newClusters, cluster)
		}
	}
	aq.clusters = newClusters
}

// MergeClusters merges two clusters
func (aq *AdaptiveQuantizer) mergeClusters(i, j int) {
	cluster1 := aq.clusters[i]
	cluster2 := aq.clusters[j]

	// Calculate weighted center
	totalWeight := cluster1.Weight + cluster2.Weight
	newCenter := (cluster1.Center*cluster1.Weight + cluster2.Center*cluster2.Weight) / totalWeight

	// Calculate new radius
	newRadius := math.Max(cluster1.Radius, cluster2.Radius) * 1.2

	// Merge values
	mergedValues := append(cluster1.Values, cluster2.Values...)
	if len(mergedValues) > aq.clusterSize {
		mergedValues = mergedValues[len(mergedValues)-aq.clusterSize:]
	}

	// Update cluster1
	cluster1.Center = newCenter
	cluster1.Radius = newRadius
	cluster1.Weight = totalWeight
	cluster1.Count = cluster1.Count + cluster2.Count
	cluster1.Values = mergedValues
	cluster1.Updated = time.Now()
}

// SplitLargeClusters splits clusters that are too large
func (aq *AdaptiveQuantizer) splitLargeClusters() {
	for _, cluster := range aq.clusters {
		if cluster.Count > aq.clusterSize*2 && len(aq.clusters) < aq.numClusters*2 {
			// Split cluster
			newCluster := aq.splitCluster(cluster)
			if newCluster != nil {
				aq.clusters = append(aq.clusters, newCluster)
			}
		}
	}
}

// SplitCluster splits a large cluster into two
func (aq *AdaptiveQuantizer) splitCluster(cluster *QuantizationCluster) *QuantizationCluster {
	if len(cluster.Values) < 10 {
		return nil
	}

	// Find median value
	sorted := make([]float64, len(cluster.Values))
	copy(sorted, cluster.Values)
	sort.Float64s(sorted)
	median := sorted[len(sorted)/2]

	// Create new cluster
	newCluster := &QuantizationCluster{
		ID:      len(aq.clusters),
		Center:  median,
		Radius:  cluster.Radius * 0.8,
		Weight:  cluster.Weight * 0.5,
		Count:   0,
		Values:  make([]float64, 0),
		Updated: time.Now(),
	}

	// Redistribute values
	var cluster1Values, cluster2Values []float64
	for _, value := range cluster.Values {
		if value < median {
			cluster1Values = append(cluster1Values, value)
		} else {
			cluster2Values = append(cluster2Values, value)
		}
	}

	// Update clusters
	cluster.Values = cluster1Values
	cluster.Count = len(cluster1Values)
	cluster.Weight *= 0.5
	cluster.Center = cluster.Center * 0.8

	newCluster.Values = cluster2Values
	newCluster.Count = len(cluster2Values)

	return newCluster
}

// GetClusterStats returns statistics about quantization clusters
func (aq *AdaptiveQuantizer) GetClusterStats() map[string]interface{} {
	aq.mutex.RLock()
	defer aq.mutex.RUnlock()

	stats := map[string]interface{}{
		"num_clusters": len(aq.clusters),
		"clusters":     make([]map[string]interface{}, len(aq.clusters)),
	}

	for i, cluster := range aq.clusters {
		stats["clusters"].([]map[string]interface{})[i] = map[string]interface{}{
			"id":      cluster.ID,
			"center":  cluster.Center,
			"radius":  cluster.Radius,
			"weight":  cluster.Weight,
			"count":   cluster.Count,
			"updated": cluster.Updated,
		}
	}

	return stats
}
