# Fluentum Plugin System

This directory contains plugins for the Fluentum blockchain platform that can be loaded dynamically at runtime.

## Overview

The plugin system allows you to extend Fluentum's functionality without recompiling the main binary. Plugins are implemented as Go shared libraries (`.so` files) that export specific functions.

## Available Plugins

### Quantum Signer Plugin (`quantum_signer/`)

A quantum-resistant signature plugin using CRYSTALS-Dilithium.

**Features:**
- Quantum-resistant signatures using Dilithium Mode 3
- Implements the `crypto.Signer` interface
- Can be loaded and activated at runtime

**Build:**
```bash
cd quantum_signer
chmod +x build.sh
./build.sh
```

**Usage:**
```go
import "github.com/fluentum-chain/fluentum/core/plugin"

// Load the quantum signer plugin
err := plugin.LoadSignerPlugin("./quantum_signer.so")
if err != nil {
    log.Fatal("Failed to load quantum signer:", err)
}

// The quantum signer is now active
signer := crypto.GetSigner()
fmt.Println("Active signer:", signer.Name()) // Output: "dilithium"
```

## Plugin Interface

All plugins must export a function named `ExportSigner` with the signature:
```go
func ExportSigner() crypto.Signer
```

## Creating Your Own Plugin

1. **Create a new directory** in `plugins/`
2. **Implement the required interface** (e.g., `crypto.Signer`)
3. **Export the required function** (`ExportSigner`)
4. **Build as a shared library** using `go build -buildmode=plugin`

### Example Plugin Structure

```
plugins/
└── my_plugin/
    ├── go.mod
    ├── my_plugin.go
    └── build.sh
```

### Example Plugin Implementation

```go
package main

import "github.com/fluentum-chain/fluentum/core/crypto"

type MySigner struct{}

func (m *MySigner) GenerateKey() ([]byte, []byte) {
    // Implementation
}

func (m *MySigner) Sign(privateKey []byte, message []byte) []byte {
    // Implementation
}

func (m *MySigner) Verify(publicKey []byte, message []byte, signature []byte) bool {
    // Implementation
}

func (m *MySigner) Name() string {
    return "my_signer"
}

// This function must be exported
func ExportSigner() crypto.Signer {
    return &MySigner{}
}
```

## Integration with Feature System

Plugins can be integrated with the modular feature system:

1. **Load the plugin** in your feature's `Init()` method
2. **Register the signer** with the crypto system
3. **Activate the signer** when the feature is enabled

Example in a feature:
```go
func (f *MyFeature) Init(config map[string]interface{}) error {
    // Load plugin
    err := plugin.LoadSignerPlugin("./my_plugin.so")
    if err != nil {
        return err
    }
    
    // Plugin is now active
    return nil
}
```

## Security Considerations

- **Verify plugin signatures** before loading in production
- **Use trusted plugin sources** only
- **Validate plugin behavior** after loading
- **Consider plugin isolation** for security-critical operations

## Troubleshooting

### Common Issues

1. **Plugin not found**: Ensure the `.so` file exists and is readable
2. **Symbol not found**: Verify `ExportSigner` function is exported
3. **Type mismatch**: Ensure the exported function returns `crypto.Signer`
4. **Dependency issues**: Make sure all dependencies are available

### Debug Tips

- Check plugin file permissions
- Verify Go version compatibility
- Test plugin loading in isolation
- Check for missing dependencies 