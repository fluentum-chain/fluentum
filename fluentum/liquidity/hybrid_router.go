package liquidity

import (
	"context"
	"math"
	"time"

	fluentumtypes "github.com/fluentum-chain/fluentum/fluentum/types"
	"github.com/fluentum-chain/fluentum/fluentum/x/cex"
	"github.com/fluentum-chain/fluentum/fluentum/x/dex"
)

// Router handles order routing between CEX and DEX
type Router struct {
	cexClient    *cex.Client
	dexClient    *dex.Client
	threshold    int64
	lastUpdate   time.Time
	updatePeriod time.Duration
}

// NewRouter creates a new hybrid liquidity router
func NewRouter(cexClient *cex.Client, dexClient *dex.Client) *Router {
	return &Router{
		cexClient:    cexClient,
		dexClient:    dexClient,
		threshold:    1000000000, // Initial threshold: 10 FLUX
		lastUpdate:   time.Now(),
		updatePeriod: 5 * time.Minute,
	}
}

// RouteOrder routes an order to either CEX or DEX based on best price and dynamic threshold
func (r *Router) RouteOrder(ctx context.Context, order fluentumtypes.Order) error {
	// Update threshold if needed
	if time.Since(r.lastUpdate) > r.updatePeriod {
		r.updateThreshold()
	}

	// Quote both venues
	cexPrice, errCEX := r.cexClient.QuoteBestPrice(ctx, order)
	dexPrice, errDEX := r.dexClient.QuoteBestPrice(ctx, order)

	// Select best price (lower for buy, higher for sell)
	venue := ""
	var execErr error
	if (order.Side == "buy" && (errCEX == nil && (errDEX != nil || cexPrice <= dexPrice))) ||
		(order.Side == "sell" && (errCEX == nil && (errDEX != nil || cexPrice >= dexPrice))) {
		execErr = r.cexClient.ExecuteOrder(ctx, order)
		venue = "CEX"
	} else if errDEX == nil {
		execErr = r.dexClient.ExecuteOrder(ctx, order)
		venue = "DEX"
	} else {
		return errCEX // or errDEX, both failed
	}
	if execErr != nil {
		return execErr
	}
	return r.settleOrder(ctx, order, venue)
}

// settleOrder records the unified trade in the settlement layer
func (r *Router) settleOrder(ctx context.Context, order fluentumtypes.Order, venue string) error {
	// TODO: Implement settlement logic (e.g., update trade records, emit event, etc.)
	// Example: log.Printf("Settled order %s on %s", order.ID, venue)
	return nil
}

// updateThreshold calculates new threshold based on market conditions
func (r *Router) updateThreshold() {
	// Get current market conditions
	cexLiquidity := r.cexClient.GetTotalLiquidity()
	dexLiquidity := r.dexClient.GetTotalLiquidity()
	cexFees := r.cexClient.GetAverageFees()
	dexFees := r.dexClient.GetAverageFees()

	// Calculate optimal threshold based on:
	// 1. Relative liquidity between CEX and DEX
	// 2. Fee differential
	// 3. Historical order sizes
	liquidityRatio := float64(cexLiquidity) / float64(dexLiquidity)
	feeRatio := float64(cexFees) / float64(dexFees)

	// Adjust threshold based on market conditions
	adjustment := math.Sqrt(liquidityRatio * feeRatio)
	newThreshold := int64(float64(r.threshold) * adjustment)

	// Apply bounds to prevent extreme values
	minThreshold := int64(100000000)   // 1 FLUX
	maxThreshold := int64(10000000000) // 100 FLUX
	r.threshold = clamp(newThreshold, minThreshold, maxThreshold)
	r.lastUpdate = time.Now()
}

// clamp ensures a value stays within specified bounds
func clamp(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
