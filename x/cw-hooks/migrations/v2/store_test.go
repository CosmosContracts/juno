package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/testutil/setup"
	v2 "github.com/CosmosContracts/juno/v31/x/cw-hooks/migrations/v2"
	"github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

func TestMigrateStoreToCollections(t *testing.T) {
	t.Parallel()

	app := setup.Setup(false, t.TempDir(), "cwhooks-migration-test", t)
	ctx := app.NewContextLegacy(false, tmproto.Header{})

	cwKeeper := &app.AppKeepers.CWHooksKeeper
	cdc := cwKeeper.GetCdc()
	storeService := cwKeeper.GetStoreService()
	store := runtime.KVStoreAdapter(storeService.OpenKVStore(ctx))

	// seed legacy params
	legacyParams := types.Params{
		ContractGasLimit: 777_777,
	}
	store.Set(v2.ParamsKey, cdc.MustMarshal(&legacyParams))

	stakingAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	govAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	prefix.NewStore(store, v2.KeyPrefixStaking).Set(stakingAddr.Bytes(), []byte{})
	prefix.NewStore(store, v2.KeyPrefixGov).Set(govAddr.Bytes(), []byte{})

	err := v2.MigrateStoreToCollections(ctx, storeService, cdc, cwKeeper)
	require.NoError(t, err)

	params, err := app.AppKeepers.CWHooksKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, legacyParams.ContractGasLimit, params.ContractGasLimit)

	stakingContracts, err := cwKeeper.GetAllContracts(ctx, types.StakingPrefixKey)
	require.NoError(t, err)
	require.Len(t, stakingContracts, 1)
	require.Equal(t, stakingAddr.String(), stakingContracts[0].ContractAddress)
	require.Zero(t, stakingContracts[0].FailureCounter)

	govContracts, err := cwKeeper.GetAllContracts(ctx, types.GovPrefixKey)
	require.NoError(t, err)
	require.Len(t, govContracts, 1)
	require.Equal(t, govAddr.String(), govContracts[0].ContractAddress)
	require.Zero(t, govContracts[0].FailureCounter)
}
