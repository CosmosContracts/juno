package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	helpers "github.com/CosmosContracts/juno/v29/app/helpers"
)

func (k Keeper) SetContract(ctx context.Context, keyPrefix []byte, contractAddr sdk.AccAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, keyPrefix)
	prefix.Set(contractAddr.Bytes(), []byte{})
}

func (k Keeper) IsContractRegistered(ctx context.Context, keyPrefix []byte, contractAddr sdk.AccAddress) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, keyPrefix)
	return prefix.Has(contractAddr.Bytes())
}

func (k Keeper) IterateContracts(
	ctx context.Context,
	keyPrefix []byte,
	handlerFn func(contractAddr []byte) (stop bool),
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		keyAddr := iterator.Key()[len(keyPrefix):]
		addr := sdk.AccAddress(keyAddr)

		if handlerFn(addr) {
			break
		}
	}
}

func (k Keeper) GetAllContracts(ctx context.Context, keyPrefix []byte) (list []sdk.Address) {
	k.IterateContracts(ctx, keyPrefix, func(addr []byte) bool {
		list = append(list, sdk.AccAddress(addr))
		return false
	})
	return
}

func (k Keeper) GetAllContractsBech32(ctx context.Context, keyPrefix []byte) []string {
	contracts := k.GetAllContracts(ctx, keyPrefix)

	list := make([]string, 0, len(contracts))
	for _, c := range contracts {
		list = append(list, c.String())
	}
	return list
}

func (k Keeper) DeleteContract(ctx context.Context, keyPrefix []byte, contractAddr sdk.AccAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, keyPrefix)
	prefix.Delete(contractAddr)
}

func (k Keeper) ExecuteMessageOnContracts(ctx context.Context, keyPrefix []byte, msgBz []byte) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	p := k.GetParams(ctx)

	for _, c := range k.GetAllContracts(ctx, keyPrefix) {
		gasLimitCtx := sdkCtx.WithGasMeter(storetypes.NewGasMeter(p.ContractGasLimit))
		addr := sdk.AccAddress(c.Bytes())

		var err error
		helpers.ExecuteContract(k.GetContractKeeper(), gasLimitCtx, addr, msgBz, &err)
		if err != nil {
			k.Logger(ctx).Error("ExecuteMessageOnContracts err", err, "contract", addr.String())
			return err
		}
	}

	return nil
}
