package core

import (
	"fmt"

	cfg "github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/core/crypto"
	"github.com/fluentum-chain/fluentum/core/plugin"
	ctypes "github.com/fluentum-chain/fluentum/rpc/core/types"
	rpcserver "github.com/fluentum-chain/fluentum/rpc/jsonrpc/server"
	rpctypes "github.com/fluentum-chain/fluentum/rpc/jsonrpc/types"
)

// QuantumAPI provides RPC methods for quantum cryptography operations
type QuantumAPI struct{}

// Reload reloads the quantum signer plugin
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/quantum_reload
func (api *QuantumAPI) Reload(ctx *rpctypes.Context) (*ctypes.ResultQuantumReload, error) {
	// Get current quantum config from environment
	if env == nil || env.Config.ListenAddress == "" {
		return nil, fmt.Errorf("RPC environment not properly configured")
	}

	// For now, we'll use a default config since the full config isn't in env
	// In a real implementation, you'd want to pass the full config to the environment
	quantumConfig := cfg.DefaultQuantumConfig()

	// Unload current signer by setting to default ECDSA
	crypto.SetActiveSigner("ecdsa")

	// Reload new quantum signer
	if err := plugin.LoadQuantumSigner(quantumConfig.LibPath); err != nil {
		// Revert to default ECDSA signer on failure
		crypto.SetActiveSigner("ecdsa")
		return &ctypes.ResultQuantumReload{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ctypes.ResultQuantumReload{
		Success: true,
		Error:   "",
	}, nil
}

// Status returns the current quantum signing status
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/quantum_status
func (api *QuantumAPI) Status(ctx *rpctypes.Context) (*ctypes.ResultQuantumStatus, error) {
	activeSigner := crypto.GetSigner()
	if activeSigner == nil {
		return &ctypes.ResultQuantumStatus{
			Enabled:    false,
			SignerName: "none",
			Mode:       "",
		}, nil
	}

	return &ctypes.ResultQuantumStatus{
		Enabled:    activeSigner.Name() != "ecdsa",
		SignerName: activeSigner.Name(),
		Mode:       "", // Could be enhanced to return the specific quantum mode
	}, nil
}

// Provide a function map for registration
var QuantumRoutes = map[string]*rpcserver.RPCFunc{
	"quantum_reload": rpcserver.NewRPCFunc(QuantumReload, ""),
	"quantum_status": rpcserver.NewRPCFunc(QuantumStatus, ""),
}

// Route functions for the quantum API

// QuantumReload is the RPC route function for reloading quantum signers
func QuantumReload(ctx *rpctypes.Context) (*ctypes.ResultQuantumReload, error) {
	api := &QuantumAPI{}
	return api.Reload(ctx)
}

// QuantumStatus is the RPC route function for getting quantum status
func QuantumStatus(ctx *rpctypes.Context) (*ctypes.ResultQuantumStatus, error) {
	api := &QuantumAPI{}
	return api.Status(ctx)
}
