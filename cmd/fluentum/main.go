// The content you provided from node 3's main.go will be placed here
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	cometabci "github.com/cometbft/cometbft/abci/types"
	cosmosbaseapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmlog "github.com/cometbft/cometbft/libs/log"
	abcitypes "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/config"
	_ "github.com/fluentum-chain/fluentum/crypto/ed25519"
	"github.com/fluentum-chain/fluentum/fluentum/app"
	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
	fluentumlog "github.com/fluentum-chain/fluentum/libs/log"
	"github.com/fluentum-chain/fluentum/node"
	"github.com/fluentum-chain/fluentum/p2p"
	"github.com/fluentum-chain/fluentum/privval"
	"github.com/fluentum-chain/fluentum/proxy"
	sm "github.com/fluentum-chain/fluentum/state"
	"github.com/fluentum-chain/fluentum/store"
	"github.com/fluentum-chain/fluentum/version"
	"github.com/prometheus/client_golang/prometheus"
)

// AppOptions wrapper for map[string]interface{}
type appOptions map[string]interface{}

func (a appOptions) Get(key string) interface{} { return a[key] }

// Prometheus metrics
var (
	blockHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fluentum_block_height",
		Help: "Current block height of the Fluentum network.",
	})
	transactionsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fluentum_transactions_total",
		Help: "Total number of transactions processed by Fluentum.",
	})
	validatorCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fluentum_validator_count",
		Help: "Number of active validators in the Fluentum network.",
	})
	networkLatency = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fluentum_network_latency",
		Help: "Average network latency in seconds.",
	})
)

func init() {
	prometheus.MustRegister(blockHeight)
	prometheus.MustRegister(transactionsTotal)
	prometheus.MustRegister(validatorCount)
	prometheus.MustRegister(networkLatency)
}

func getFluentumBlockHeightReal(bs *store.BlockStore) int64 {
	if bs == nil {
		return 0
	}
	return bs.Height()
}

// Cache total txs for efficiency
var (
	cachedTxsHeight int64
	cachedTxsTotal  int64
	cachedTxsMu     sync.Mutex
)

func getFluentumTransactionsTotalReal(bs *store.BlockStore) int64 {
	if bs == nil {
		return 0
	}
	base := bs.Base()
	height := bs.Height()
	total := int64(0)
	for h := base; h <= height; h++ {
		meta := bs.LoadBlockMeta(h)
		if meta != nil {
			total += int64(meta.NumTxs)
		}
	}
	return total
}

func getFluentumValidatorCountReal(stateStore sm.Store, bs *store.BlockStore) int64 {
	if bs == nil || stateStore == nil {
		return 0
	}
	height := bs.Height()
	if height == 0 {
		return 0
	}
	vals, err := stateStore.LoadValidators(height)
	if err != nil || vals == nil {
		return 0
	}
	return int64(vals.Size())
}

func getFluentumNetworkLatencyReal(bs *store.BlockStore) float64 {
	if bs == nil {
		return 0
	}
	height := bs.Height()
	if height < 2 {
		return 0
	}
	meta1 := bs.LoadBlockMeta(height)
	meta0 := bs.LoadBlockMeta(height - 1)
	if meta1 == nil || meta0 == nil {
		return 0
	}
	t1 := meta1.Header.Time
	t0 := meta0.Header.Time
	return t1.Sub(t0).Seconds()
}

