package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types" // TODO: temp
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	stakingKeeper  slashingtypes.StakingKeeper
	govKeeper      govkeeper.Keeper
	wk             wasmkeeper.Keeper
	contractKeeper wasmkeeper.PermissionedKeeper

	authority string
}

func NewKeeper(
	key storetypes.StoreKey,
	cdc codec.BinaryCodec,
	stakingKeeper slashingtypes.StakingKeeper,
	govKeeper govkeeper.Keeper,
	wasmkeeper wasmkeeper.Keeper,
	contractKeeper wasmkeeper.PermissionedKeeper,
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
func (k Keeper) GetContractKeeper() wasmkeeper.PermissionedKeeper {
	return k.contractKeeper
}

func (k Keeper) GetWasmKeeper() wasmkeeper.Keeper {
	return k.wk
}

func (k Keeper) GetStakingKeeper() slashingtypes.StakingKeeper {
	return k.stakingKeeper
}
