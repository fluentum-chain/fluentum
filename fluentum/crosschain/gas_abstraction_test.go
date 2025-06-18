package crosschain

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/fluentum-chain/fluentum/types"
)

type mockKeeper struct {
	balances map[string]*big.Int
	events   []types.GasPaymentEvent
}

func newMockKeeper() *mockKeeper {
	return &mockKeeper{
		balances: make(map[string]*big.Int),
		events:   make([]types.GasPaymentEvent, 0),
	}
}

func (mk *mockKeeper) GetBalance(ctx context.Context, address string) *big.Int {
	if balance, exists := mk.balances[address]; exists {
		return balance
	}
	return big.NewInt(0)
}

func (mk *mockKeeper) BurnTokens(ctx context.Context, address string, amount *big.Int) error {
	balance := mk.GetBalance(ctx, address)
	if balance.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}
	mk.balances[address] = new(big.Int).Sub(balance, amount)
	return nil
}

func (mk *mockKeeper) MintTokens(ctx context.Context, address string, amount *big.Int) error {
	balance := mk.GetBalance(ctx, address)
	mk.balances[address] = new(big.Int).Add(balance, amount)
	return nil
}

func (mk *mockKeeper) EmitEvent(ctx context.Context, eventType string, event interface{}) error {
	if eventType == "gas_payment" {
		mk.events = append(mk.events, event.(types.GasPaymentEvent))
	}
	return nil
}

type mockPriceFeed struct {
	gasPrice    *big.Int
	congestion  float64
	shouldError bool
}

func newMockPriceFeed() *mockPriceFeed {
	return &mockPriceFeed{
		gasPrice:   big.NewInt(20000000000), // 20 Gwei
		congestion: 0.5,
	}
}

func (mpf *mockPriceFeed) GetGasPrice(ctx context.Context) (*big.Int, error) {
	if mpf.shouldError {
		return nil, errors.New("mock error")
	}
	return mpf.gasPrice, nil
}

func (mpf *mockPriceFeed) GetNetworkCongestion(ctx context.Context) (float64, error) {
	if mpf.shouldError {
		return 0, errors.New("mock error")
	}
	return mpf.congestion, nil
}

func TestGasAbstraction(t *testing.T) {
	keeper := newMockKeeper()
	priceFeed := newMockPriceFeed()

	ga := NewGasAbstraction(keeper)
	ga.priceFeeds["ethereum"] = priceFeed

	// Test cases
	tests := []struct {
		name           string
		sender         string
		initialBalance *big.Int
		gasLimit       int64
		shouldSucceed  bool
	}{
		{
			name:           "sufficient balance",
			sender:         "0x123",
			initialBalance: big.NewInt(10000000000), // 100 FLU
			gasLimit:       21000,
			shouldSucceed:  true,
		},
		{
			name:           "insufficient balance",
			sender:         "0x456",
			initialBalance: big.NewInt(100000000), // 1 FLU
			gasLimit:       21000,
			shouldSucceed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initial balance
			keeper.balances[tt.sender] = tt.initialBalance

			// Create test transaction
			tx := types.Tx{
				Sender:   tt.sender,
				GasLimit: tt.gasLimit,
			}

			// Attempt gas payment
			err := ga.PayGasWithFLU(context.Background(), tx, "ethereum")

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("expected success, got error: %v", err)
				}

				// Verify balance was reduced
				finalBalance := keeper.GetBalance(context.Background(), tt.sender)
				if finalBalance.Cmp(tt.initialBalance) >= 0 {
					t.Error("balance was not reduced")
				}

				// Verify event was emitted
				if len(keeper.events) == 0 {
					t.Error("no gas payment event was emitted")
				}
			} else {
				if err == nil {
					t.Error("expected error, got success")
				}

				// Verify balance was not changed
				finalBalance := keeper.GetBalance(context.Background(), tt.sender)
				if finalBalance.Cmp(tt.initialBalance) != 0 {
					t.Error("balance was changed when it shouldn't have been")
				}
			}
		})
	}
}

func TestGasConfigUpdates(t *testing.T) {
	keeper := newMockKeeper()
	priceFeed := newMockPriceFeed()

	ga := NewGasAbstraction(keeper)
	ga.priceFeeds["ethereum"] = priceFeed

	// Get initial config
	config, err := ga.getGasConfig("ethereum")
	if err != nil {
		t.Fatalf("failed to get gas config: %v", err)
	}

	// Verify initial values
	if config.BaseGasPrice.Cmp(big.NewInt(20000000000)) != 0 {
		t.Error("unexpected initial base gas price")
	}
	if config.GasMultiplier != 1.1 {
		t.Error("unexpected initial gas multiplier")
	}

	// Update price feed values
	priceFeed.gasPrice = big.NewInt(30000000000) // 30 Gwei
	priceFeed.congestion = 0.8

	// Force config update
	config.LastUpdate = time.Now().Add(-6 * time.Minute)

	// Get updated config
	config, err = ga.getGasConfig("ethereum")
	if err != nil {
		t.Fatalf("failed to get updated gas config: %v", err)
	}

	// Verify updated values
	if config.BaseGasPrice.Cmp(big.NewInt(30000000000)) != 0 {
		t.Error("base gas price was not updated")
	}
	if config.GasMultiplier != 1.16 { // 1.0 + (0.8 * 0.2)
		t.Error("gas multiplier was not updated correctly")
	}
}
