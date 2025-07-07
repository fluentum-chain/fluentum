package main

import (
    "os"
    "github.com/fluentum-chain/fluentum/config"
)

func main() {
    fmt.Fprintln(os.Stdout, "DEBUG: entered main function")
    // Load main config using defaults
    cfg := config.DefaultConfig()
    cfg.Quantum = config.DefaultQuantumConfig()
    // Continue with the rest of your main function...
}