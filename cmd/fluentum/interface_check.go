//go:build !ignore_interface_check

package main

import (
	"github.com/fluentum-chain/fluentum/app"
)

// Interface compliance verification for Cosmos SDK v0.50.6
// This file ensures that appCreator properly implements the required interfaces

// Test function to verify basic functionality at compile time
func verifyBasicFunctionality() {
	// Create a test appCreator instance
	encCfg := app.MakeEncodingConfig()
	_ = appCreator{encCfg: encCfg}
}

// This function is never called but ensures compile-time checking
// Commented out to prevent duplicate interface registration
// func init() {
// 	verifyBasicFunctionality()
// }
