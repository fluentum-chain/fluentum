package app

import (
	"cosmossdk.io/client"
	"cosmossdk.io/codec"
	codectypes "cosmossdk.io/codec/types"
	sdk "cosmossdk.io/types"
	"cosmossdk.io/x/auth/tx"

	fluentumtypes "github.com/fluentum-chain/fluentum/fluentum/x/fluentum/types"
)

// RegisterLegacyAminoCodec registers Amino codec types
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&fluentumtypes.MsgCreateFluentum{}, "fluentum/CreateFluentum", nil)
	cdc.RegisterConcrete(&fluentumtypes.MsgUpdateFluentum{}, "fluentum/UpdateFluentum", nil)
	cdc.RegisterConcrete(&fluentumtypes.MsgDeleteFluentum{}, "fluentum/DeleteFluentum", nil)
}

// RegisterInterfaces registers the x/auth interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&fluentumtypes.MsgCreateFluentum{},
		&fluentumtypes.MsgUpdateFluentum{},
		&fluentumtypes.MsgDeleteFluentum{},
	)
}

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates an EncodingConfig for testing
// This follows the recommended pattern for Cosmos SDK v0.50.6
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(cdc, tx.DefaultSignModes)

	encodingConfig := EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         cdc,
		TxConfig:          txCfg,
		Amino:             amino,
	}

	RegisterLegacyAminoCodec(encodingConfig.Amino)
	RegisterInterfaces(encodingConfig.InterfaceRegistry)

	return encodingConfig
}

// MakeTestEncodingConfig creates an EncodingConfig for testing
// This is the recommended pattern for Cosmos SDK v0.50.6 testing
func MakeTestEncodingConfig() EncodingConfig {
	// For now, we'll use the same implementation as MakeEncodingConfig
	// In a full implementation with cosmos-sdk/testutil, this would use:
	// encCfg := moduletestutil.MakeTestEncodingConfig(ModuleBasics)
	return MakeEncodingConfig()
}
