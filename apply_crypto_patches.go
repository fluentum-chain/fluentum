package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Use the provided path for CometBFT v0.37.1
	cometbftBasePath := filepath.FromSlash("/home/ktang/go/pkg/mod/github.com/cometbft/cometbft@v0.37.1")

	// Apply secp256k1 patch
	secp256k1Path := filepath.Join(cometbftBasePath, "crypto", "secp256k1", "secp256k1.go")
	if err := applySecp256k1Patch(secp256k1Path); err != nil {
		fmt.Printf("Error applying secp256k1 patch: %v\n", err)
	} else {
		fmt.Println("Successfully applied secp256k1 patch")
	}

	// Apply sr25519 patch
	sr25519Path := filepath.Join(cometbftBasePath, "crypto", "sr25519", "pubkey.go")
	if err := applySr25519Patch(sr25519Path); err != nil {
		fmt.Printf("Error applying sr25519 patch: %v\n", err)
	} else {
		fmt.Println("Successfully applied sr25519 patch")
	}
}

func applySecp256k1Patch(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}
	lines := strings.Split(string(content), "\n")
	fixed := false
	for i, line := range lines {
		if strings.Contains(line, "signature, err := ecdsa.SignCompact") &&
			strings.Contains(line, "crypto.Sha256(msg)") &&
			strings.Contains(line, "false") {
			lines[i] = strings.Replace(line,
				"signature, err := ecdsa.SignCompact(privKey, crypto.Sha256(msg), false)",
				"signature := ecdsa.SignCompact(privKey, crypto.Sha256(msg))", 1)
			fixed = true
			fmt.Printf("Fixed secp256k1 line %d: %s\n", i+1, lines[i])
		}
	}
	if !fixed {
		fmt.Println("No secp256k1 problematic lines found to fix")
		return nil
	}
	err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("error writing fixed file: %v", err)
	}
	return nil
}

func applySr25519Patch(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}
	lines := strings.Split(string(content), "\n")
	fixed := false
	for i, line := range lines {
		if strings.Contains(line, "return verify(pub, msg, sig), nil") {
			lines[i] = strings.Replace(line,
				"return verify(pub, msg, sig), nil",
				"return verify(pub, msg, sig)", 1)
			fixed = true
			fmt.Printf("Fixed sr25519 line %d: %s\n", i+1, lines[i])
		}
	}
	if !fixed {
		fmt.Println("No sr25519 problematic lines found to fix")
		return nil
	}
	err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("error writing fixed file: %v", err)
	}
	return nil
}
