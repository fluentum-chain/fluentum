# Quantum-Safe Signing Implementation

This package provides a quantum-resistant digital signature implementation using the CRYSTALS-Dilithium algorithm. It's designed to be used in the Fluentum blockchain for post-quantum cryptographic security.

## Features

- Quantum-resistant digital signatures using CRYSTALS-Dilithium
- Multiple security levels (Dilithium2, Dilithium3, Dilithium5)
- Thread-safe implementation
- Performance metrics tracking
- Simple and clean API
- Comprehensive test coverage

## Security Levels

- **Dilithium2**: 128-bit security (smallest key and signature sizes)
- **Dilithium3**: 192-bit security (recommended default)
- **Dilithium5**: 256-bit security (highest security level)

## Installation

```bash
go get github.com/fluentum/quantum_signing
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"log"

	"quantum_signing"
)

func main() {
	// Create a new signer instance
	signer, err := quantum_signing.NewDilithiumSigner()
	if err != nil {
		log.Fatal(err)
	}

	// Generate a key pair
	publicKey, privateKey, err := signer.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Sign a message
	message := []byte("Hello, quantum world!")
	signature, err := signer.Sign(privateKey, message)
	if err != nil {
		log.Fatal(err)
	}

	// Verify the signature
	valid, err := signer.Verify(publicKey, message, signature)
	if err != nil {
		log.Fatal(err)
	}

	if valid {
		fmt.Println("Signature is valid!")
	} else {
		fmt.Println("Signature is invalid!")
	}
}
```

### Performance Metrics

The package includes built-in performance tracking:

```go
// Create a performance metrics tracker
metrics := quantum_signing.NewPerformanceMetrics(100) // Keep last 100 samples

// Record operation times
start := time.Now()
// ... perform operation ...
metrics.RecordSign(time.Since(start))

// Get performance statistics
stats := metrics.GetStats()
fmt.Println(stats)
```

## Example

See the `example.go` file for a complete example demonstrating key generation, signing, and verification.

## Running Tests

```bash
go test -v
```

## Performance

Performance will vary based on the security level and hardware. Typical performance characteristics:

- **Key Generation**: ~10-100ms
- **Signing**: ~1-10ms
- **Verification**: ~0.1-1ms

## Security Considerations

- Always use cryptographically secure random number generators
- Protect private keys with appropriate access controls
- Regularly rotate keys based on your security requirements
- Keep the underlying `github.com/cloudflare/circl` package updated

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
