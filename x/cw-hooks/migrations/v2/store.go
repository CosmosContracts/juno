package v2

import (
	"context"

	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/x/cw-hooks/keeper"
	"github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
)

var (
	ParamsKey        = []byte{0x00}
	KeyPrefixStaking = []byte{0x01}
	KeyPrefixGov     = []byte{0x02}
)

func MigrateStoreToCollections(ctx context.Context, storeService store.KVStoreService, cdc codec.BinaryCodec, k *keeper.Keeper) error {
	kvstore := runtime.KVStoreAdapter(storeService.OpenKVStore(ctx))
	paramsBz := kvstore.Get(ParamsKey)
	var params types.Params
	err := cdc.Unmarshal(paramsBz, &params)
	if err != nil {
		return err
	}

	if err := k.Params.Set(ctx, params); err != nil {
		return err
	}

	stakingIterator := storetypes.KVStorePrefixIterator(kvstore, KeyPrefixStaking)
	defer stakingIterator.Close() //nolint:errcheck

	for ; stakingIterator.Valid(); stakingIterator.Next() {
		key := stakingIterator.Key()
		if len(key) <= len(KeyPrefixStaking) {
			continue
		}

		addr := sdk.AccAddress(key[len(KeyPrefixStaking):])

		err = k.SetContract(ctx, types.StakingPrefixKey, types.ContractInfo{
			ContractAddress: addr.String(),
			FailureCounter:  0,
		})
		if err != nil {
			return err
		}
	}

	govIterator := storetypes.KVStorePrefixIterator(kvstore, KeyPrefixGov)
	defer govIterator.Close() //nolint:errcheck

	for ; govIterator.Valid(); govIterator.Next() {
		key := govIterator.Key()
		if len(key) <= len(KeyPrefixGov) {
			continue
		}

		govAddr := sdk.AccAddress(key[len(KeyPrefixGov):])

		err = k.SetContract(ctx, types.GovPrefixKey, types.ContractInfo{
			ContractAddress: govAddr.String(),
			FailureCounter:  0,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
