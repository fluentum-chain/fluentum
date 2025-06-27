package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Find the CometBFT secp256k1 file in the module cache
	goModCache := os.Getenv("GOMODCACHE")
	if goModCache == "" {
		homeDir, _ := os.UserHomeDir()
		goModCache = filepath.Join(homeDir, "go", "pkg", "mod")
	}

	cometbftPath := filepath.Join(goModCache, "github.com", "cometbft", "cometbft@v0.38.0", "crypto", "secp256k1", "secp256k1.go")

	if _, err := os.Stat(cometbftPath); os.IsNotExist(err) {
		fmt.Println("CometBFT file not found at:", cometbftPath)
		fmt.Println("Trying alternative paths...")

		// Try to find the file in different possible locations
		possiblePaths := []string{
			filepath.Join(goModCache, "github.com", "cometbft", "cometbft@v0.38.0"),
			filepath.Join(goModCache, "github.com", "cometbft", "cometbft@v0.37.0"),
			filepath.Join(goModCache, "github.com", "cometbft", "cometbft@v0.36.0"),
		}

		for _, basePath := range possiblePaths {
			secpPath := filepath.Join(basePath, "crypto", "secp256k1", "secp256k1.go")
			if _, err := os.Stat(secpPath); err == nil {
				cometbftPath = secpPath
				fmt.Println("Found CometBFT file at:", cometbftPath)
				break
			}
		}
	}

	// Read the file
	content, err := os.ReadFile(cometbftPath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", cometbftPath, err)
		return
	}

	lines := strings.Split(string(content), "\n")

	// Find and fix the problematic line
	fixed := false
	for i, line := range lines {
		if strings.Contains(line, "compactSig, err := ecdsa.SignCompact") && strings.Contains(line, "2 variables") {
			// Fix the assignment mismatch
			lines[i] = strings.Replace(line, "compactSig, err := ecdsa.SignCompact", "compactSig := ecdsa.SignCompact", 1)
			fixed = true
			fmt.Printf("Fixed line %d: %s\n", i+1, lines[i])
		}
	}

	if !fixed {
		fmt.Println("No problematic lines found to fix")
		return
	}

	// Write the fixed content back
	err = os.WriteFile(cometbftPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Printf("Error writing fixed file: %v\n", err)
		return
	}

	fmt.Println("Successfully fixed CometBFT secp256k1 compilation error")
}
