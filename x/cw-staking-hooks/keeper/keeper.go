package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types" // TODO: temp

	"github.com/CosmosContracts/juno/v17/x/cw-staking-hooks/types"
)

// Keeper of the juno staking keeper store
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	stakingKeeper  slashingtypes.StakingKeeper
	contractKeeper wasmkeeper.PermissionedKeeper

	authority string
}

func NewKeeper(
	key storetypes.StoreKey,
	cdc codec.BinaryCodec,
	stakingKeeper slashingtypes.StakingKeeper,
	contractKeeper wasmkeeper.PermissionedKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		contractKeeper: contractKeeper,
		stakingKeeper:  stakingKeeper,
		authority:      authority,
	}
}

// GetAuthority returns the x/cw-staking-hooks module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetParams sets the x/cw-staking-hooks module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams returns the current x/cw-staking-hooks module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// GetContractKeeper returns the x/wasm module's contract keeper.
func (k Keeper) GetContractKeeper() wasmkeeper.PermissionedKeeper {
	return k.contractKeeper
}

func (k Keeper) GetStakingKeeper() slashingtypes.StakingKeeper {
	return k.stakingKeeper
}
