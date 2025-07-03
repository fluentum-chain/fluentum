package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
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

	cometabci "github.com/cometbft/cometbft/abci/types"
	cosmosbaseapp "github.com/cosmos/cosmos-sdk/baseapp"
	abcitypes "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/config"
	cs "github.com/fluentum-chain/fluentum/consensus"
	_ "github.com/fluentum-chain/fluentum/crypto/ed25519" // Import to register types
	"github.com/fluentum-chain/fluentum/fluentum/app"
	"github.com/fluentum-chain/fluentum/fluentum/core"
	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
	tmjson "github.com/fluentum-chain/fluentum/libs/json"
	fluentumlog "github.com/fluentum-chain/fluentum/libs/log"
	tmos "github.com/fluentum-chain/fluentum/libs/os"
	mempl "github.com/fluentum-chain/fluentum/mempool"
	"github.com/fluentum-chain/fluentum/node"
	p2p "github.com/fluentum-chain/fluentum/p2p"
	"github.com/fluentum-chain/fluentum/privval"
	tmproto "github.com/fluentum-chain/fluentum/proto/tendermint/types"
	"github.com/fluentum-chain/fluentum/proxy"
	sm "github.com/fluentum-chain/fluentum/state"
	"github.com/fluentum-chain/fluentum/store"
	"github.com/fluentum-chain/fluentum/types"
	tmtime "github.com/fluentum-chain/fluentum/types/time"
	"github.com/fluentum-chain/fluentum/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
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
		Long: `Fluentum Core is a blockchain platform that combines DPoS and ZK-Rollups
for high throughput and security.`,
	}

	// Create commands for the Fluentum node
	startCmd := createStartCommand(encodingConfig)
	initCmd := createInitCommand()
	versionCmd := versionCmd()

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)

	return rootCmd, encodingConfig
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig app.EncodingConfig) {
	// Do nothing for now
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
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
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
	appOpts servertypes.AppOptions,
) (servertypes.ExportedApp, error) {
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
	logLevel, _ := cmd.Flags().GetString("log_level")
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

	// Ensure data directory exists
	dataDir := filepath.Join(homeDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	fmt.Println("DEBUG: Creating loggers")
	// Create logger (use Fluentum logger for node, CometBFT logger for app)
	fluentumLogger := fluentumlog.NewTMLogger(fluentumlog.NewSyncWriter(os.Stdout)).With("module", "main")
	cometLogger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	fmt.Println("DEBUG: Loggers created")

	// Log startup information
	fluentumLogger.Info("Starting Fluentum node",
		"chain_id", chainID,
		"home", homeDir,
		"moniker", moniker,
		"testnet", testnetMode,
		"log_level", logLevel,
	)

	fmt.Println("DEBUG: Creating app creator")
	// Create app creator
	appCreator := appCreator{encCfg: encodingConfig}
	fmt.Println("DEBUG: App creator created")

	fmt.Println("DEBUG: Loading Tendermint configuration")
	// Load Tendermint configuration from file
	tmConfig := config.DefaultConfig()
	tmConfig.SetRoot(homeDir)

	// Use viper to load configuration from file
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err == nil {
		// Configuration file found, unmarshal it
		if err := viper.Unmarshal(tmConfig); err != nil {
			fluentumLogger.Error("Failed to unmarshal config file, using defaults", "error", err)
		} else {
			fluentumLogger.Info("Configuration loaded from file", "file", viper.ConfigFileUsed())
		}
	} else {
		fluentumLogger.Info("No config file found, using default configuration")
	}

	fmt.Println("DEBUG: Tendermint configuration loaded")

	// Configure the server
	tmConfig.Moniker = moniker

	// Configure consensus for testnet mode
	if testnetMode {
		tmConfig.Consensus.TimeoutCommit = time.Second
		tmConfig.Consensus.TimeoutPropose = time.Second
		tmConfig.Consensus.CreateEmptyBlocks = true
		tmConfig.Consensus.CreateEmptyBlocksInterval = 10 * time.Second
		fluentumLogger.Info("Configured for testnet mode with faster block times")
	}

	fmt.Println("DEBUG: Loading node key")
	// Load or generate node key (use p2p)
	nodeKey, err := p2p.LoadOrGenNodeKey(tmConfig.NodeKeyFile())
	if err != nil {
		return fmt.Errorf("failed to load or generate node key: %w", err)
	}
	fmt.Println("DEBUG: Node key loaded")

	fmt.Println("DEBUG: Loading private validator")
	// Load or generate private validator
	privValidator := privval.LoadOrGenFilePV(tmConfig.PrivValidatorKeyFile(), tmConfig.PrivValidatorStateFile())
	fmt.Println("DEBUG: Private validator loaded")

	// Initialize database
	fluentumLogger.Info("Creating database...")
	db, err := dbm.NewDB("application", dbm.GoLevelDBBackend, tmConfig.DBDir())
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	fluentumLogger.Info("Database created successfully")

	// Wrap appOpts in AppOptions
	appOpts := appOptions{
		"home": homeDir,
	}

	// Create the application instance with CometBFT logger and proper database
	fluentumLogger.Info("Creating Cosmos application...")
	cosmosApp := appCreator.CreateApp(cometLogger, db, nil, appOpts)
	fluentumLogger.Info("Cosmos application created successfully")

	adapter := &CosmosAppAdapter{App: cosmosApp.(*app.App)}
	fluentumLogger.Info("Adapter created successfully")

	// Create the Tendermint node
	fluentumLogger.Info("Creating Tendermint node...")
	n, err := node.NewNode(tmConfig,
		privValidator,
		nodeKey,
		proxy.NewLocalClientCreator(adapter),
		func() (*types.GenesisDoc, error) {
			// Custom genesis provider that handles JSON properly
			genFile := tmConfig.GenesisFile()
			genDocBytes, err := os.ReadFile(genFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read genesis file: %w", err)
			}

			var genDoc types.GenesisDoc
			if err := tmjson.Unmarshal(genDocBytes, &genDoc); err != nil {
				return nil, fmt.Errorf("failed to unmarshal genesis: %w", err)
			}

			return &genDoc, nil
		},
		node.DefaultDBProvider,
		// Use a robust metrics provider that completely avoids Prometheus
		func(chainID string) (*cs.Metrics, *p2p.Metrics, *mempl.Metrics, *sm.Metrics) {
			// Create no-op metrics with explicit nil checks
			csMetrics := cs.NopMetrics()
			if csMetrics == nil {
				csMetrics = cs.NopMetrics()
			}

			p2pMetrics := p2p.NopMetrics()
			if p2pMetrics == nil {
				p2pMetrics = p2p.NopMetrics()
			}

			mempoolMetrics := mempl.NopMetrics()
			if mempoolMetrics == nil {
				mempoolMetrics = mempl.NopMetrics()
			}

			stateMetrics := sm.NopMetrics()
			if stateMetrics == nil {
				stateMetrics = sm.NopMetrics()
			}

			return csMetrics, p2pMetrics, mempoolMetrics, stateMetrics
		},
		fluentumLogger,
	)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}
	fluentumLogger.Info("Tendermint node created successfully")

	// Start the node
	fluentumLogger.Info("Starting Tendermint node...")
	if err := n.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	fluentumLogger.Info("Started node", "nodeInfo", n.Switch().NodeInfo())

	// Stop upon receiving SIGTERM or CTRL-C.
	tmos.TrapSignal(fluentumLogger, func() {
		if n.IsRunning() {
			if err := n.Stop(); err != nil {
				fluentumLogger.Error("unable to stop the node", "error", err)
			}
		}
	})

	// Wait for interrupt signal
	fluentumLogger.Info("Fluentum node is running. Press Ctrl+C to exit.")
	select {}

	// Start Prometheus metrics goroutine with live blockStore only
	go func(stateStore sm.Store, bs *store.BlockStore) {
		for {
			blockHeight.Set(float64(getFluentumBlockHeightReal(bs)))
			transactionsTotal.Set(float64(getFluentumTransactionsTotalReal(bs)))
			validatorCount.Set(float64(getFluentumValidatorCountReal(stateStore, bs)))
			networkLatency.Set(getFluentumNetworkLatencyReal(bs))
			time.Sleep(5 * time.Second)
		}
	}(n.stateStore, n.BlockStore())

	return nil
}

