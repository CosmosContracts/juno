package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/x/feeshare/types"
)

// GetFeeShares returns all registered FeeShares.
func (k Keeper) GetFeeShares(ctx context.Context) []types.FeeShare {
	feeshares := []types.FeeShare{}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixFeeShare)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var feeshare types.FeeShare
		k.cdc.MustUnmarshal(iterator.Value(), &feeshare)

		feeshares = append(feeshares, feeshare)
	}

	return feeshares
}

// IterateFeeShares iterates over all registered contracts and performs a
// callback with the corresponding FeeShare.
func (k Keeper) IterateFeeShares(
	ctx context.Context,
	handlerFn func(fee types.FeeShare) (stop bool),
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixFeeShare)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var feeshare types.FeeShare
		k.cdc.MustUnmarshal(iterator.Value(), &feeshare)

		if handlerFn(feeshare) {
			break
		}
	}
}

// GetFeeShare returns the FeeShare for a registered contract
func (k Keeper) GetFeeShare(
	ctx context.Context,
	contract sdk.Address,
) (types.FeeShare, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixFeeShare)
	bz := prefix.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.FeeShare{}, false
	}

	var feeshare types.FeeShare
	k.cdc.MustUnmarshal(bz, &feeshare)
	return feeshare, true
}

// SetFeeShare stores the FeeShare for a registered contract.
func (k Keeper) SetFeeShare(ctx context.Context, feeshare types.FeeShare) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixFeeShare)
	key := feeshare.GetContractAddr()
	bz := k.cdc.MustMarshal(&feeshare)
	prefix.Set(key.Bytes(), bz)
}

// DeleteFeeShare deletes a FeeShare of a registered contract.
func (k Keeper) DeleteFeeShare(ctx context.Context, fee types.FeeShare) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixFeeShare)
	key := fee.GetContractAddr()
	prefix.Delete(key.Bytes())
}

// SetDeployerMap stores a contract-by-deployer mapping
func (k Keeper) SetDeployerMap(
	ctx context.Context,
	deployer sdk.AccAddress,
	contract sdk.Address,
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	prefix.Set(key, []byte{1})
}

// DeleteDeployerMap deletes a contract-by-deployer mapping
func (k Keeper) DeleteDeployerMap(
	ctx context.Context,
	deployer sdk.AccAddress,
	contract sdk.Address,
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	prefix.Delete(key)
}

// SetWithdrawerMap stores a contract-by-withdrawer mapping
func (k Keeper) SetWithdrawerMap(
	ctx context.Context,
	withdrawer sdk.AccAddress,
	contract sdk.Address,
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	prefix.Set(key, []byte{1})
}

// DeleteWithdrawMap deletes a contract-by-withdrawer mapping
func (k Keeper) DeleteWithdrawerMap(
	ctx context.Context,
	withdrawer sdk.AccAddress,
	contract sdk.Address,
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	prefix.Delete(key)
}

// IsFeeShareRegistered checks if a contract was registered for receiving
// transaction fees
func (k Keeper) IsFeeShareRegistered(
	ctx context.Context,
	contract sdk.Address,
) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixFeeShare)
	return prefix.Has(contract.Bytes())
}

// IsDeployerMapSet checks if a given contract-by-withdrawer mapping is set in
// store
func (k Keeper) IsDeployerMapSet(
	ctx context.Context,
	deployer sdk.AccAddress,
	contract sdk.Address,
) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	return prefix.Has(key)
}

// IsWithdrawerMapSet checks if a give contract-by-withdrawer mapping is set in
// store
func (k Keeper) IsWithdrawerMapSet(
	ctx context.Context,
	withdrawer sdk.AccAddress,
	contract sdk.Address,
) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := prefix.NewStore(store, types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	return prefix.Has(key)
}
