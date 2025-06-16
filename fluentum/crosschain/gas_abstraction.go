package crosschain

import (
	"context"
	"errors"
	"math/big"
	"time"

	"fluentum/types"
	"fluentum/x/fluentum"
)

// GasConfig holds configuration for gas costs on different chains
type GasConfig struct {
	BaseGasPrice    *big.Int
	GasMultiplier   float64
	MinGasRequired  *big.Int
	MaxGasAllowed   *big.Int
	UpdateInterval  time.Duration
	LastUpdate      time.Time
}

// GasAbstraction handles cross-chain gas payments using FLU tokens
type GasAbstraction struct {
	keeper     *fluentum.Keeper
	configs    map[string]*GasConfig
	priceFeeds map[string]PriceFeed
}

// NewGasAbstraction creates a new gas abstraction handler
func NewGasAbstraction(keeper *fluentum.Keeper) *GasAbstraction {
	return &GasAbstraction{
		keeper:     keeper,
		configs:    make(map[string]*GasConfig),
		priceFeeds: make(map[string]PriceFeed),
	}
}

// PayGasWithFLU handles gas payment using FLU tokens
func (ga *GasAbstraction) PayGasWithFLU(ctx context.Context, tx types.Tx, chainID string) error {
	// Get or create gas config for chain
	config, err := ga.getGasConfig(chainID)
	if err != nil {
		return err
	}

	// Calculate gas cost in FLU
	gasCost, err := ga.calculateGasCost(ctx, tx, chainID, config)
	if err != nil {
		return err
	}

	// Verify sender has sufficient balance
	balance := ga.keeper.GetBalance(ctx, tx.Sender())
	if balance.Cmp(gasCost) < 0 {
		return errors.New("insufficient FLU balance for gas payment")
	}

	// Burn FLU tokens
	err = ga.keeper.BurnTokens(ctx, tx.Sender(), gasCost)
	if err != nil {
		return err
	}

	// Emit cross-chain event
	err = ga.emitGasPaymentEvent(ctx, tx, chainID, gasCost)
	if err != nil {
		// If event emission fails, refund the burned tokens
		_ = ga.keeper.MintTokens(ctx, tx.Sender(), gasCost)
		return err
	}

	return nil
}

// calculateGasCost calculates the gas cost in FLU tokens
func (ga *GasAbstraction) calculateGasCost(ctx context.Context, tx types.Tx, chainID string, config *GasConfig) (*big.Int, error) {
	// Get current gas price from price feed
	priceFeed, exists := ga.priceFeeds[chainID]
	if !exists {
		return nil, errors.New("price feed not found for chain")
	}

	gasPrice, err := priceFeed.GetGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate base cost
	baseCost := new(big.Int).Mul(gasPrice, big.NewInt(tx.GasLimit))

	// Apply multiplier
	multiplier := new(big.Float).SetFloat64(config.GasMultiplier)
	cost := new(big.Float).SetInt(baseCost)
	cost.Mul(cost, multiplier)

	// Convert to big.Int
	costInt := new(big.Int)
	cost.Int(costInt)

	// Apply bounds
	if costInt.Cmp(config.MinGasRequired) < 0 {
		costInt = config.MinGasRequired
	}
	if costInt.Cmp(config.MaxGasAllowed) > 0 {
		costInt = config.MaxGasAllowed
	}

	return costInt, nil
}

// getGasConfig returns the gas configuration for a chain
func (ga *GasAbstraction) getGasConfig(chainID string) (*GasConfig, error) {
	config, exists := ga.configs[chainID]
	if !exists {
		// Create default config
		config = &GasConfig{
			BaseGasPrice:    big.NewInt(20000000000), // 20 Gwei
			GasMultiplier:   1.1,                      // 10% buffer
			MinGasRequired:  big.NewInt(1000000000),  // 1 FLU
			MaxGasAllowed:   big.NewInt(10000000000), // 10 FLU
			UpdateInterval:  5 * time.Minute,
			LastUpdate:      time.Now(),
		}
		ga.configs[chainID] = config
	}

	// Update config if needed
	if time.Since(config.LastUpdate) > config.UpdateInterval {
		err := ga.updateGasConfig(chainID, config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

// updateGasConfig updates the gas configuration for a chain
func (ga *GasAbstraction) updateGasConfig(chainID string, config *GasConfig) error {
	// Get current market conditions
	priceFeed, exists := ga.priceFeeds[chainID]
	if !exists {
		return errors.New("price feed not found for chain")
	}

	// Update base gas price
	gasPrice, err := priceFeed.GetGasPrice(context.Background())
	if err != nil {
		return err
	}
	config.BaseGasPrice = gasPrice

	// Update multiplier based on network congestion
	congestion, err := priceFeed.GetNetworkCongestion(context.Background())
	if err != nil {
		return err
	}
	config.GasMultiplier = 1.0 + (congestion * 0.2) // Up to 20% increase based on congestion

	config.LastUpdate = time.Now()
	return nil
}

// emitGasPaymentEvent emits a cross-chain gas payment event
func (ga *GasAbstraction) emitGasPaymentEvent(ctx context.Context, tx types.Tx, chainID string, gasCost *big.Int) error {
	event := types.GasPaymentEvent{
		TxHash:      tx.Hash(),
		ChainID:     chainID,
		Sender:      tx.Sender(),
		GasCost:     gasCost,
		Timestamp:   time.Now().Unix(),
		BlockHeight: tx.Height(),
	}

	return ga.keeper.EmitEvent(ctx, "gas_payment", event)
} 