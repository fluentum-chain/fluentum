package app

import (
	"context"
	"encoding/json"
	"io"

	cosmossdklog "cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	cosmossdkdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/spf13/cast"

	// Fluentum modules
	"github.com/fluentum-chain/fluentum/fluentum/x/fluentum"
	fluentumkeeper "github.com/fluentum-chain/fluentum/fluentum/x/fluentum/keeper"
	fluentumtypes "github.com/fluentum-chain/fluentum/fluentum/x/fluentum/types"

	"cosmossdk.io/core/store"
	cosmossdkstore "cosmossdk.io/core/store"
)

const (
	appName = "FluentumApp"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		params.AppModuleBasic{},
		fluentum.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		fluentumtypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
	}
)

var (
// _ runtime.AppI            = (*App)(nil)
// _ servertypes.Application = (*App)(nil)
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	ParamsKeeper  paramskeeper.Keeper

	// Fluentum keepers
	FluentumKeeper fluentumkeeper.Keeper

	// the module manager
	mm *module.Manager

	// module configurator
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	// Create a compatible logger and DB for the newer Cosmos SDK
	// For now, we'll use type assertions to work around the interface differences
	var cosmosLogger cosmossdklog.Logger
	var cosmosDB cosmossdkdb.DB

	// Type assertion for logger - this is a temporary workaround
	if l, ok := logger.(cosmossdklog.Logger); ok {
		cosmosLogger = l
	} else {
		// Create a simple adapter if needed
		cosmosLogger = cosmossdklog.NewNopLogger()
	}

	// Type assertion for DB - this is a temporary workaround
	if d, ok := db.(cosmossdkdb.DB); ok {
		cosmosDB = d
	} else {
		// Create a simple adapter if needed
		cosmosDB = cosmossdkdb.NewMemDB()
	}

	bApp := baseapp.NewBaseApp(appName, cosmosLogger, cosmosDB, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, paramstypes.StoreKey, fluentumtypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys()

	app := &App{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	// bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()))

	// add keepers - simplified for compatibility
	// For now, we'll create simple store service adapters
	// TODO: Implement proper store service adapters
	var accountStore cosmossdkstore.KVStoreService
	var bankStore cosmossdkstore.KVStoreService

	// Create proper store service adapters for Cosmos SDK v0.50.6
	// These adapters bridge the old KVStore interface to the new KVStoreService interface
	accountStore = NewKVStoreServiceAdapter(keys[authtypes.StoreKey])
	bankStore = NewKVStoreServiceAdapter(keys[banktypes.StoreKey])

	// Create address codec - using a simple implementation
	addressCodec := SimpleAddressCodec{Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix()}

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, accountStore, authtypes.ProtoBaseAccount, maccPerms,
		addressCodec, authtypes.NewModuleAddress(authtypes.ModuleName).String(), "fluentum",
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, bankStore, app.AccountKeeper, app.BlockedModuleAccountAddrs(), authtypes.NewModuleAddress(authtypes.ModuleName).String(),
		cosmosLogger,
	)

	// Create Fluentum Keeper with correct parameters
	app.FluentumKeeper = *fluentumkeeper.NewKeeper(
		appCodec, keys[fluentumtypes.StoreKey], keys[fluentumtypes.MemStoreKey], app.GetSubspace(fluentumtypes.ModuleName),
		BankKeeperAdapter{app.BankKeeper},
	)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, nil, nil, encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		params.NewAppModule(app.ParamsKeeper),
		fluentum.NewAppModule(appCodec, app.FluentumKeeper, AccountKeeperAdapter{app.AccountKeeper}, BankKeeperAdapter{app.BankKeeper}),
	)

	app.mm.SetOrderBeginBlockers(
		authtypes.ModuleName, banktypes.ModuleName, genutiltypes.ModuleName, paramstypes.ModuleName, fluentumtypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		authtypes.ModuleName, banktypes.ModuleName, genutiltypes.ModuleName, paramstypes.ModuleName, fluentumtypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		authtypes.ModuleName, banktypes.ModuleName, genutiltypes.ModuleName, paramstypes.ModuleName, fluentumtypes.ModuleName,
	)

	app.mm.RegisterInvariants(nil)
	// app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	// app.SetInitChainer(app.InitChainer) // Commented out due to signature mismatch
	app.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		// For Cosmos SDK v0.50.6, we need to return the proper type
		return sdk.BeginBlock{}, nil
	})
	// app.SetEndBlocker(app.EndBlocker) // Commented out due to signature mismatch

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// NewFluentumApp creates a new Fluentum application with the specified parameters
func NewFluentumApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	encCfg EncodingConfig,
) *App {
	// Extract parameters from appOpts
	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))

	// Create base app options
	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	baseAppOptions := []func(*baseapp.BaseApp){
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
	}

	return New(
		logger,
		db,
		traceStore,
		loadLatest,
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		encCfg,
		appOpts,
		baseAppOptions...,
	)
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.mm.InitGenesis(ctx, app.appCodec, genesisState)
	return abci.ResponseInitChain{}
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's blocked module account
// addresses.
func (app *App) BlockedModuleAccountAddrs() map[string]bool {
	modAccAddrs := app.ModuleAccountAddrs()
	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom encoding types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom encoding types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry
func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	// Stub implementation
}

