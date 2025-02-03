package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/x/feeshare/types"
)

// GetFeeShares returns all registered FeeShares.
func (k Keeper) GetFeeShares(ctx context.Context) []types.FeeShare {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	feeshares := []types.FeeShare{}

	store := sdkCtx.KVStore(k.legacyStoreKey)
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.legacyStoreKey)
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixFeeShare)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.FeeShare{}, false
	}

	var feeshare types.FeeShare
	k.cdc.MustUnmarshal(bz, &feeshare)
	return feeshare, true
}

// SetFeeShare stores the FeeShare for a registered contract.
func (k Keeper) SetFeeShare(ctx context.Context, feeshare types.FeeShare) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixFeeShare)
	key := feeshare.GetContractAddr()
	bz := k.cdc.MustMarshal(&feeshare)
	store.Set(key.Bytes(), bz)
}

// DeleteFeeShare deletes a FeeShare of a registered contract.
func (k Keeper) DeleteFeeShare(ctx context.Context, fee types.FeeShare) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixFeeShare)
	key := fee.GetContractAddr()
	store.Delete(key.Bytes())
}

// SetDeployerMap stores a contract-by-deployer mapping
func (k Keeper) SetDeployerMap(
	ctx context.Context,
	deployer sdk.AccAddress,
	contract sdk.Address,
) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	store.Set(key, []byte{1})
}

// DeleteDeployerMap deletes a contract-by-deployer mapping
func (k Keeper) DeleteDeployerMap(
	ctx context.Context,
	deployer sdk.AccAddress,
	contract sdk.Address,
) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	store.Delete(key)
}

// SetWithdrawerMap stores a contract-by-withdrawer mapping
func (k Keeper) SetWithdrawerMap(
	ctx context.Context,
	withdrawer sdk.AccAddress,
	contract sdk.Address,
) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	store.Set(key, []byte{1})
}

// DeleteWithdrawMap deletes a contract-by-withdrawer mapping
func (k Keeper) DeleteWithdrawerMap(
	ctx context.Context,
	withdrawer sdk.AccAddress,
	contract sdk.Address,
) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	store.Delete(key)
}

// IsFeeShareRegistered checks if a contract was registered for receiving
// transaction fees
func (k Keeper) IsFeeShareRegistered(
	ctx context.Context,
	contract sdk.Address,
) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixFeeShare)
	return store.Has(contract.Bytes())
}

// IsDeployerMapSet checks if a given contract-by-withdrawer mapping is set in
// store
func (k Keeper) IsDeployerMapSet(
	ctx context.Context,
	deployer sdk.AccAddress,
	contract sdk.Address,
) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	return store.Has(key)
}

// IsWithdrawerMapSet checks if a give contract-by-withdrawer mapping is set in
// store
func (k Keeper) IsWithdrawerMapSet(
	ctx context.Context,
	withdrawer sdk.AccAddress,
	contract sdk.Address,
) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.legacyStoreKey), types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	return store.Has(key)
}
