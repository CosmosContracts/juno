package app_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/CosmosContracts/juno/v31/testutil/setup"
	clocktypes "github.com/CosmosContracts/juno/v31/x/clock/types"
	cwhookstypes "github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
	feepaytypes "github.com/CosmosContracts/juno/v31/x/feepay/types"
)

func TestExportNormalizesLegacyContractCaps(t *testing.T) {
	chainID := "juno-export-legacy-caps-1"
	app := setup.Setup(false, t.TempDir(), chainID)
	ctx := app.NewContextLegacy(false, cmtproto.Header{Height: 1, ChainID: chainID})

	legacyClock := clocktypes.Params{ContractGasLimit: 250_000, MaxContracts: 0}
	clockStore := app.AppKeepers.ClockKeeper.GetStoreService().OpenKVStore(ctx)
	require.NoError(t, clockStore.Set(clocktypes.ParamsKey, app.AppKeepers.ClockKeeper.GetCdc().MustMarshal(&legacyClock)))
	legacyCWHooks := cwhookstypes.Params{
		ContractGasLimit:                250_000,
		ContractFailureRemovalThreshold: 3,
		MaxContracts:                    0,
	}
	require.NoError(t, app.AppKeepers.CWHooksKeeper.Params.Set(ctx, legacyCWHooks))
	ctx.MultiStore().(storetypes.CacheMultiStore).Write()
	_, err := app.Commit()
	require.NoError(t, err)

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	var state map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(exported.AppState, &state))
	var clockGenesis clocktypes.GenesisState
	require.NoError(t, app.AppCodec().UnmarshalJSON(state[clocktypes.ModuleName], &clockGenesis))
	var cwHooksGenesis cwhookstypes.GenesisState
	require.NoError(t, app.AppCodec().UnmarshalJSON(state[cwhookstypes.ModuleName], &cwHooksGenesis))
	require.Equal(t, clocktypes.DefaultMaxContracts, clockGenesis.Params.MaxContracts)
	require.Equal(t, legacyClock.ContractGasLimit, clockGenesis.Params.ContractGasLimit)
	require.Equal(t, cwhookstypes.DefaultMaxContracts, cwHooksGenesis.Params.MaxContracts)
	require.Equal(t, legacyCWHooks.ContractGasLimit, cwHooksGenesis.Params.ContractGasLimit)
	require.Equal(t, legacyCWHooks.ContractFailureRemovalThreshold, cwHooksGenesis.Params.ContractFailureRemovalThreshold)
}

func TestExportContractCapNormalizationRespectsModuleSelection(t *testing.T) {
	app := setup.Setup(false, t.TempDir(), "juno-export-module-selection-1")
	ctx := app.NewContextLegacy(false, cmtproto.Header{Height: 1})
	ctx.MultiStore().(storetypes.CacheMultiStore).Write()
	_, err := app.Commit()
	require.NoError(t, err)

	_, err = app.ExportAppStateAndValidators(false, nil, []string{"bank"})
	require.NoError(t, err)
}

func TestExportImportPreservesV31State(t *testing.T) {
	const chainID = "juno-export-import-1"
	firstHome := t.TempDir()
	first := setup.Setup(false, firstHome, chainID)
	ctx := first.NewContextLegacy(false, cmtproto.Header{Height: 1, ChainID: chainID})
	_, err := first.AppKeepers.VotingSnapshotKeeper.Params.Get(ctx)
	require.NoError(t, err, "default genesis must initialize voting-snapshot before state export")

	contractAddr := "juno1qsrercqegvs4ye0yqg93knv73ye5dc3prqwd6jcdcuj8ggp6w0us66deup"
	walletAddr := "juno1p30mp2fh2p6603h9mkxc8alw6wplss72dfd385"
	contract := feepaytypes.FeePayContract{
		ContractAddress: contractAddr,
		Balance:         1_000_000,
		WalletLimit:     10,
	}
	backing := sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000))
	require.NoError(t, first.AppKeepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, backing))
	require.NoError(t, first.AppKeepers.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, feepaytypes.ModuleName, backing))
	first.AppKeepers.FeePayKeeper.SetFeePayContract(ctx, contract)
	require.NoError(t, first.AppKeepers.FeePayKeeper.IncrementContractUses(ctx, &contract, walletAddr, 3))
	preExportPower, err := first.AppKeepers.VotingSnapshotKeeper.TotalVotingPowerAt(ctx, ctx.BlockHeight())
	require.NoError(t, err)
	require.False(t, preExportPower.IsZero())
	preExportVersions, err := first.AppKeepers.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	ctx.MultiStore().(storetypes.CacheMultiStore).Write()
	_, err = first.Commit()
	require.NoError(t, err)

	exported, err := first.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, exported.AppState)
	require.Positive(t, exported.Height)

	second := setup.Setup(true, t.TempDir(), chainID)
	_, err = second.InitChain(&abci.RequestInitChain{
		ConsensusParams: &exported.ConsensusParams,
		AppStateBytes:   exported.AppState,
		ChainId:         chainID,
		InitialHeight:   exported.Height,
	})
	require.NoError(t, err)

	restoredCtx := second.NewContextLegacy(false, cmtproto.Header{Height: exported.Height, ChainID: chainID})
	restored, err := second.AppKeepers.FeePayKeeper.GetContract(restoredCtx, contractAddr)
	require.NoError(t, err)
	require.Equal(t, contract, *restored)
	uses, err := second.AppKeepers.FeePayKeeper.GetContractUses(restoredCtx, restored, walletAddr)
	require.NoError(t, err)
	require.Equal(t, uint64(3), uses)
	require.Equal(t, backing.AmountOf("stake"), second.AppKeepers.BankKeeper.GetBalance(
		restoredCtx,
		second.AppKeepers.AccountKeeper.GetModuleAddress(feepaytypes.ModuleName),
		"stake",
	).Amount)
	restoredPower, err := second.AppKeepers.VotingSnapshotKeeper.TotalVotingPowerAt(restoredCtx, exported.Height)
	require.NoError(t, err)
	require.Equal(t, preExportPower, restoredPower)
	restoredVersions, err := second.AppKeepers.UpgradeKeeper.GetModuleVersionMap(restoredCtx)
	require.NoError(t, err)
	require.Equal(t, preExportVersions, restoredVersions)
}