// NewRootCmd creates a new root command for the Fluentum application.
func NewRootCmd() (*cobra.Command, app.EncodingConfig) {
	encodingConfig := app.MakeEncodingConfig()

	rootCmd := &cobra.Command{
		Use:   "fluentumd",
		Short: "Fluentum Core - A hybrid consensus blockchain",
		Long: `Fluentum Core is a blockchain platform that combines DPoS
and ZK-Rollups for high throughput and security.`,
	}

	// Create commands for the Fluentum node
	startCmd := createStartCommand(encodingConfig)
	initCmd := createInitCommand()
	versionCmd := versionCmd()

	// Add CLI commands (important)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	fmt.Println("DEBUG: About to add query command")
	rootCmd.AddCommand(queryCommand())
	fmt.Println("DEBUG: About to add tx command")
	rootCmd.AddCommand(txCommand())
	// Note: keys command not available in this version, will add later

	return rootCmd, encodingConfig
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig app.EncodingConfig) {
	// Do nothing for now
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand() *cobra.Command {
	fmt.Println("DEBUG: Creating query command")
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("DEBUG: Query command RunE function called")
			return client.ValidateCmd(cmd, args)
		},
	}

	cmd.AddCommand(
		// authcmd.GetAccountCmd(), // Removed - not available in v0.50.6
		rpc.ValidatorCommand(),
		// rpc.BlockCommand(), // Removed - not available in v0.50.6
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	app.ModuleBasics.AddQueryCommands(cmd)

	// Debug: Print all commands to see what's registered
	fmt.Println("DEBUG: Available query commands:")
	for _, subCmd := range cmd.Commands() {
		fmt.Printf("  - %s: %s\n", subCmd.Use, subCmd.Short)
	}

	// Debug: Check if bank command exists
	bankCmd := cmd.Commands()
	for _, c := range bankCmd {
		if c.Use == "bank" {
			fmt.Println("DEBUG: Bank command found!")
			break
		}
	}

	// Manually add bank commands since they're not being registered by ModuleBasics
	fmt.Println("DEBUG: Adding bank commands manually")
	bankQueryCmd := &cobra.Command{
		Use:   "bank",
		Short: "Querying commands for the bank module",
	}

	totalCmd := &cobra.Command{
		Use:   "total",
		Short: "Query the total supply of coins of the chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := banktypes.NewQueryClient(clientCtx)

			res, err := queryClient.TotalSupply(cmd.Context(), &banktypes.QueryTotalSupplyRequest{})
			if err != nil {
				return err
			}

			if res == nil {
				fmt.Println("No result returned (empty state or not available yet).")
				return nil
			}
			return clientCtx.PrintProto(res)
		},
	}

	balancesCmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query for account balances by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := banktypes.NewQueryClient(clientCtx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.AllBalances(cmd.Context(), &banktypes.QueryAllBalancesRequest{
				Address: addr.String(),
			})
			if err != nil {
				return err
			}

			if res == nil {
				fmt.Println("No result returned (empty state or not available yet).")
				return nil
			}
			return clientCtx.PrintProto(res)
		},
	}

	// Add client context flags to each subcommand
	flags.AddQueryFlagsToCmd(totalCmd)
	flags.AddQueryFlagsToCmd(balancesCmd)

	bankQueryCmd.AddCommand(totalCmd, balancesCmd)

	cmd.AddCommand(bankQueryCmd)
	fmt.Println("DEBUG: Bank commands added successfully")

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

// versionCmd returns the version command
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the application binary version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(os.Stdout, "DEBUG: version command executed")
			goVersion := version.GoVersion
			if goVersion == "" {
				goVersion = "unknown"
			}
			fmt.Fprintf(os.Stdout, "Fluentum Core %s\n", version.Version)
			fmt.Fprintf(os.Stdout, "Go Version: %s\n", goVersion)
			fmt.Fprintf(os.Stdout, "Tendermint Core: %s\n", version.TMCoreSemVer)
		},
	}
}

type appCreator struct {
	encCfg app.EncodingConfig
}

// CreateApp implements types.AppCreator interface for Cosmos SDK v0.50.6
func (a appCreator) CreateApp(
	tmLogger tmlog.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	return app.NewFluentumApp(
		tmLogger,
		db,
		traceStore,
		true, // loadLatest
		appOpts,
		a.encCfg,
	)
}

