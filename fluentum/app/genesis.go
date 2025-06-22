package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	fluentumtypes "github.com/fluentum-chain/fluentum/fluentum/x/fluentum/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GenesisState represents the genesis state of the blockchain
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}

// ValidateGenesis performs genesis state validation for the given app.
func ValidateGenesis(gs GenesisState, cdc codec.JSONCodec, txConfig client.TxEncodingConfig) error {
	return ModuleBasics.ValidateGenesis(cdc, txConfig, gs)
}

// InitGenesis performs genesis initialization for the app. It returns
// the validator updates (by bond) and the next validator set.
func (app *App) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs GenesisState) []abci.ValidatorUpdate {
	// Initialize params keeper and module subspaces
	app.ParamsKeeper.Subspace(authtypes.ModuleName)
	app.ParamsKeeper.Subspace(banktypes.ModuleName)
	app.ParamsKeeper.Subspace(fluentumtypes.ModuleName)

	// Initialize modules
	app.mm.InitGenesis(ctx, cdc, gs)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the app.
func (app *App) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) GenesisState {
	gs, _ := app.mm.ExportGenesis(ctx, cdc)
	return gs
}
