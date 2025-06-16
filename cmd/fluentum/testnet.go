package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kellyadamtan/tendermint/config"
	"github.com/kellyadamtan/tendermint/p2p"
	"github.com/kellyadamtan/tendermint/privval"
	"github.com/kellyadamtan/tendermint/types"
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
	err := os.MkdirAll(nodeDir, 0755)
	if err != nil {
		return err
	}

	// Generate node key
	nodeKey, err := p2p.GenerateNodeKey()
	if err != nil {
		return err
	}
	err = nodeKey.SaveAs(filepath.Join(nodeDir, "config", "node_key.json"))
	if err != nil {
		return err
	}

	// Generate validator key
	valKey := privval.GenFilePV(
		filepath.Join(nodeDir, "config", "priv_validator_key.json"),
		filepath.Join(nodeDir, "config", "priv_validator_state.json"),
	)

	// Create node config
	nodeConfig := config.DefaultConfig()
	nodeConfig.SetRoot(nodeDir)
	nodeConfig.ChainID = config.ChainID
	nodeConfig.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", 26656+index)
	nodeConfig.RPC.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", 26657+index)
	nodeConfig.P2P.PersistentPeers = generatePersistentPeers(config.NumValidators, index)

	// Save config
	err = nodeConfig.SaveAs(filepath.Join(nodeDir, "config", "config.toml"))
	if err != nil {
		return err
	}

	// Save genesis
	genesis := generateGenesis(config.ChainID, config.NumValidators)
	err = genesis.SaveAs(filepath.Join(nodeDir, "config", "genesis.json"))
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

func generateGenesis(chainID string, numValidators int) *types.GenesisDoc {
	genesis := &types.GenesisDoc{
		ChainID:     chainID,
		GenesisTime: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		ConsensusParams: &types.ConsensusParams{
			Block: types.BlockParams{
				MaxBytes:   22020096,
				MaxGas:     -1,
				TimeIotaMs: 1000,
			},
			Evidence: types.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  48 * time.Hour,
				MaxBytes:        1048576,
			},
			Validator: types.ValidatorParams{
				PubKeyTypes: []string{types.ABCIPubKeyTypeDilithium},
			},
		},
	}

	// Add Fluentum-specific consensus parameters
	genesis.ConsensusParams.FluentumParams = &types.FluentumParams{
		ZKEnabled:        true,
		QuantumEnabled:   true,
		FreeGasThreshold: 5000000000, // 50 FLU
	}

	// Add validators
	for i := 0; i < numValidators; i++ {
		nodeDir := filepath.Join("testnet", fmt.Sprintf("node%d", i))
		valKey := privval.LoadFilePV(
			filepath.Join(nodeDir, "config", "priv_validator_key.json"),
			filepath.Join(nodeDir, "config", "priv_validator_state.json"),
		)
		genesis.Validators = append(genesis.Validators, types.GenesisValidator{
			Address: valKey.GetPubKey().Address(),
			PubKey:  valKey.GetPubKey(),
			Power:   100,
			Name:    fmt.Sprintf("Genesis Validator %d", i+1),
		})
	}

	// Add app state
	genesis.AppState = map[string]interface{}{
		"flu_token": map[string]interface{}{
			"denom":          "aflu",
			"initial_supply": "1000000000000000000", // 1B FLU
		},
		"staking": map[string]interface{}{
			"params": map[string]interface{}{
				"min_stake": "50000000000", // 50,000 FLU
			},
		},
	}

	return genesis
} 