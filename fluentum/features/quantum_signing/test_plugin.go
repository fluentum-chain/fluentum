package main

import (
    "log"
    "github.com/fluentum-chain/fluentum/fluentum/core/plugin"
)

func main() {
    config := plugin.PluginConfig{
        SecurityLevel: "Dilithium3",
    }

    signer, err := plugin.Instance().LoadSignerWithConfig(config)
    if err != nil {
        log.Fatalf("Failed to load quantum signer plugin: %v", err)
    }
    log.Printf("Quantum signer plugin loaded: %+v\n", signer)
}