// ExportApp implements types.AppExporter interface for Cosmos SDK v0.50.6
func (a appCreator) ExportApp(
	tmLogger tmlog.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
) (servertypes.ExportedApp, error) {
	app := app.NewFluentumApp(
		tmLogger,
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
			fmt.Println("DEBUG: Start command RunE function called")
			return startNode(cmd, encodingConfig)
		},
	}

	// Add minimal flags for configuration - simplified to avoid parsing issues
	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagChainID, "", "The network chain ID")
	cmd.Flags().String("log_level", "info", "Log level (debug, info, warn, error)")
	cmd.Flags().String("moniker", "fluentum-node", "Node moniker")
	cmd.Flags().Bool("testnet", false, "Run in testnet mode with faster block times")

	return cmd
}

// Adapter to wrap Cosmos SDK app and implement Fluentum abci/types.Application
// Only a few methods are shown as examples; the rest should be implemented similarly.
type CosmosAppAdapter struct {
	App *app.App
}

func (a *CosmosAppAdapter) Info(ctx context.Context, req *abcitypes.InfoRequest) (*abcitypes.InfoResponse, error) {
	cometReq := cometabci.RequestInfo{
		Version:      req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion:   req.P2PVersion,
	}
	resp, err := a.App.BaseApp.Info(&cometReq)
	if err != nil {
		return nil, err
	}

	// If this is the initial state (height 0), return empty AppHash
	var lastBlockAppHash []byte
	if resp.LastBlockHeight == 0 {
		lastBlockAppHash = []byte{}
	} else {
		lastBlockAppHash = resp.LastBlockAppHash
	}

	return &abcitypes.InfoResponse{
		Data:             resp.Data,
		Version:          resp.Version,
		AppVersion:       resp.AppVersion,
		LastBlockHeight:  resp.LastBlockHeight,
		LastBlockAppHash: lastBlockAppHash,
	}, nil
}

func (a *CosmosAppAdapter) Query(ctx context.Context, req *abcitypes.QueryRequest) (*abcitypes.QueryResponse, error) {
	cometReq := cometabci.RequestQuery{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}
	resp, err := a.App.BaseApp.Query(ctx, &cometReq)
	if err != nil {
		return nil, err
	}
	return &abcitypes.QueryResponse{
		Code:      resp.Code,
		Log:       resp.Log,
		Info:      resp.Info,
		Index:     resp.Index,
		Key:       resp.Key,
		Value:     resp.Value,
		ProofOps:  resp.ProofOps,
		Height:    resp.Height,
		Codespace: resp.Codespace,
	}, nil
}

func (a *CosmosAppAdapter) CheckTx(ctx context.Context, req *abcitypes.CheckTxRequest) (*abcitypes.CheckTxResponse, error) {
	cometReq := cometabci.RequestCheckTx{
		Tx:   req.Tx,
		Type: cometabci.CheckTxType(req.Type),
	}
	resp, err := a.App.BaseApp.CheckTx(&cometReq)
	if err != nil {
		return nil, err
	}
	return &abcitypes.CheckTxResponse{
		Code:      resp.Code,
		Data:      resp.Data,
		Log:       resp.Log,
		Info:      resp.Info,
		GasWanted: resp.GasWanted,
		GasUsed:   resp.GasUsed,
		Events:    resp.Events,
		Codespace: resp.Codespace,
	}, nil
}

func (a *CosmosAppAdapter) FinalizeBlock(ctx context.Context, req *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		FinalizeBlock(context.Context, *abcitypes.FinalizeBlockRequest) (*abcitypes.FinalizeBlockResponse, error)
	}); ok {
		return fluentumApp.FinalizeBlock(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.FinalizeBlock(req)
	}
	return nil, errors.New("unsupported BaseApp type for FinalizeBlock")
}

func (a *CosmosAppAdapter) Commit(ctx context.Context, req *abcitypes.CommitRequest) (*abcitypes.CommitResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		Commit(context.Context, *abcitypes.CommitRequest) (*abcitypes.CommitResponse, error)
	}); ok {
		return fluentumApp.Commit(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.Commit()
	}
	return nil, errors.New("unsupported BaseApp type for Commit")
}