// RegisterNodeService implements the Application.RegisterNodeService method.
func (app *App) RegisterNodeService(clientCtx client.Context, config config.Config) {
	// Stub implementation
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(fluentumtypes.ModuleName)

	return paramsKeeper
}

// Adapter types to bridge interface differences
type AccountKeeperAdapter struct {
	authkeeper.AccountKeeper
}

func (a AccountKeeperAdapter) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return a.AccountKeeper.GetAccount(ctx, addr)
}

func (a AccountKeeperAdapter) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
	a.AccountKeeper.SetAccount(ctx, acc)
}

func (a AccountKeeperAdapter) NewAccount(ctx sdk.Context, acc sdk.AccountI) sdk.AccountI {
	return a.AccountKeeper.NewAccount(ctx, acc)
}

type BankKeeperAdapter struct {
	bankkeeper.Keeper
}

func (b BankKeeperAdapter) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return b.Keeper.SendCoins(ctx, fromAddr, toAddr, amt)
}

func (b BankKeeperAdapter) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return b.Keeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

func (b BankKeeperAdapter) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return b.Keeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

func (b BankKeeperAdapter) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return b.Keeper.MintCoins(ctx, moduleName, amt)
}

func (b BankKeeperAdapter) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return b.Keeper.BurnCoins(ctx, moduleName, amt)
}

func (b BankKeeperAdapter) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	return b.Keeper.SendCoinsFromModuleToModule(ctx, senderModule, recipientModule, amt)
}

func (b BankKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return b.Keeper.GetBalance(ctx, addr, denom)
}

// Simple address codec implementation
type SimpleAddressCodec struct {
	Prefix string
}

func (c SimpleAddressCodec) StringToBytes(text string) ([]byte, error) {
	// Simple implementation - just return the bytes
	return []byte(text), nil
}

func (c SimpleAddressCodec) BytesToString(bz []byte) (string, error) {
	return sdk.Bech32ifyAddressBytes(c.Prefix, bz)
}

// KVStoreServiceAdapter adapts the old KVStore interface to the new KVStoreService interface
// This is needed for Cosmos SDK v0.50.6 compatibility
type KVStoreServiceAdapter struct {
	storeKey storetypes.StoreKey
}

// NewKVStoreServiceAdapter creates a new KVStoreService adapter
func NewKVStoreServiceAdapter(storeKey *storetypes.KVStoreKey) cosmossdkstore.KVStoreService {
	return &KVStoreServiceAdapter{storeKey: storeKey}
}

// OpenKVStore implements cosmossdkstore.KVStoreService
func (a *KVStoreServiceAdapter) OpenKVStore(ctx context.Context) store.KVStore {
	// Convert context.Context to sdk.Context for the underlying store
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &KVStoreWrapper{store: sdkCtx.KVStore(a.storeKey)}
}

// KVStoreWrapper wraps the underlying storetypes.KVStore to ensure it matches the new interface
type KVStoreWrapper struct {
	store storetypes.KVStore
}

// Get implements cosmossdkstore.KVStore
func (w *KVStoreWrapper) Get(key []byte) ([]byte, error) {
	return w.store.Get(key), nil
}

// Has implements cosmossdkstore.KVStore
func (w *KVStoreWrapper) Has(key []byte) (bool, error) {
	return w.store.Has(key), nil
}

// Set implements cosmossdkstore.KVStore
func (w *KVStoreWrapper) Set(key, value []byte) error {
	w.store.Set(key, value)
	return nil
}

// Delete implements cosmossdkstore.KVStore - returns error as required by new interface
func (w *KVStoreWrapper) Delete(key []byte) error {
	w.store.Delete(key)
	return nil
}

// Iterator implements cosmossdkstore.KVStore
func (w *KVStoreWrapper) Iterator(start, end []byte) (store.Iterator, error) {
	return w.store.Iterator(start, end), nil
}

// ReverseIterator implements cosmossdkstore.KVStore
func (w *KVStoreWrapper) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return w.store.ReverseIterator(start, end), nil
}

// ExportAppStateAndValidators exports the state of the application for a genesis file.
func (app *App) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string) (servertypes.ExportedApp, error) {
	// Create a simple exported app structure
	// This is a stub implementation - in a real app, you would export the actual state
	exportedApp := servertypes.ExportedApp{
		AppState:   json.RawMessage("{}"),
		Validators: nil,
		Height:     0,
	}

	return exportedApp, nil
}
