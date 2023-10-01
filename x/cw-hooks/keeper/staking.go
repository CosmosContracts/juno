package keeper

import (
	"github.com/CosmosContracts/juno/v17/x/cw-hooks/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetStakingContract(ctx sdk.Context, c types.Contract) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixStakingRegister)
	key := sdk.MustAccAddressFromBech32(c.GetContractAddress())
	bz := k.cdc.MustMarshal(&c)
	store.Set(key.Bytes(), bz)
}

func (k Keeper) GetStakingContract(ctx sdk.Context, addr sdk.AccAddress) (types.Contract, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixStakingRegister)
	bz := store.Get(addr.Bytes())
	if bz == nil {
		return types.Contract{}, false
	}
	var c types.Contract
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

func (k Keeper) IsStakingContractRegistered(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixStakingRegister)
	return store.Has(addr.Bytes())
}

func (k Keeper) IterateStakingContracts(
	ctx sdk.Context,
	handlerFn func(fee types.Contract) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixStakingRegister)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var c types.Contract
		k.cdc.MustUnmarshal(iterator.Value(), &c)

		if handlerFn(c) {
			break
		}
	}
}

func (k Keeper) GetAllStakingContract(ctx sdk.Context) (list []types.Contract) {
	k.IterateStakingContracts(ctx, func(c types.Contract) bool {
		list = append(list, c)
		return false
	})
	return
}

func (k Keeper) DeleteStakingContract(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixStakingRegister)
	store.Delete(addr.Bytes())
}

func (k Keeper) ExecuteMessageOnStakingContracts(ctx sdk.Context, msgBz []byte) error {
	for _, c := range k.GetAllStakingContract(ctx) {
		// TODO: make this a gas limit param
		gasLimitCtx := ctx.WithGasMeter(sdk.NewGasMeter(250_000))
		if _, err := k.contractKeeper.Sudo(gasLimitCtx, sdk.MustAccAddressFromBech32(c.GetContractAddress()), msgBz); err != nil {
			return err
		}
	}

	return nil
}
