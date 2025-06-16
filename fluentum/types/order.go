package types

import "time"

// OrderType represents the type of order (market or limit)
type OrderType int

const (
	// MarketOrder is an order that executes immediately at the best available price
	MarketOrder OrderType = iota
	// LimitOrder is an order that executes at a specified price or better
	LimitOrder
)

// Order represents a trading order
type Order struct {
	ID        string    `json:"id"`
	Type      OrderType `json:"type"`
	Amount    int64     `json:"amount"`     // Amount in base units (e.g., satoshis)
	Price     int64     `json:"price"`      // Price in base units
	Side      string    `json:"side"`       // "buy" or "sell"
	Timestamp time.Time `json:"timestamp"`
	// Additional fields for order routing
	MaxSlippage int64  `json:"max_slippage"` // Maximum allowed slippage in basis points
	RouteHint   string `json:"route_hint"`   // Optional hint for preferred route
} 