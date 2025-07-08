// Main function with fixed struct initialization
func main() {
	// Set Bech32 prefix for Fluentum chain
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("fluentum", "fluentumpub")
	config.SetBech32PrefixForValidator("fluentumvaloper", "fluentumvaloperpub")
	config.SetBech32PrefixForConsensusNode("fluentumvalcons", "fluentumvalconspub")
	config.Seal()

	fmt.Fprintln(os.Stdout, "DEBUG: entered main function")
	// Load main config - stub implementation for now
	cfg := &config.Config{
		Quantum: &config.QuantumConfig{
			Enabled: false,
			Mode:    "mode3",
			LibPath: "",
		},
	}

	// Load quantum signer first
	if err := loadQuantumSigner(cfg); err != nil {
		fmt.Println("[Quantum] Quantum load failed:", err)
	}

	// Load and start modular features
	featureConfigPath := "config/features.toml"
	nodeVersion := "v0.1.0" // TODO: dynamically set from build/version
	featureLoader := core.NewFeatureLoader(featureConfigPath, nodeVersion)

	if err := featureLoader.LoadConfiguration(); err != nil {
		fmt.Println("[FeatureLoader] Failed to load feature configuration:", err)
		os.Exit(1)
	}

	if err := featureLoader.ValidateConfiguration(); err != nil {
		fmt.Println("[FeatureLoader] Feature configuration invalid:", err)
		os.Exit(1)
	}

	if err := featureLoader.InitializeFeatures(); err != nil {
		fmt.Println("[FeatureLoader] Failed to initialize features:", err)
		os.Exit(1)
	}

	if err := featureLoader.StartFeatures(); err != nil {
		fmt.Println("[FeatureLoader] Failed to start features:", err)
		os.Exit(1)
	}

	fmt.Println("[FeatureLoader] Features loaded and started:", featureLoader.GetFeatureStatus())

	fmt.Println("DEBUG: Creating root command")
	rootCmd, _ := NewRootCmd()
	fmt.Println("DEBUG: Root command created")

	// Debug: Check what commands are in the root command
	fmt.Println("DEBUG: Root command subcommands:")
	for _, subCmd := range rootCmd.Commands() {
		fmt.Printf("  - %s: %s\n", subCmd.Use, subCmd.Short)
	}

	// Debug: Check what commands are in the query command specifically
	queryCmd := rootCmd.Commands()
	for _, cmd := range queryCmd {
		if cmd.Use == "query" {
			fmt.Println("DEBUG: Found query command, checking its subcommands:")
			for _, subCmd := range cmd.Commands() {
				fmt.Printf("  - %s: %s\n", subCmd.Use, subCmd.Short)
			}
			break
		}
	}

	// Start stats HTTP server in a goroutine
	go func() {
		http.Handle("/stats", apiKeyAuthMiddleware(http.HandlerFunc(statsHandler)))
		fmt.Println("[Stats API] Listening on :8080 for /stats endpoint (API key protected)")
		http.ListenAndServe(":8080", nil)
	}()

	fmt.Println("DEBUG: About to execute root command")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("DEBUG: Root command execution failed with error:", err)
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("DEBUG: Root command executed successfully")
}
