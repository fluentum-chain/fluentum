package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fluentum-chain/fluentum/features"
	"github.com/spf13/cobra"
)

// FeatureCommands returns the feature management commands
func FeatureCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feature",
		Short: "Manage Fluentum features",
	}

	cmd.AddCommand(
		featureListCommand(),
		featureInstallCommand(),
		featureUninstallCommand(),
		featureEnableCommand(),
		featureDisableCommand(),
		featureUpdateCommand(),
	)

	return cmd
}

func featureListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all installed features",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the feature manager from the command context
			fm, ok := cmd.Context().Value("featureManager").(*features.FeatureManager)
			if !ok {
				return fmt.Errorf("failed to get feature manager from context")
			}

			// List features
			featureList, err := fm.ListFeatures()
			if err != nil {
				return fmt.Errorf("failed to list features: %w", err)
			}

			// Print features
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tVERSION\tENABLED\tSTATUS")

			for _, f := range featureList {
				status := "Not loaded"
				if f.Plugin != nil {
					status = "Loaded"
				}
				fmt.Fprintf(w, "%s\t%s\t%v\t%s\n",
					f.Name,
					f.Version,
					f.Enabled,
					status,
				)
			}

			return w.Flush()
		},
	}
}

func featureInstallCommand() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "install <name> [version]",
		Short: "Install a feature",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			featureName := args[0]
			if version == "" {
				version = "latest"
			}

			// Get the feature manager from the command context
			fm, ok := cmd.Context().Value("featureManager").(*features.FeatureManager)
			if !ok {
				return fmt.Errorf("failed to get feature manager from context")
			}

			// Check if feature is already installed
			if _, exists := fm.GetFeature(featureName); exists {
				return fmt.Errorf("feature %s is already installed", featureName)
			}

			// Install the feature
			if err := fm.InstallFeature(featureName, version); err != nil {
				return fmt.Errorf("failed to install feature %s: %w", featureName, err)
			}

			fmt.Printf("Successfully installed feature: %s@%s\n", featureName, version)
			return nil
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "Version of the feature to install")

	return cmd
}

func featureUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <name>",
		Short: "Uninstall a feature",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			featureName := args[0]

			// Get the feature manager from the command context
			fm, ok := cmd.Context().Value("featureManager").(*features.FeatureManager)
			if !ok {
				return fmt.Errorf("failed to get feature manager from context")
			}

			// Check if feature exists before uninstalling
			if _, exists := fm.GetFeature(featureName); !exists {
				return fmt.Errorf("feature %s not found", featureName)
			}

			// Uninstall the feature
			if err := fm.UninstallFeature(featureName); err != nil {
				return fmt.Errorf("failed to uninstall feature %s: %w", featureName, err)
			}

			fmt.Printf("Successfully uninstalled feature: %s\n", featureName)
			return nil
		},
	}
}

func featureEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <name>",
		Short: "Enable a feature",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			featureName := args[0]
			// Get the feature manager from the command context
			fm, ok := cmd.Context().Value("featureManager").(*features.FeatureManager)
			if !ok {
				return fmt.Errorf("failed to get feature manager from context")
			}

			// Check if feature exists before enabling
			if _, exists := fm.GetFeature(featureName); !exists {
				return fmt.Errorf("feature %s not found", featureName)
			}

			// Enable the feature
			if err := fm.Enable(featureName); err != nil {
				return fmt.Errorf("failed to enable feature %s: %w", featureName, err)
			}

			fmt.Printf("Successfully enabled feature: %s\n", featureName)
			return nil
		},
	}
}

func featureDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <name>",
		Short: "Disable a feature",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			featureName := args[0]

			// Get the feature manager from the command context
			fm, ok := cmd.Context().Value("featureManager").(*features.FeatureManager)
			if !ok {
				return fmt.Errorf("failed to get feature manager from context")
			}

			// Check if feature exists before disabling
			if _, exists := fm.GetFeature(featureName); !exists {
				return fmt.Errorf("feature %s not found", featureName)
			}

			// Disable the feature
			if err := fm.Disable(featureName); err != nil {
				return fmt.Errorf("failed to disable feature %s: %w", featureName, err)
			}

			fmt.Printf("Successfully disabled feature: %s\n", featureName)
			return nil
		},
	}
}

func featureUpdateCommand() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update features",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the feature manager from the command context
			fm, ok := cmd.Context().Value("featureManager").(*features.FeatureManager)
			if !ok {
				return fmt.Errorf("failed to get feature manager from context")
			}

			if all {
				// Update all features
				featureList, err := fm.ListFeatures()
				if err != nil {
					return fmt.Errorf("failed to list features: %w", err)
				}

				for _, f := range featureList {
					if err := fm.InstallFeature(f.Name, "latest"); err != nil {
						fmt.Printf("Failed to update feature %s: %v\n", f.Name, err)
					} else {
						fmt.Printf("Successfully updated feature: %s@latest\n", f.Name)
					}
				}

				return nil
			}

			// Update a specific feature
			if len(args) == 0 {
				return fmt.Errorf("feature name is required when --all is not set")
			}

			featureName := args[0]

			// Check if feature exists before updating
			if _, exists := fm.GetFeature(featureName); !exists {
				return fmt.Errorf("feature %s not found", featureName)
			}

			// Update the feature
			if err := fm.InstallFeature(featureName, "latest"); err != nil {
				return fmt.Errorf("failed to update feature %s: %w", featureName, err)
			}

			fmt.Printf("Successfully updated feature: %s@latest\n", featureName)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Update all features")

	return cmd
}

// RegisterFeatureCommands registers the feature commands with the root command
func RegisterFeatureCommands(rootCmd *cobra.Command) {
	// Add the feature command
	featureCmd := FeatureCommands()

	// Find the existing feature command if it exists
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "feature" {
			// Remove the existing feature command
			rootCmd.RemoveCommand(cmd)
			break
		}
	}

	// Add the new feature command
	rootCmd.AddCommand(featureCmd)

	// Add feature subcommands to the help
	featureCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Default help
		cmd.Root().HelpFunc()(cmd, args)

		// Additional help text
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # List all installed features")
		fmt.Println("  fluentumd feature list")
		fmt.Println()
		fmt.Println("  # Install a feature")
		fmt.Println("  fluentumd feature install qmoe_validator")
		fmt.Println()
		fmt.Println("  # Enable a feature")
		fmt.Println("  fluentumd feature enable qmoe_validator")
		fmt.Println()
		fmt.Println("  # Update all features")
		fmt.Println("  fluentumd feature update --all")
	})

	// Add aliases for common commands
	rootCmd.AddCommand(&cobra.Command{
		Use:   "features",
		Short: "Manage Fluentum features (alias for 'feature')",
		RunE:  featureListCommand().RunE,
	})

	// Add completion for feature names
	featureCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// In a real implementation, this would return the list of available features
		return []string{"qmoe_validator", "quantum_signer"}, cobra.ShellCompDirectiveNoFileComp
	}
}