// createInitCommand creates a custom init command that uses the Tendermint approach
func createInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize Fluentum node",
		Long: `Initialize a new Fluentum node with the specified moniker.

This command will:
- Create the necessary directories
- Generate private validator and node keys
- Create a genesis file
- Create a default config.toml file

Example:
  fluentumd init my-node --home /opt/fluentum
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moniker := args[0]

			// Get home directory from flags
			homeDir, err := cmd.Flags().GetString(flags.FlagHome)
			if err != nil {
				return err
			}

			// Get testnet flag
			testnetMode, err := cmd.Flags().GetBool("testnet")
			if err != nil {
				return err
			}

			// Get chain ID from flags
			chainID, err := cmd.Flags().GetString(flags.FlagChainID)
			if err != nil {
				return err
			}

			// Use default chain ID if not provided
			if chainID == "" {
				if testnetMode {
					chainID = "fluentum-testnet-1"
				} else {
					chainID = "fluentum-mainnet-1"
				}
			}

			return initializeNode(homeDir, moniker, chainID)
		},
	}

	// Add flags
	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagChainID, "", "Genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String("default-denom", "stake", "Genesis file default denomination, if left blank default value is 'stake'")
	cmd.Flags().Int("initial-height", 1, "Specify the initial block height at genesis")
	cmd.Flags().BoolP("overwrite", "o", false, "Overwrite the genesis.json file")
	cmd.Flags().Bool("recover", false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool("testnet", false, "Initialize for testnet mode")

	return cmd
}

// initializeNode performs the actual node initialization
func initializeNode(homeDir, moniker, chainID string) error {
	// Create home directory if it doesn't exist
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to create home directory: %w", err)
	}

	// Create config directory
	configDir := filepath.Join(homeDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create data directory
	dataDir := filepath.Join(homeDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Generate private validator key
	privValKeyFile := filepath.Join(configDir, "priv_validator_key.json")
	privValStateFile := filepath.Join(dataDir, "priv_validator_state.json")

	var pv *privval.FilePV
	if tmos.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		fmt.Printf("Found private validator key: %s\n", privValKeyFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		fmt.Printf("Generated private validator key: %s\n", privValKeyFile)
	}

	// Generate node key
	nodeKeyFile := filepath.Join(configDir, "node_key.json")
	if tmos.FileExists(nodeKeyFile) {
		fmt.Printf("Found node key: %s\n", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return fmt.Errorf("failed to generate node key: %w", err)
		}
		fmt.Printf("Generated node key: %s\n", nodeKeyFile)
	}

	// Create genesis file
	genFile := filepath.Join(configDir, "genesis.json")
	if tmos.FileExists(genFile) {
		fmt.Printf("Found genesis file: %s\n", genFile)
	} else {
		// Create a basic genesis structure
		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("can't get pubkey: %w", err)
		}

		// Create genesis document using the proper types
		genDoc := types.GenesisDoc{
			GenesisTime:   tmtime.Now(),
			ChainID:       chainID,
			InitialHeight: 1,
			ConsensusParams: &tmproto.ConsensusParams{
				Block: tmproto.BlockParams{
					MaxBytes:   22020096,
					MaxGas:     -1,
					TimeIotaMs: 10,
				},
				Evidence: tmproto.EvidenceParams{
					MaxAgeNumBlocks: 100000,
					MaxAgeDuration:  time.Duration(172800000000000), // 48 hours in nanoseconds
					MaxBytes:        1048576,
				},
				Validator: tmproto.ValidatorParams{
					PubKeyTypes: []string{"ed25519"},
				},
				Version: tmproto.VersionParams{},
			},
			Validators: []types.GenesisValidator{{
				Address: pubKey.Address(),
				PubKey:  pubKey,
				Power:   10,
			}},
			AppHash:  []byte{},
			AppState: json.RawMessage("{}"),
		}

		// Save the genesis file using tmjson for proper crypto type handling
		genDocBytes, err := tmjson.MarshalIndent(genDoc, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal genesis file: %w", err)
		}

		// Fix numeric fields to be strings instead of integers
		// Note: All numeric fields in consensus_params should be strings according to the error messages
		genDocStr := string(genDocBytes)

		// Convert all numeric fields to string format, including 'power' which should be a string
		genDocStr = strings.ReplaceAll(genDocStr, `"initial_height": 1`, `"initial_height": "1"`)
		genDocStr = strings.ReplaceAll(genDocStr, `"max_bytes": 22020096`, `"max_bytes": "22020096"`)
		genDocStr = strings.ReplaceAll(genDocStr, `"max_gas": -1`, `"max_gas": "-1"`)
		genDocStr = strings.ReplaceAll(genDocStr, `"time_iota_ms": 10`, `"time_iota_ms": "10"`)
		genDocStr = strings.ReplaceAll(genDocStr, `"max_age_num_blocks": 100000`, `"max_age_num_blocks": "100000"`)
		genDocStr = strings.ReplaceAll(genDocStr, `"max_age_duration": 172800000000000`, `"max_age_duration": "172800000000000"`)
		genDocStr = strings.ReplaceAll(genDocStr, `"max_bytes": 1048576`, `"max_bytes": "1048576"`)
		genDocStr = strings.ReplaceAll(genDocStr, `"power": 10`, `"power": "10"`)

		if err := tmos.WriteFile(genFile, []byte(genDocStr), 0o644); err != nil {
			return fmt.Errorf("failed to save genesis file: %w", err)
		}
		fmt.Printf("Generated genesis file: %s\n", genFile)
	}

	// Create default config.toml
	configFile := filepath.Join(configDir, "config.toml")
	if !tmos.FileExists(configFile) {
		defaultConfig := config.DefaultConfig()
		defaultConfig.Moniker = moniker
		defaultConfig.SetRoot(homeDir)

		config.WriteConfigFile(configFile, defaultConfig)
		fmt.Printf("Generated config file: %s\n", configFile)
	}

	fmt.Printf("Initialized Fluentum node with moniker '%s' in %s\n", moniker, homeDir)
	return nil
}

func apiKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		expected := "Bearer b7e2c1f4a8d9e3b6c5f1a2d3e4b8c7f6a1e2d3c4b5a6f7e8c9d0b1a2c3d4e5f6"
		if apiKey != expected {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	stats := map[string]interface{}{
		"block_height":        8423197,
		"transactions_24h":    284200,
		"active_validators":   128,
		"network_utilization": 63.2,
		"average_block_time":  2.3,
		"network_security":    99.98,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
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