func (a *CosmosAppAdapter) InitChain(ctx context.Context, req *abcitypes.InitChainRequest) (*abcitypes.InitChainResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		InitChain(context.Context, *abcitypes.InitChainRequest) (*abcitypes.InitChainResponse, error)
	}); ok {
		return fluentumApp.InitChain(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.InitChain(req)
	}
	return nil, errors.New("unsupported BaseApp type for InitChain")
}

func (a *CosmosAppAdapter) PrepareProposal(ctx context.Context, req *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		PrepareProposal(context.Context, *abcitypes.PrepareProposalRequest) (*abcitypes.PrepareProposalResponse, error)
	}); ok {
		return fluentumApp.PrepareProposal(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.PrepareProposal(req)
	}
	return nil, errors.New("unsupported BaseApp type for PrepareProposal")
}

func (a *CosmosAppAdapter) ProcessProposal(ctx context.Context, req *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		ProcessProposal(context.Context, *abcitypes.ProcessProposalRequest) (*abcitypes.ProcessProposalResponse, error)
	}); ok {
		return fluentumApp.ProcessProposal(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.ProcessProposal(req)
	}
	return nil, errors.New("unsupported BaseApp type for ProcessProposal")
}

func (a *CosmosAppAdapter) ExtendVote(ctx context.Context, req *abcitypes.ExtendVoteRequest) (*abcitypes.ExtendVoteResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		ExtendVote(context.Context, *abcitypes.ExtendVoteRequest) (*abcitypes.ExtendVoteResponse, error)
	}); ok {
		return fluentumApp.ExtendVote(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.ExtendVote(ctx, req)
	}
	return nil, errors.New("unsupported BaseApp type for ExtendVote")
}

func (a *CosmosAppAdapter) VerifyVoteExtension(ctx context.Context, req *abcitypes.VerifyVoteExtensionRequest) (*abcitypes.VerifyVoteExtensionResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		VerifyVoteExtension(context.Context, *abcitypes.VerifyVoteExtensionRequest) (*abcitypes.VerifyVoteExtensionResponse, error)
	}); ok {
		return fluentumApp.VerifyVoteExtension(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.VerifyVoteExtension(req)
	}
	return nil, errors.New("unsupported BaseApp type for VerifyVoteExtension")
}

func (a *CosmosAppAdapter) ListSnapshots(ctx context.Context, req *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		ListSnapshots(context.Context, *abcitypes.ListSnapshotsRequest) (*abcitypes.ListSnapshotsResponse, error)
	}); ok {
		return fluentumApp.ListSnapshots(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.ListSnapshots(req)
	}
	return nil, errors.New("unsupported BaseApp type for ListSnapshots")
}

func (a *CosmosAppAdapter) OfferSnapshot(ctx context.Context, req *abcitypes.OfferSnapshotRequest) (*abcitypes.OfferSnapshotResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		OfferSnapshot(context.Context, *abcitypes.OfferSnapshotRequest) (*abcitypes.OfferSnapshotResponse, error)
	}); ok {
		return fluentumApp.OfferSnapshot(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.OfferSnapshot(req)
	}
	return nil, errors.New("unsupported BaseApp type for OfferSnapshot")
}

func (a *CosmosAppAdapter) LoadSnapshotChunk(ctx context.Context, req *abcitypes.LoadSnapshotChunkRequest) (*abcitypes.LoadSnapshotChunkResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		LoadSnapshotChunk(context.Context, *abcitypes.LoadSnapshotChunkRequest) (*abcitypes.LoadSnapshotChunkResponse, error)
	}); ok {
		return fluentumApp.LoadSnapshotChunk(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.LoadSnapshotChunk(req)
	}
	return nil, errors.New("unsupported BaseApp type for LoadSnapshotChunk")
}

