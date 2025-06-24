package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	"github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/fluentum/app"
	"github.com/fluentum-chain/fluentum/fluentum/core"
	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
)

// NewRootCmd creates a new root command for the Fluentum application.
func NewRootCmd() (*cobra.Command, app.EncodingConfig) {
	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   "fluentumd",
		Short: "Fluentum Core - A hybrid consensus blockchain",
		Long: `Fluentum Core is a blockchain platform that combines DPoS and ZK-Rollups
for high throughput and security. It features:
- Hybrid consensus mechanism
- Zero-knowledge proofs
- Quantum-resistant signatures
- Cross-chain gas abstraction
- Hybrid liquidity routing`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = clientconfig.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd, "", nil, nil)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	return rootCmd, encodingConfig
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig app.EncodingConfig) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	// Create address codec for genutil commands
	addressCodec := app.SimpleAddressCodec{Prefix: cfg.GetBech32AccountAddrPrefix()}

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, nil, addressCodec),
		genutilcli.MigrateGenesisCmd(nil), // TODO: implement migration map
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, addressCodec),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		debug.Cmd(),
		// config.Cmd(), // Removed - not available in v0.50.6
	)

	// Create a proper start command for the Fluentum node
	startCmd := createStartCommand(encodingConfig)
	rootCmd.AddCommand(startCmd)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		// rpc.StatusCommand(), // Removed - not available in v0.50.6
		queryCommand(),
		txCommand(),
		keys.Commands(), // Removed home parameter
	)
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// authcmd.GetAccountCmd(), // Removed - not available in v0.50.6
		rpc.ValidatorCommand(),
		// rpc.BlockCommand(), // Removed - not available in v0.50.6
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	app.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		flags.LineBreak,
	)

	app.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

type appCreator struct {
	encCfg app.EncodingConfig
}

// CreateApp implements types.AppCreator interface for Cosmos SDK v0.50.6
func (a appCreator) CreateApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts types.AppOptions,
) types.Application {
	return app.NewFluentumApp(
		logger,
		db,
		traceStore,
		true, // loadLatest
		appOpts,
		a.encCfg,
	)
}

// ExportApp implements types.AppExporter interface for Cosmos SDK v0.50.6
func (a appCreator) ExportApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts types.AppOptions,
) (types.ExportedApp, error) {
	app := app.NewFluentumApp(
		logger,
		db,
		traceStore,
		height == 0, // loadLatest if height=0
		appOpts,
		a.encCfg,
	)
	return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
}

// AppBlockHeight implements types.AppExporter interface for Cosmos SDK v0.50.6
func (a appCreator) AppBlockHeight() (int64, error) {
	// For now, return 0 as we don't have a persistent app instance
	// In a real implementation, you would return the actual block height
	return 0, nil
}

// loadQuantumSigner loads the quantum signer plugin if enabled in config.
func loadQuantumSigner(cfg *config.Config) error {
	if !cfg.Quantum.Enabled {
		return nil
	}

	if err := plugin.LoadQuantumSigner(cfg.Quantum.LibPath); err != nil {
		return fmt.Errorf("failed to load quantum signer: %v", err)
	}

	fmt.Println("[Quantum] Quantum signing enabled", "mode", cfg.Quantum.Mode)
	return nil
}

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
func AddGenesisAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid address: %s", args[0])
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse coins: %w", err)
			}

			// create concrete account type based on input parameters
			var genAccount authtypes.GenesisAccount

			balances := banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
			baseAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)

			genAccount = baseAccount

			if err := genAccount.Validate(); err != nil {
				return fmt.Errorf("failed to validate new genesis account: %w", err)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			if accs.Contains(addr) {
				return fmt.Errorf("cannot add account at existing address %s", addr)
			}

			// Add the new account to the set of genesis accounts and sanitize the
			// accounts afterwards.
			accs = append(accs, genAccount)
			accs = authtypes.SanitizeGenesisAccounts(accs)

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[authtypes.ModuleName] = authGenStateBz

			bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)
			bankGenState.Balances = append(bankGenState.Balances, balances)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}

			appState[banktypes.ModuleName] = bankGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// createStartCommand creates the start command for the Fluentum node
