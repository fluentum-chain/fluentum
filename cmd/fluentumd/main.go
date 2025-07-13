package main

import (
	"context"
	"os"

	"github.com/fluentum-chain/fluentum/features"
	"github.com/fluentum-chain/fluentum/cmd/fluentumd/commands"
	"github.com/fluentum-chain/fluentum/libs/log"
	"github.com/spf13/cobra"
)

func main() {
	// Initialize logger
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// Initialize feature manager
	featureManager := features.NewFeatureManager(logger)

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "fluentumd",
		Short: "Fluentum blockchain node",
		Long: `Fluentum is a high-performance blockchain with AI-powered validation
and quantum-resistant security features.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Add feature manager to context
			ctx := context.WithValue(cmd.Context(), "featureManager", featureManager)
			cmd.SetContext(ctx)
			return nil
		},
	}

	// Register subcommands
	commands.RegisterFeatureCommands(rootCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
