//go:build !ignore_interface_check

package main

import (
	"io"

	dbm "github.com/cometbft/cometbft-db"
	tmlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/server/types"

	"github.com/fluentum-chain/fluentum/fluentum/app"
)

// Interface compliance verification for Cosmos SDK v0.50.6
// This file ensures that appCreator properly implements the required interfaces

// Verify AppCreator interface compliance
var _ types.AppCreator = (*appCreator)(nil)

// Verify AppExporter interface compliance
var _ types.AppExporter = (*appCreator)(nil)

// Test function to verify interface compliance at compile time
func verifyInterfaceCompliance() {
	// Create a test appCreator instance
	encCfg := app.MakeEncodingConfig()
	creator := appCreator{encCfg: encCfg}

	// Test AppCreator interface methods
	var appCreator types.AppCreator = creator

	// Test CreateApp method signature
	_ = func() types.Application {
		return appCreator.CreateApp(
			tmlog.NewNopLogger(),
			dbm.NewMemDB(),
			io.Discard,
			types.AppOptions{},
		)
	}

	// Test AppExporter interface methods
	var appExporter types.AppExporter = creator

	// Test ExportApp method signature
	_ = func() (types.ExportedApp, error) {
		return appExporter.ExportApp(
			tmlog.NewNopLogger(),
			dbm.NewMemDB(),
			io.Discard,
			0,          // height
			false,      // forZeroHeight
			[]string{}, // jailAllowedAddrs
			types.AppOptions{},
		)
	}

	// Test AppBlockHeight method signature
	_ = func() (int64, error) {
		return appExporter.AppBlockHeight()
	}
}

// This function is never called but ensures compile-time interface checking
func init() {
	verifyInterfaceCompliance()
}