func createStartCommand(encodingConfig app.EncodingConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Fluentum node",
		Long: `Start the Fluentum blockchain node.

This command starts the Fluentum node with the following features:
- ABCI application server
- P2P networking
- Consensus engine
- RPC server
- API server (if enabled)

The node will connect to the network and begin participating in consensus.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startNode(cmd, encodingConfig)
		},
	}

	// Add flags for configuration
	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagChainID, "", "The network chain ID")
	cmd.Flags().String("log_level", "info", "Log level (debug, info, warn, error)")
	cmd.Flags().Bool("api", true, "Enable the API server")
	cmd.Flags().String("api.address", "tcp://0.0.0.0:1317", "The API server listen address")
	cmd.Flags().Bool("grpc", true, "Enable the gRPC server")
	cmd.Flags().String("grpc.address", "0.0.0.0:9090", "The gRPC server listen address")
	cmd.Flags().Bool("grpc-web", true, "Enable the gRPC-Web server")
	cmd.Flags().String("grpc-web.address", "0.0.0.0:9091", "The gRPC-Web server listen address")
	cmd.Flags().String("rpc.address", "tcp://0.0.0.0:26657", "The RPC server listen address")
	cmd.Flags().String("p2p.address", "tcp://0.0.0.0:26656", "The P2P server listen address")
	cmd.Flags().String("seeds", "", "Comma-separated list of seed nodes")
	cmd.Flags().String("persistent_peers", "", "Comma-separated list of persistent peers")
	cmd.Flags().Bool("testnet", false, "Run in testnet mode with faster block times")
	cmd.Flags().String("moniker", "fluentum-node", "Node moniker")

	return cmd
}

// startNode starts the Fluentum node
func startNode(cmd *cobra.Command, encodingConfig app.EncodingConfig) error {
	// Get configuration from flags
	homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	logLevel, _ := cmd.Flags().GetString("log_level")
	apiEnabled, _ := cmd.Flags().GetBool("api")
	apiAddress, _ := cmd.Flags().GetString("api.address")
	grpcEnabled, _ := cmd.Flags().GetBool("grpc")
	grpcAddress, _ := cmd.Flags().GetString("grpc.address")
	grpcWebEnabled, _ := cmd.Flags().GetBool("grpc-web")
	grpcWebAddress, _ := cmd.Flags().GetString("grpc-web.address")
	rpcAddress, _ := cmd.Flags().GetString("rpc.address")
	p2pAddress, _ := cmd.Flags().GetString("p2p.address")
	seeds, _ := cmd.Flags().GetString("seeds")
	persistentPeers, _ := cmd.Flags().GetString("persistent_peers")
	testnetMode, _ := cmd.Flags().GetBool("testnet")
	moniker, _ := cmd.Flags().GetString("moniker")

	// Set default chain ID if not provided
	if chainID == "" {
		if testnetMode {
			chainID = "fluentum-testnet-1"
		} else {
			chainID = "fluentum-mainnet-1"
		}
	}

	// Create logger
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger = logger.With("module", "main")

	// Log startup information
	logger.Info("Starting Fluentum node",
		"chain_id", chainID,
		"home", homeDir,
		"moniker", moniker,
		"testnet", testnetMode,
		"log_level", logLevel,
	)

	// Create app creator
	appCreator := appCreator{encCfg: encodingConfig}

	// Create server context
	serverCtx := server.NewDefaultContext()
	serverCtx.Config.SetRoot(homeDir)

	// Configure the server
	serverCtx.Config.Moniker = moniker
	serverCtx.Config.RPC.ListenAddress = rpcAddress
	serverCtx.Config.P2P.ListenAddress = p2pAddress
	serverCtx.Config.P2P.Seeds = seeds
	serverCtx.Config.P2P.PersistentPeers = persistentPeers

	// Configure consensus for testnet mode
	if testnetMode {
		serverCtx.Config.Consensus.TimeoutCommit = "1s"
		serverCtx.Config.Consensus.TimeoutPropose = "1s"
		serverCtx.Config.Consensus.CreateEmptyBlocks = true
		serverCtx.Config.Consensus.CreateEmptyBlocksInterval = "10s"
		logger.Info("Configured for testnet mode with faster block times")
	}

	// Create app options
	appOpts := server.DefaultStartOptions()
	appOpts.AppCreator = appCreator
	appOpts.AppCreatorOpts = []interface{}{encodingConfig}

	// Configure API
	if apiEnabled {
		appOpts.APIEnable = true
		appOpts.APIAddress = apiAddress
		appOpts.APISwagger = true
		logger.Info("API server enabled", "address", apiAddress)
	}

	// Configure gRPC
	if grpcEnabled {
		appOpts.GRPCEnable = true
		appOpts.GRPCAddress = grpcAddress
		logger.Info("gRPC server enabled", "address", grpcAddress)
	}

	// Configure gRPC-Web
	if grpcWebEnabled {
		appOpts.GRPCWebEnable = true
		appOpts.GRPCWebAddress = grpcWebAddress
		logger.Info("gRPC-Web server enabled", "address", grpcWebAddress)
	}

	// Start the server
	logger.Info("Starting server...")
	if err := server.Start(serverCtx, appOpts); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for interrupt signal
	logger.Info("Fluentum node is running. Press Ctrl+C to exit.")
	select {}
}

func main() {
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

	rootCmd, _ := NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
