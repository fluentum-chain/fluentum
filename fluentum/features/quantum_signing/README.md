# Quantum Signing Feature

A quantum-resistant signing implementation for the Fluentum blockchain using CRYSTALS-Dilithium.

## Features

- Quantum-resistant digital signatures using CRYSTALS-Dilithium
- Support for multiple security levels (Mode2, Mode3, Mode5)
- Thread-safe implementation for concurrent operations
- Performance metrics and monitoring
- Easy integration with the Fluentum node

## Installation

```bash
go get github.com/fluentum/quantum_signing
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/fluentum/quantum_signing"
)

func main() {
	// Create a new quantum signer instance
	signer, err := quantum_signing.NewDilithiumSigner()
	if err != nil {
		panic(fmt.Sprintf("Failed to create signer: %v", err))
	}

	// Generate a new key pair
	pubKey, privKey, err := signer.GenerateKey()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate key pair: %v", err))
	}

	// Sign a message
	message := []byte("Hello, quantum world!")
	signature, err := signer.Sign(privKey, message)
	if err != nil {
		panic(fmt.Sprintf("Failed to sign message: %v", err))
	}

	// Verify the signature
	valid, err := signer.Verify(pubKey, message, signature)
	if err != nil {
		panic(fmt.Sprintf("Verification failed: %v", err))
	}

	fmt.Printf("Signature valid: %v\n", valid)
}
```

## Security Levels

The library supports three security levels:

- **Mode2**: 128-bit security (smallest key and signature sizes)
- **Mode3**: 192-bit security (recommended default)
- **Mode5**: 256-bit security (highest security level)

## Performance

Performance metrics can be accessed through the `PerformanceMetrics` method:

```go
metrics := signer.PerformanceMetrics()
fmt.Printf("Average sign time: %.2f ms\n", metrics["avg_sign_time_ms"])
fmt.Printf("Average verify time: %.2f ms\n", metrics["avg_verify_time_ms"])
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
