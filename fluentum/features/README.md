# Fluentum Feature System

This directory contains the implementation of Fluentum's modular feature system, which allows for dynamic loading and management of features like the QMoE validator and quantum signer.

## Features

### QMoE Validator

A Quantized Mixture-of-Experts (QMoE) based validator that provides:

- Predictive transaction batching
- Dynamic quantization for reduced memory usage
- Sparse neural network activation
- Adaptive thresholding for transaction validation

### Quantum Signer

A quantum-resistant signer that provides:

- Post-quantum cryptographic signatures
- Multiple algorithm support (Dilithium, etc.)
- Key management and rotation

## Getting Started

### Prerequisites

- Go 1.19 or later
- Fluentum node
- Required dependencies (see `go.mod`)

### Building Features

Features are built as Go plugins. To build a feature:

```bash
# Build the QMoE validator
cd fluentum/features/qmoe_validator
go build -buildmode=plugin -o ../../../plugins/qmoe_validator.so

# Build the quantum signer
cd ../quantum_signer
go build -buildmode=plugin -o ../../../plugins/quantum_signer.so
```

### Configuration

Features are configured in the node's `config.toml`:

```toml
[features]
enabled = ["qmoe_validator", "quantum_signer"]
auto_update = true
update_check_interval = "24h"

[features.registry]
local_path = "$HOME/.fluentum/features"
remote_registry = "https://features.fluentum.xyz"

[features.qmoe_validator]
enabled = true
quantization = true
sparse_activation = true
num_experts = 8
confidence_threshold = 0.7
gas_savings_threshold = 0.3
model_path = "$HOME/.fluentum/models/qmoe.bin"

[features.quantum_signer]
enabled = true
key_type = "dilithium3"
key_path = "$HOME/.fluentum/keys/quantum"
signing_algorithm = "dilithium"
key_size = 2048
```

## Using the CLI

The `fluentumd` CLI includes commands for managing features:

```bash
# List installed features
fluentumd feature list

# Install a feature
fluentumd feature install qmoe_validator

# Enable a feature
fluentumd feature enable qmoe_validator

# Update all features
fluentumd feature update --all
```

## Creating a New Feature

To create a new feature:

1. Create a new directory under `features/`
2. Implement the `FeatureInterface` from `features/interface.go`
3. Export your feature as a plugin symbol named `Feature`
4. Add configuration to `config/features.go`
5. Update the build system to include your feature

Example feature structure:

```
features/
  my_feature/
    main.go          # Plugin entry point
    feature.go       # Feature implementation
    config.go        # Feature-specific configuration
    README.md        # Documentation
```

## Security Considerations

- Features run in the same process as the node, so they have full access to the node's state
- Only install features from trusted sources
- Review the code of any third-party features before enabling them
- Use the feature isolation option for untrusted features (when implemented)

## License

This code is licensed under the [Apache 2.0 License](LICENSE).
