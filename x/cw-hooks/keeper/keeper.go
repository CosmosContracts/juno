package keeper

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/CosmosContracts/juno/v20/x/cw-hooks/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	stakingKeeper  slashingtypes.StakingKeeper
	govKeeper      govkeeper.Keeper
	wk             wasmkeeper.Keeper
	contractKeeper wasmtypes.ContractOpsKeeper

	authority string
}

func NewKeeper(
	key storetypes.StoreKey,
	cdc codec.BinaryCodec,
	stakingKeeper slashingtypes.StakingKeeper,
	govKeeper govkeeper.Keeper,
	wasmkeeper wasmkeeper.Keeper,
	contractKeeper wasmtypes.ContractOpsKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       key,
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

func (k Keeper) GetStakingKeeper() slashingtypes.StakingKeeper {
	return k.stakingKeeper
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
