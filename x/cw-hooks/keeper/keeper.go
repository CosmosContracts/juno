package keeper

import (
	"context"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/log"

	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/CosmosContracts/juno/v27/x/cw-hooks/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	stakingKeeper  stakingkeeper.Keeper
	govKeeper      govkeeper.Keeper
	wk             wasmkeeper.Keeper
	contractKeeper wasmtypes.ContractOpsKeeper

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	stakingKeeper stakingkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	wasmkeeper wasmkeeper.Keeper,
	contractKeeper wasmtypes.ContractOpsKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeService:   ss,
		stakingKeeper:  stakingKeeper,
		govKeeper:      govKeeper,
		contractKeeper: contractKeeper,
		authority:      authority,
		wk:             wasmkeeper,
	}
}

// GetAuthority returns the x/cw-hooks module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetContractKeeper returns the x/wasm module's contract keeper.
func (k Keeper) GetContractKeeper() wasmtypes.ContractOpsKeeper {
	return k.contractKeeper
}

func (k Keeper) GetWasmKeeper() wasmkeeper.Keeper {
	return k.wk
}

func (k Keeper) GetStakingKeeper() stakingkeeper.Keeper {
	return k.stakingKeeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}
