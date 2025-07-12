// Package quantum_signing provides quantum-resistant digital signature functionality
// for the Fluentum blockchain using the CRYSTALS-Dilithium algorithm.
//
// This package implements post-quantum cryptographic signatures to protect against
// attacks from quantum computers. It provides both a direct API for cryptographic
// operations and a feature implementation that can be integrated with the Fluentum
// node's feature management system.
//
// # Features
//
//   - Quantum-resistant signatures using CRYSTALS-Dilithium
//   - Multiple security levels (Dilithium2, Dilithium3, Dilithium5)
//   - Integration with Fluentum's feature management system
//   - Thread-safe implementation
//   - Performance metrics and monitoring
//
// # Usage
//
// To use the quantum signing feature in a Fluentum node:
//
//  1. Import the package:
//     import "github.com/fluentum-chain/fluentum/features/quantum_signing"
//
//  2. Register the feature with the feature manager:
//     quantum_signing.Register(featureManager)
//
//  3. The feature can then be managed through the feature manager's API
//
// # Security Levels
//
// The package supports three security levels:
//
//   - Dilithium2: 128-bit security (smallest key and signature sizes)
//   - Dilithium3: 192-bit security (recommended default)
//   - Dilithium5: 256-bit security (highest security level)
//
// # Example
//
//  // Create a new quantum signer
//  signer, err := quantum_signing.NewDilithiumSigner()
//  if err != nil {
//      log.Fatal(err)
//  }
//
//  // Generate a key pair
//  publicKey, privateKey, err := signer.GenerateKey()
//  if err != nil {
//      log.Fatal(err)
//  }
//
//  // Sign a message
//  message := []byte("Hello, quantum world!")
//  signature, err := signer.Sign(privateKey, message)
//  if err != nil {
//      log.Fatal(err)
//  }
//
//  // Verify the signature
//  valid, err := signer.Verify(publicKey, message, signature)
//  if err != nil {
//      log.Fatal(err)
//  }
//  if valid {
//      fmt.Println("Signature is valid!")
//  }
package quantum_signing
