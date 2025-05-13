package keeper

import (
	"context"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	revtypes "github.com/CosmosContracts/juno/v29/x/feeshare/types"
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

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
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	bk bankkeeper.Keeper,
	wk wasmkeeper.Keeper,
	ak authkeeper.AccountKeeper,
	feeCollector string,
	authority string,
) Keeper {
	return Keeper{
		cdc:              cdc,
		storeService:     ss,
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
