package main

import (
	"fmt"
	"os"

	"github.com/kellyadamtan/tendermint/config"
	"github.com/kellyadamtan/tendermint/node"
	"github.com/kellyadamtan/tendermint/version"
	"github.com/spf13/cobra"
)

var (
	// RootCmd is the root command for Fluentum Core
	RootCmd = &cobra.Command{
		Use:   "fluentum",
		Short: "Fluentum Core - A hybrid consensus blockchain",
		Long: `Fluentum Core is a blockchain platform that combines DPoS and ZK-Rollups
for high throughput and security. It features:
- Hybrid consensus mechanism
- Zero-knowledge proofs
- Quantum-resistant signatures
- Cross-chain gas abstraction
- Hybrid liquidity routing`,
		RunE: runNode,
	}

	// Config holds the node configuration
	config = config.DefaultConfig()
)

func init() {
	// Add version command
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Fluentum Core %s\n", version.Version)
		},
	})

	// Add configuration flags
	RootCmd.PersistentFlags().StringVar(&config.RootDir, "home", config.RootDir, "directory for config and data")
	RootCmd.PersistentFlags().StringVar(&config.P2P.ListenAddress, "p2p.laddr", config.P2P.ListenAddress, "node listen address")
	RootCmd.PersistentFlags().StringVar(&config.RPC.ListenAddress, "rpc.laddr", config.RPC.ListenAddress, "RPC listen address")
	RootCmd.PersistentFlags().StringVar(&config.Consensus.CreateEmptyBlocksInterval, "consensus.create_empty_blocks_interval", config.Consensus.CreateEmptyBlocksInterval, "interval between empty blocks")
}

func runNode(cmd *cobra.Command, args []string) error {
	// Create and start node
	n, err := node.NewNode(config)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	if err := n.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	// Wait for interrupt signal
	<-cmd.Context().Done()

	// Stop node gracefully
	if err := n.Stop(); err != nil {
		return fmt.Errorf("failed to stop node: %w", err)
	}

	return nil
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 