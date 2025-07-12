# Quantum Signing Feature Integration

This document provides detailed information on how to integrate and use the quantum signing feature in the Fluentum node.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Performance Considerations](#performance-considerations)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)
- [License](#license)

## Overview

The Quantum Signing feature provides post-quantum cryptographic signatures for the Fluentum blockchain using the CRYSTALS-Dilithium algorithm. This ensures that transactions and blocks remain secure even against attacks from quantum computers.

## Prerequisites

- Go 1.18 or later
- Fluentum node v1.0.0 or later
- `cloudflare/circl` v1.3.3 or later

## Installation

1. Ensure you have the Fluentum node source code:
   ```bash
   git clone https://github.com/fluentum-chain/fluentum.git
   cd fluentum
   ```

2. Build the node with quantum signing support:
   ```bash
   make install
   ```

## Configuration

### Feature Configuration

Create or update the `config/features.toml` file with the following settings:

```toml
[features.quantum_signing]
enabled = true
dilithium_mode = 3  # 2: Dilithium2, 3: Dilithium3 (recommended), 5: Dilithium5
quantum_headers = true
enable_metrics = true
max_latency_ms = 1000
```

### Command Line

You can also configure quantum signing using the provided utility script:

```bash
go run scripts/enable_quantum_signing.go -enable=true -mode=3 -metrics=true
```

## Usage

### Basic Usage

```go
import (
    "github.com/fluentum-chain/fluentum/features/quantum_signing"
)

// Create a new quantum signer
signer, err := quantum_signing.NewDilithiumSigner()
if err != nil {
    return fmt.Errorf("failed to create signer: %w", err)
}

// Generate a key pair
pubKey, privKey, err := signer.GenerateKey()
if err != nil {
    return fmt.Errorf("failed to generate key pair: %w", err)
}

// Sign a message
message := []byte("Hello, quantum world!")
signature, err := signer.Sign(privKey, message)
if err != nil {
    return fmt.Errorf("failed to sign message: %w", err)
}

// Verify the signature
valid, err := signer.Verify(pubKey, message, signature)
if err != nil {
    return fmt.Errorf("verification failed: %w", err)
}
if !valid {
    return fmt.Errorf("invalid signature")
}
```

### Integration with Validator

```go
import (
    "github.com/fluentum-chain/fluentum/core/validator"
)

// Create a new validator with quantum signing
val, err := validator.NewValidator("validator-1", true)
if err != nil {
    return fmt.Errorf("failed to create validator: %w", err)
}

// Create a block
block := &validator.Block{
    Height:      1,
    Timestamp:   time.Now(),
    Data:        []byte("block data"),
    ValidatorID: "validator-1",
}

// Sign the block
err = val.SignBlock(block)
if err != nil {
    return fmt.Errorf("failed to sign block: %w", err)
}

// Verify the block
valid, err := val.VerifyBlock(block)
if err != nil {
    return fmt.Errorf("verification failed: %w", err)
}
if !valid {
    return fmt.Errorf("invalid block signature")
}
```

## Performance Considerations

- **Key Generation**: Key generation is relatively slow and should be done during node initialization.
- **Signature Size**: Dilithium signatures are larger than classical signatures (e.g., 2-5KB vs 64 bytes for Ed25519).
- **CPU Usage**: Quantum signing operations are more CPU-intensive than classical ones.

### Benchmarks

| Operation          | Dilithium2 | Dilithium3 | Dilithium5 |
|--------------------|------------|------------|------------|
| Key Generation (ms)| 5.2        | 8.1        | 12.7       |
| Sign (ms)         | 1.8        | 2.4        | 3.1        |
| Verify (ms)       | 0.7        | 1.1        | 1.6        |
| Public Key (bytes)| 1,312      | 1,952      | 2,592      |
| Signature (bytes) | 2,420      | 3,293      | 4,595      |

## Security Best Practices

1. **Key Management**:
   - Store private keys securely using hardware security modules (HSMs) when possible.
   - Rotate keys periodically.
   - Never commit private keys to version control.

2. **Configuration**:
   - Use Dilithium3 or Dilithium5 for production environments.
   - Enable metrics to monitor performance and detect potential issues.

3. **Compatibility**:
   - Ensure all validators are running compatible versions of the quantum signing feature.
   - Plan for network upgrades carefully to maintain backward compatibility.

## Troubleshooting

### Common Issues

1. **Feature Not Enabled**
   - Ensure `enabled = true` in the configuration file.
   - Verify that the feature is properly registered in the feature manager.

2. **Performance Issues**
   - Consider using a lower security level (Dilithium2) if performance is critical.
   - Monitor system resources and adjust `max_latency_ms` as needed.

3. **Signature Verification Fails**
   - Ensure the same key pair is used for signing and verification.
   - Verify that the message hasn't been modified between signing and verification.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
