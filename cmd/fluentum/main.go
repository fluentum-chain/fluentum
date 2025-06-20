package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fluentum-chain/fluentum/fluentum/abci/myapp"

	"github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/libs/log"
	"github.com/fluentum-chain/fluentum/node"
	"github.com/fluentum-chain/fluentum/proxy"
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

	// Configuration variables
	homeDir string
	p2pAddr string
	rpcAddr string
)

func init() {
	// Add version command
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Fluentum Core v0.1.0\n")
		},
	})

	// Add init command
	RootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize a new Fluentum node",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Initializing Fluentum node...")
			// TODO: Implement node initialization
			fmt.Println("Node initialized successfully!")
		},
	})

	// Add testnet command
	RootCmd.AddCommand(&cobra.Command{
		Use:   "testnet",
		Short: "Generate a testnet configuration",
		RunE:  runTestnet,
	})

	// Add configuration flags
	RootCmd.PersistentFlags().StringVar(&homeDir, "home", ".fluentum", "directory for config and data")
	RootCmd.PersistentFlags().StringVar(&p2pAddr, "p2p.laddr", "tcp://0.0.0.0:26656", "node listen address")
	RootCmd.PersistentFlags().StringVar(&rpcAddr, "rpc.laddr", "tcp://0.0.0.0:26657", "RPC listen address")
}

func runNode(cmd *cobra.Command, args []string) error {
	fmt.Printf("Starting Fluentum Core node...\n")
	fmt.Printf("Home directory: %s\n", homeDir)
	fmt.Printf("P2P address: %s\n", p2pAddr)
	fmt.Printf("RPC address: %s\n", rpcAddr)

	// 1. Load Tendermint config
	cfg := config.DefaultConfig()
	cfg.SetRoot(homeDir)

	// 2. Create the custom ABCI application
	app := myapp.NewApplication()

	// 3. Create Tendermint node
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	n, err := node.NewNode(
		cfg,
		nil, // privValidator: nil uses default file-based
		nil, // nodeKey: nil uses default file-based
		proxy.NewLocalClientCreator(app),
		node.DefaultGenesisDocProviderFunc(cfg),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(cfg.Instrumentation),
		logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	// 4. Start node
	if err := n.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}
	defer n.Stop()

	fmt.Println("Node started successfully! Press Ctrl+C to stop.")

	// 5. Wait for signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down node...")
	return nil
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
