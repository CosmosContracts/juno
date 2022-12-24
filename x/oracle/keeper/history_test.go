package keeper

import (
	"testing"
	"time"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func Test_storeHistoricalData(t *testing.T) {
	timeNow := time.Now().UTC()
	ctx, keepers := CreateTestInput(t, false)
	oracleKeeper := keepers.OracleKeeper

	phEntrys := []types.PriceHistoryEntry{
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 10,
			PriceUpdateTime: timeNow,
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 11,
			PriceUpdateTime: timeNow.Add(time.Minute * 2),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 12,
			PriceUpdateTime: timeNow.Add(time.Minute * 4),
		},
	}

	for _, phEntry := range phEntrys {
		oracleKeeper.storeHistoricalData(ctx, "Denom", phEntry)
	}

	phStore, err := oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrys[0].PriceUpdateTime.Add(-time.Minute))
	require.Error(t, err)
	require.Equal(t, types.PriceHistoryEntry{}, phStore)
	phStore, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrys[0].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrys[0], phStore)
	phStore, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrys[0].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrys[0], phStore)
	phStore, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrys[1].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrys[1], phStore)
	phStore, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrys[1].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrys[1], phStore)

	phStores, err := oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrys[0].PriceUpdateTime.Add(-time.Minute),
		phEntrys[2].PriceUpdateTime.Add(time.Minute),
	)
	require.NoError(t, err)
	require.Equal(t, phStores, phEntrys)

	phStores, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrys[0].PriceUpdateTime,
		phEntrys[2].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, phStores, phEntrys)
}