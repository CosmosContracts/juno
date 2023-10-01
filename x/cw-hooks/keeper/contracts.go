package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetContract(ctx sdk.Context, keyPrefix []byte, contractAddr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), keyPrefix)
	store.Set(contractAddr.Bytes(), []byte{})
}

func (k Keeper) IsContractRegistered(ctx sdk.Context, keyPrefix []byte, contractAddr sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), keyPrefix)
	return store.Has(contractAddr.Bytes())
}

func (k Keeper) IterateContracts(
	ctx sdk.Context,
	keyPrefix []byte,
	handlerFn func(contractAddr []byte) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		keyAddr := iterator.Key()[len(keyPrefix):]
		addr := sdk.AccAddress(keyAddr)

		if handlerFn(addr) {
			break
		}
	}
}

func (k Keeper) GetAllContracts(ctx sdk.Context, keyPrefix []byte) (list []sdk.Address) {
	k.IterateContracts(ctx, keyPrefix, func(addr []byte) bool {
		list = append(list, sdk.AccAddress(addr))
		return false
	})
	return
}
func (k Keeper) GetAllContractsBech32(ctx sdk.Context, keyPrefix []byte) []string {
	contracts := k.GetAllContracts(ctx, keyPrefix)

	list := make([]string, 0, len(contracts))
	for _, c := range contracts {

		c := sdk.MustBech32ifyAddressBytes("juno", c.Bytes())
		list = append(list, c)
	}
	return list
}

func (k Keeper) DeleteContract(ctx sdk.Context, keyPrefix []byte, contractAddr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), keyPrefix)
	store.Delete(contractAddr)
}

func (k Keeper) ExecuteMessageOnContracts(ctx sdk.Context, keyPrefix []byte, msgBz []byte) error {
	p := k.GetParams(ctx)

	for _, c := range k.GetAllContracts(ctx, keyPrefix) {
		gasLimitCtx := ctx.WithGasMeter(sdk.NewGasMeter(p.ContractGasLimit))
		if _, err := k.GetContractKeeper().Sudo(gasLimitCtx, sdk.AccAddress(c.Bytes()), msgBz); err != nil {
			fmt.Println("ExecuteMessageOnContracts error: ", err)
			return err
		}
	}

	return nil
}
