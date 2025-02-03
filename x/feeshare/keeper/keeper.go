package keeper

import (
	"context"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"cosmossdk.io/log"

	storetypes "cosmossdk.io/core/store"
	legacystoretypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	revtypes "github.com/CosmosContracts/juno/v27/x/feeshare/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	cdc            codec.BinaryCodec
	storeService   storetypes.KVStoreService
	legacyStoreKey legacystoretypes.StoreKey

	bankKeeper    bankkeeper.Keeper
	wasmKeeper    wasmkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper

	feeCollectorName string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	legacyStoreKey legacystoretypes.StoreKey,
	ss storetypes.KVStoreService,
	cdc codec.BinaryCodec,
	bk bankkeeper.Keeper,
	wk wasmkeeper.Keeper,
	ak authkeeper.AccountKeeper,
	feeCollector string,
	authority string,
) Keeper {
	return Keeper{
		legacyStoreKey:   legacyStoreKey,
		storeService:     ss,
		cdc:              cdc,
		bankKeeper:       bk,
		wasmKeeper:       wk,
		accountKeeper:    ak,
		feeCollectorName: feeCollector,
		authority:        authority,
	}
}

// GetAuthority returns the x/feeshare module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", revtypes.ModuleName))
}
