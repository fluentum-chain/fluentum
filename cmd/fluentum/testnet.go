package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type TestnetConfig struct {
	NumValidators int
	OutputDir     string
	ChainID       string
}

func runTestnet(cmd *cobra.Command, args []string) error {
	// Parse flags
	config := &TestnetConfig{}
	flag.IntVar(&config.NumValidators, "v", 4, "Number of validators")
	flag.StringVar(&config.OutputDir, "o", "./testnet", "Output directory")
	flag.StringVar(&config.ChainID, "chain-id", "test-chain", "Chain ID")
	flag.Parse()

	// Create output directory
	err := os.MkdirAll(config.OutputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate validator keys and configs
	for i := 0; i < config.NumValidators; i++ {
		nodeDir := filepath.Join(config.OutputDir, fmt.Sprintf("node%d", i))
		err := generateNodeConfig(nodeDir, i, config)
		if err != nil {
			return fmt.Errorf("failed to generate config for node %d: %w", i, err)
		}
	}

	fmt.Printf("Generated testnet with %d validators in %s\n", config.NumValidators, config.OutputDir)
	return nil
}

func generateNodeConfig(nodeDir string, index int, config *TestnetConfig) error {
	// Create node directory
	err := os.MkdirAll(filepath.Join(nodeDir, "config"), 0755)
	if err != nil {
		return err
	}

	// Generate simple node key (placeholder)
	nodeKeyData := fmt.Sprintf(`{
		"priv_key": {
			"type": "tendermint/PrivKeyEd25519",
			"value": "node_key_%d_placeholder"
		}
	}`, index)

	err = os.WriteFile(filepath.Join(nodeDir, "config", "node_key.json"), []byte(nodeKeyData), 0644)
	if err != nil {
		return err
	}

	// Generate simple validator key (placeholder)
	valKeyData := fmt.Sprintf(`{
		"address": "validator_%d_address",
		"pub_key": {
			"type": "tendermint/PubKeyEd25519",
			"value": "validator_%d_pubkey_placeholder"
		},
		"priv_key": {
			"type": "tendermint/PrivKeyEd25519",
			"value": "validator_%d_privkey_placeholder"
		}
	}`, index, index, index)

	err = os.WriteFile(filepath.Join(nodeDir, "config", "priv_validator_key.json"), []byte(valKeyData), 0644)
	if err != nil {
		return err
	}

	// Create simple node config
	nodeConfigData := fmt.Sprintf(`# Fluentum Node Configuration
chain_id = "%s"
moniker = "node%d"

[p2p]
laddr = "tcp://0.0.0.0:%d"
persistent_peers = "%s"

[rpc]
laddr = "tcp://0.0.0.0:%d"

[consensus]
timeout_commit = "1s"
timeout_propose = "1s"
`, config.ChainID, index, 26656+index, generatePersistentPeers(config.NumValidators, index), 26657+index)

	err = os.WriteFile(filepath.Join(nodeDir, "config", "config.toml"), []byte(nodeConfigData), 0644)
	if err != nil {
		return err
	}

	// Create simple genesis
	genesisData := fmt.Sprintf(`{
		"genesis_time": "%s",
		"chain_id": "%s",
		"consensus_params": {
			"block": {
				"max_bytes": "22020096",
				"max_gas": "-1",
				"time_iota_ms": "1000"
			},
			"evidence": {
				"max_age_num_blocks": "100000",
				"max_age_duration": "172800000000000",
				"max_bytes": "1048576"
			},
			"validator": {
				"pub_key_types": ["ed25519"]
			}
		},
		"validators": [
			{
				"address": "validator_%d_address",
				"pub_key": {
					"type": "tendermint/PubKeyEd25519",
					"value": "validator_%d_pubkey_placeholder"
				},
				"power": "100",
				"name": "Genesis Validator %d"
			}
		],
		"app_hash": "",
		"app_state": {
			"flumx_token": {
				"denom": "aflumx",
				"initial_supply": "1000000000000000000"
			}
		}
	}`, time.Now().UTC().Format(time.RFC3339), config.ChainID, index, index, index+1)

	err = os.WriteFile(filepath.Join(nodeDir, "config", "genesis.json"), []byte(genesisData), 0644)
	if err != nil {
		return err
	}

	return nil
}

func generatePersistentPeers(numValidators, currentIndex int) string {
	var peers []string
	for i := 0; i < numValidators; i++ {
		if i == currentIndex {
			continue
		}
		peers = append(peers, fmt.Sprintf("node%d@127.0.0.1:%d", i, 26656+i))
	}
	return strings.Join(peers, ",")
}
