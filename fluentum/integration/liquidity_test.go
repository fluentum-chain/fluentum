package integration

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fluentum-chain/fluentum/liquidity"
	"github.com/fluentum-chain/fluentum/types"
	"github.com/fluentum-chain/fluentum/x/cex"
	"github.com/fluentum-chain/fluentum/x/dex"
)

// TestHybridLiquidityRouting tests the hybrid liquidity routing system
func TestHybridLiquidityRouting(t *testing.T) {
	// Setup test environment
	cfg := ResetConfig("liquidity_test")
	defer os.RemoveAll(cfg.RootDir)

	// Create CEX and DEX clients
	cexClient := createTestCEXClient(t)
	dexClient := createTestDEXClient(t)

	// Create router
	router := liquidity.NewRouter(cexClient, dexClient)
	require.NotNil(t, router)

	// Test cases
	tests := []struct {
		name          string
		order         types.Order
		expectedRoute string
	}{
		{
			name: "large order to CEX",
			order: types.Order{
				ID:        "order1",
				Type:      types.MarketOrder,
				Amount:    2000000000, // 20 FLU
				Price:     1000000000, // 10 FLU
				Side:      "buy",
				Timestamp: time.Now(),
			},
			expectedRoute: "cex",
		},
		{
			name: "small order to DEX",
			order: types.Order{
				ID:        "order2",
				Type:      types.MarketOrder,
				Amount:    500000000,  // 5 FLU
				Price:     1000000000, // 10 FLU
				Side:      "sell",
				Timestamp: time.Now(),
			},
			expectedRoute: "dex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Route order
			err := router.RouteOrder(context.Background(), tt.order)
			require.NoError(t, err)

			// Verify routing
			if tt.expectedRoute == "cex" {
				require.True(t, cexClient.HasOrder(tt.order.ID))
				require.False(t, dexClient.HasOrder(tt.order.ID))
			} else {
				require.False(t, cexClient.HasOrder(tt.order.ID))
				require.True(t, dexClient.HasOrder(tt.order.ID))
			}
		})
	}
}

// TestDynamicThreshold tests the dynamic threshold adjustment
func TestDynamicThreshold(t *testing.T) {
	cfg := ResetConfig("threshold_test")
	defer os.RemoveAll(cfg.RootDir)

	cexClient := createTestCEXClient(t)
	dexClient := createTestDEXClient(t)
	router := liquidity.NewRouter(cexClient, dexClient)

	// Set initial market conditions
	cexClient.SetLiquidity(big.NewInt(100000000000)) // 1000 FLU
	dexClient.SetLiquidity(big.NewInt(50000000000))  // 500 FLU
	cexClient.SetFees(1000)                          // 0.1%
	dexClient.SetFees(3000)                          // 0.3%

	// Wait for threshold update
	time.Sleep(6 * time.Minute)

	// Test order routing with new threshold
	order := types.Order{
		ID:        "order3",
		Type:      types.MarketOrder,
		Amount:    1500000000, // 15 FLU
		Price:     1000000000, // 10 FLU
		Side:      "buy",
		Timestamp: time.Now(),
	}

	err := router.RouteOrder(context.Background(), order)
	require.NoError(t, err)

	// Verify routing based on new threshold
	require.True(t, cexClient.HasOrder(order.ID))
	require.False(t, dexClient.HasOrder(order.ID))
}

// Helper functions

func createTestCEXClient(t *testing.T) *cex.Client {
	client := cex.NewClient("test_key", "test_secret", "http://localhost:8080")
	client.SetLiquidity(big.NewInt(100000000000)) // 1000 FLU
	client.SetFees(1000)                          // 0.1%
	return client
}

func createTestDEXClient(t *testing.T) *dex.Client {
	client := dex.NewClient("0x123", 1)
	client.SetLiquidity(big.NewInt(50000000000)) // 500 FLU
	client.SetFees(3000)                         // 0.3%
	return client
}