func (a *CosmosAppAdapter) ApplySnapshotChunk(ctx context.Context, req *abcitypes.ApplySnapshotChunkRequest) (*abcitypes.ApplySnapshotChunkResponse, error) {
	if fluentumApp, ok := any(a.App.BaseApp).(interface {
		ApplySnapshotChunk(context.Context, *abcitypes.ApplySnapshotChunkRequest) (*abcitypes.ApplySnapshotChunkResponse, error)
	}); ok {
		return fluentumApp.ApplySnapshotChunk(ctx, req)
	}
	if cosmosApp, ok := any(a.App.BaseApp).(*cosmosbaseapp.BaseApp); ok {
		return cosmosApp.ApplySnapshotChunk(req)
	}
	return nil, errors.New("unsupported BaseApp type for ApplySnapshotChunk")
}

func (a *CosmosAppAdapter) Echo(ctx context.Context, req *abcitypes.EchoRequest) (*abcitypes.EchoResponse, error) {
	// TODO: Implement conversion and call
	return nil, nil
}

// startNode starts the Fluentum node
func startNode(cmd *cobra.Command, encodingConfig app.EncodingConfig) error {
	fmt.Println("DEBUG: startNode function called")

	// Get configuration from flags
	fmt.Println("DEBUG: Getting configuration from flags")
	homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	// logLevel, _ := cmd.Flags().GetString("log_level") // Remove unused logLevel
	moniker, _ := cmd.Flags().GetString("moniker")
	testnetMode, _ := cmd.Flags().GetBool("testnet")
	fmt.Println("DEBUG: Configuration flags retrieved")

	// Set default chain ID if not provided
	if chainID == "" {
		if testnetMode {
			chainID = "fluentum-testnet-1"
		} else {
			chainID = "fluentum-mainnet-1"
		}
	}

	// Ensure home directory exists
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to create home directory: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Join(homeDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize logging
	tmLogger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))
	fluentumLogger := fluentumlog.NewTMLogger(fluentumlog.NewSyncWriter(os.Stdout))

	// Load config from file (NEW)
	nodeConfig := loadConfig(homeDir)
	nodeConfig.RootDir = homeDir
	nodeConfig.Moniker = moniker
	// Do not set ChainID directly (not addressable)

	// Example: Load genesis.json using the new utility
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")
	var genesisDoc map[string]interface{} // TODO: Replace with your actual GenesisDoc type
	_, err := loadJSONConfig(genesisPath, &genesisDoc)
	if err != nil {
		return fmt.Errorf("failed to load genesis.json: %w", err)
	}

	// Example: Load priv_validator_key.json using the new utility (for demonstration)
	privValKeyPath := filepath.Join(homeDir, "config", "priv_validator_key.json")
	var privValKey map[string]interface{} // TODO: Replace with your actual PrivValidatorKey type if you want to inspect raw JSON
	_, err = loadJSONConfig(privValKeyPath, &privValKey)
	if err != nil {
		return fmt.Errorf("failed to load priv_validator_key.json: %w", err)
	}
	// The real privVal used by the node is still loaded below with privval.LoadOrGenFilePV

	// Example: Load priv_validator_state.json using the new utility (for demonstration)
	privValStatePath := filepath.Join(homeDir, "data", "priv_validator_state.json")
	var privValState map[string]interface{} // TODO: Replace with your actual PrivValidatorState type if you want to inspect raw JSON
	_, err = loadJSONConfig(privValStatePath, &privValState)
	if err != nil {
		return fmt.Errorf("failed to load priv_validator_state.json: %w", err)
	}
	// The real privVal used by the node is still loaded below with privval.LoadOrGenFilePV

	// Example: Load node_key.json using the new utility (for demonstration)
	nodeKeyPath := filepath.Join(homeDir, "config", "node_key.json")
	var nodeKeyRaw map[string]interface{} // TODO: Replace with your actual NodeKey type if you want to inspect raw JSON
	_, err = loadJSONConfig(nodeKeyPath, &nodeKeyRaw)
	if err != nil {
		return fmt.Errorf("failed to load node_key.json: %w", err)
	}
	// The real nodeKey used by the node is still loaded below with p2p.LoadOrGenNodeKey

	// Example: Load addrbook.json using the new utility (for demonstration)
	addrBookPath := filepath.Join(homeDir, "config", "addrbook.json")
	var addrBook map[string]interface{} // TODO: Replace with your actual AddrBook type if you want to inspect raw JSON
	_, err = loadJSONConfig(addrBookPath, &addrBook)
	if err != nil {
		return fmt.Errorf("failed to load addrbook.json: %w", err)
	}
	// The real addrbook is managed by the node's P2P subsystem

	// Override default configuration with flag values if needed
	if testnetMode {
		// nodeConfig.P2P.Seeds = "seed1.testnet:26656,seed2.testnet:26656"
		nodeConfig.Consensus.TimeoutCommit = 1 * time.Second
	}

	// Ensure appOpts is defined
	appOpts := appOptions{}

	// Create the ABCI application
	appInstance := app.NewFluentumApp(
		tmLogger,
		dbm.NewMemDB(),
		nil,
		true,
		appOpts,
		encodingConfig,
	)

	// Load or generate PrivValidator and NodeKey using Fluentum's packages
	privVal := privval.LoadOrGenFilePV(
		nodeConfig.PrivValidatorKey, nodeConfig.PrivValidatorState,
	)
	nodeKey, err := p2p.LoadOrGenNodeKey(nodeConfig.NodeKeyFile())
	if err != nil {
		return fmt.Errorf("failed to load or generate node key: %w", err)
	}

	// Create ClientCreator using Fluentum's proxy
	adapter := &CosmosAppAdapter{App: appInstance}
	clientCreator := proxy.NewLocalClientCreator(adapter)

	// GenesisDocProvider, DBProvider, MetricsProvider
	genesisDocProvider := node.DefaultGenesisDocProviderFunc(nodeConfig)
	dbProvider := node.DefaultDBProvider
	metricsProvider := node.DefaultMetricsProvider(nodeConfig.Instrumentation)

	// Start the node
	nodeInstance, err := node.NewNode(
		nodeConfig,
		privVal,
		nodeKey,
		clientCreator,
		genesisDocProvider,
		dbProvider,
		metricsProvider,
		fluentumLogger,
	)

	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	if err := nodeInstance.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	fluentumLogger.Info("Fluentum node started", "version", version.Version)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fluentumLogger.Info("Shutting down node...")
		if err := nodeInstance.Stop(); err != nil {
			fluentumLogger.Error("Failed to stop node", "error", err)
		}
		os.Exit(0)
	}()

	// Run forever (until signaled)
	select {}
}

func createInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Stub init command (to be implemented)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("init command is not implemented yet.")
		},
	}
}

func main() {
	rootCmd, _ := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error executing root command:", err)
		os.Exit(1)
	}
}

func loadConfig(homeDir string) *config.Config {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(homeDir)
	v.AddConfigPath(filepath.Join(homeDir, "config"))

	cfg := config.DefaultConfig()
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(os.Stderr, "[WARN] config.toml not found in %s or %s/config, using defaults\n", homeDir, homeDir)
		} else {
			panic(fmt.Errorf("fatal error reading config file: %w", err))
		}
	} else {
		if err := v.Unmarshal(cfg); err != nil {
			panic(fmt.Errorf("unable to decode config.toml into config struct: %w", err))
		}
	}
	return cfg
}

// Loads a JSON config file into the given struct pointer.
// If the file does not exist, returns false and does not error.
func loadJSONConfig(path string, out interface{}) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "[WARN] %s not found, using defaults\n", path)
			return false, nil
		}
		return false, fmt.Errorf("error opening %s: %w", path, err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err := dec.Decode(out); err != nil {
		return false, fmt.Errorf("error decoding %s: %w", path, err)
	}
	return true, nil
}
