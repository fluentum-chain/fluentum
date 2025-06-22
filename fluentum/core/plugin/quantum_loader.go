package plugin

import (
	"errors"
	"plugin"
	"strings"

	"github.com/fluentum-chain/fluentum/fluentum/core/crypto"
)

const (
	// The plugin must export a function with this name that returns a crypto.Signer
	SignerExportSymbol = "ExportSigner"
)

// LoadSignerPlugin loads a signer from a Go plugin (.so file) and registers it by its Name().
// Example: path = "./quantum_signing.so"
func LoadSignerPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	sym, err := p.Lookup(SignerExportSymbol)
	if err != nil {
		return err
	}

	// The symbol should be a function: func() crypto.Signer
	exportFunc, ok := sym.(func() crypto.Signer)
	if !ok {
		return errors.New("plugin symbol ExportSigner has wrong type (must be func() crypto.Signer)")
	}

	signer := exportFunc()
	if signer == nil {
		return errors.New("plugin ExportSigner returned nil")
	}

	name := strings.ToLower(signer.Name())
	crypto.RegisterSigner(name, signer)
	crypto.SetActiveSigner(name)
	return nil
}

// LoadQuantumSigner loads a quantum signer plugin (backward compatibility)
func LoadQuantumSigner(path string) error {
	return LoadSignerPlugin(path)
}
