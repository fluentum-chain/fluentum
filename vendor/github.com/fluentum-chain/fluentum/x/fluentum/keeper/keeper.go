package keeper

import (
	"fmt"
	"strconv"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fluentum-chain/fluentum/x/fluentum/types"
)

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
		bankKeeper BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bk BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		bankKeeper: bk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// NewQuerier creates a new querier for the fluentum module
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) interface{} {
	return &Querier{
		k:                k,
		legacyQuerierCdc: legacyQuerierCdc,
	}
}

// Querier defines the querier for the fluentum module
type Querier struct {
	k                Keeper
	legacyQuerierCdc *codec.LegacyAmino
}

// SetFluentum stores a fluentum in the store
func (k Keeper) SetFluentum(ctx sdk.Context, fluentum types.Fluentum) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&fluentum)
	store.Set(types.GetFluentumKey(fluentum.Index), b)
}

// GetFluentum retrieves a fluentum from the store
func (k Keeper) GetFluentum(ctx sdk.Context, index string) (val types.Fluentum, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetFluentumKey(index))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetAllFluentum retrieves all fluentum from the store
func (k Keeper) GetAllFluentum(ctx sdk.Context) (list []types.Fluentum) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefix(types.FluentumKey))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Fluentum
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetFluentumCount retrieves the fluentum count from the store
func (k Keeper) GetFluentumCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	byteKey := types.KeyPrefix(types.FluentumCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	count, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		// Return 0 if the parsing is failed
		return 0
	}

	return count
}

// SetFluentumCount sets the fluentum count in the store
func (k Keeper) SetFluentumCount(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)
	byteKey := types.KeyPrefix(types.FluentumCountKey)
	bz := []byte(strconv.FormatUint(count, 10))
	store.Set(byteKey, bz)
}

// GetParams retrieves the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the params in the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
