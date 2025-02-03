package keeper

import (
	"context"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/log"

	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/x/clock/types"
)

// Keeper of the clock store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	wasmKeeper     wasmkeeper.Keeper
	contractKeeper wasmtypes.ContractOpsKeeper

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	wasmKeeper wasmkeeper.Keeper,
	contractKeeper wasmtypes.ContractOpsKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeService:   ss,
		wasmKeeper:     wasmKeeper,
		contractKeeper: contractKeeper,
		authority:      authority,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the x/clock module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetParams sets the x/clock module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&p)
	err := store.Set(types.ParamsKey, bz)
	if err != nil {
		return err
	}

	return nil
}

// GetParams returns the current x/clock module parameters.
func (k Keeper) GetParams(ctx context.Context) (p types.Params) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}
	if err != nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// GetContractKeeper returns the x/wasm module's contract keeper.
func (k Keeper) GetContractKeeper() wasmtypes.ContractOpsKeeper {
	return k.contractKeeper
}

// GetCdc returns the x/clock module's codec.
func (k Keeper) GetCdc() codec.BinaryCodec {
	return k.cdc
}

// GetStore returns the x/clock module's store service.
func (k Keeper) GetStoreService() storetypes.KVStoreService {
	return k.storeService
}
