package app_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/CosmosContracts/juno/v31/testutil/setup"
	feepaytypes "github.com/CosmosContracts/juno/v31/x/feepay/types"
)

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
	first.Commit()

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
