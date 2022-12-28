package keeper

import (
	"testing"
	"time"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func Test_StoreAndGetHistoricalData(t *testing.T) {
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

	phStores, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrys[0].PriceUpdateTime.Add(-time.Minute),
		phEntrys[1].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(phStores))
	require.Equal(t, phStores[0], phEntrys[0])
	require.Equal(t, phStores[1], phEntrys[1])

	phStores, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrys[0].PriceUpdateTime,
		phEntrys[2].PriceUpdateTime.Add(-time.Minute),
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(phStores))
	require.Equal(t, phStores[0], phEntrys[0])
	require.Equal(t, phStores[1], phEntrys[1])
}

func TestRemoveHistoryEntryBeforeTime(t *testing.T) {
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

	oracleKeeper.RemoveHistoryEntryBeforeTime(ctx, "Denom", phEntrys[0].PriceUpdateTime)
	phStores, err := oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrys[0].PriceUpdateTime,
		phEntrys[2].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, 3, len(phStores))
	require.Equal(t, phStores, phEntrys)

	oracleKeeper.RemoveHistoryEntryBeforeTime(ctx, "Denom", phEntrys[1].PriceUpdateTime)
	phStores, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrys[0].PriceUpdateTime,
		phEntrys[2].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(phStores))
	require.Equal(t, phStores[0], phEntrys[1])
	require.Equal(t, phStores[1], phEntrys[2])

	oracleKeeper.RemoveHistoryEntryBeforeTime(ctx, "Denom", phEntrys[1].PriceUpdateTime.Add(time.Minute))
	phStores, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrys[0].PriceUpdateTime,
		phEntrys[2].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(phStores))
	require.Equal(t, phStores[0], phEntrys[2])
}

func TestStoreAndGetMultipleHistoricalData(t *testing.T) {
	timeNow := time.Now().UTC()
	ctx, keepers := CreateTestInput(t, false)
	oracleKeeper := keepers.OracleKeeper

	phEntrysDenom := []types.PriceHistoryEntry{
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
	for _, phEntry := range phEntrysDenom {
		oracleKeeper.storeHistoricalData(ctx, "Denom", phEntry)
	}

	phEntrysDen := []types.PriceHistoryEntry{
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 100,
			PriceUpdateTime: timeNow,
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 101,
			PriceUpdateTime: timeNow.Add(time.Minute * 2),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 102,
			PriceUpdateTime: timeNow.Add(time.Minute * 6),
		},
	}
	for _, phEntry := range phEntrysDen {
		oracleKeeper.storeHistoricalData(ctx, "Den", phEntry)
	}

	phEntrysJuno := []types.PriceHistoryEntry{
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 20,
			PriceUpdateTime: timeNow,
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 21,
			PriceUpdateTime: timeNow.Add(time.Minute * 3),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 22,
			PriceUpdateTime: timeNow.Add(time.Minute * 9),
		},
	}
	for _, phEntry := range phEntrysJuno {
		oracleKeeper.storeHistoricalData(ctx, "JUNO", phEntry)
	}

	// checks for token with denom: "Denom"
	phStoreDenom, err := oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrysDenom[0].PriceUpdateTime.Add(-time.Minute))
	require.Error(t, err)
	require.Equal(t, types.PriceHistoryEntry{}, phStoreDenom)
	phStoreDenom, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrysDenom[0].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrysDenom[0], phStoreDenom)
	phStoreDenom, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrysDenom[0].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrysDenom[0], phStoreDenom)
	phStoreDenom, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrysDenom[1].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrysDenom[1], phStoreDenom)
	phStoreDenom, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Denom", phEntrysDenom[1].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrysDenom[1], phStoreDenom)

	phStoresDenom, err := oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrysDenom[0].PriceUpdateTime.Add(-time.Minute),
		phEntrysDenom[2].PriceUpdateTime.Add(time.Minute),
	)
	require.NoError(t, err)
	require.Equal(t, phStoresDenom, phEntrysDenom)

	phStoresDenom, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Denom",
		phEntrysDenom[0].PriceUpdateTime.Add(-time.Minute),
		phEntrysDenom[1].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(phStoresDenom))
	require.Equal(t, phStoresDenom[0], phEntrysDenom[0])
	require.Equal(t, phStoresDenom[1], phEntrysDenom[1])

	// checks for token with denom: "Den"
	phStoreDen, err := oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Den", phEntrysDen[0].PriceUpdateTime.Add(-time.Minute))
	require.Error(t, err)
	require.Equal(t, types.PriceHistoryEntry{}, phStoreDen)
	phStoreDen, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Den", phEntrysDen[0].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrysDen[0], phStoreDen)
	phStoreDen, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Den", phEntrysDen[0].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrysDen[0], phStoreDen)
	phStoreDen, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Den", phEntrysDen[1].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrysDen[1], phStoreDen)
	phStoreDen, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Den", phEntrysDen[1].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrysDen[1], phStoreDen)

	phStoresDen, err := oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Den",
		phEntrysDen[0].PriceUpdateTime.Add(-time.Minute),
		phEntrysDen[2].PriceUpdateTime.Add(time.Minute),
	)
	require.NoError(t, err)
	require.Equal(t, phStoresDen, phEntrysDen)

	phStoresDen, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Den",
		phEntrysDen[0].PriceUpdateTime.Add(-time.Minute),
		phEntrysDen[1].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(phStoresDen))
	require.Equal(t, phStoresDen[0], phEntrysDen[0])
	require.Equal(t, phStoresDen[1], phEntrysDen[1])

	// checks for token with denom: "Juno"
	phStoreJuno, err := oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrysJuno[0].PriceUpdateTime.Add(-time.Minute))
	require.Error(t, err)
	require.Equal(t, types.PriceHistoryEntry{}, phStoreJuno)
	phStoreJuno, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrysJuno[0].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrysJuno[0], phStoreJuno)
	phStoreJuno, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrysJuno[0].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrysJuno[0], phStoreJuno)
	phStoreJuno, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrysJuno[1].PriceUpdateTime)
	require.NoError(t, err)
	require.Equal(t, phEntrysJuno[1], phStoreJuno)
	phStoreJuno, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrysJuno[1].PriceUpdateTime.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, phEntrysJuno[1], phStoreJuno)

	phStoresdjuno, err := oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Juno",
		phEntrysJuno[0].PriceUpdateTime.Add(-time.Minute),
		phEntrysJuno[2].PriceUpdateTime.Add(time.Minute),
	)
	require.NoError(t, err)
	require.Equal(t, phStoresdjuno, phEntrysJuno)

	phStoresdjuno, err = oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Juno",
		phEntrysJuno[0].PriceUpdateTime.Add(-time.Minute),
		phEntrysJuno[1].PriceUpdateTime,
	)
	require.NoError(t, err)
	require.Equal(t, 2, len(phStoresdjuno))
	require.Equal(t, phStoresdjuno[0], phEntrysJuno[0])
	require.Equal(t, phStoresdjuno[1], phEntrysJuno[1])
}

func TestStoreAndGetNullHistoricalData(t *testing.T) {
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

	// below queries should all throw error as "Juno" denom does not exist
	phStore, err := oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrys[0].PriceUpdateTime.Add(-time.Minute))
	require.Error(t, err)
	require.Equal(t, types.PriceHistoryEntry{}, phStore)
	phStore, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrys[0].PriceUpdateTime)
	require.Error(t, err)
	require.Equal(t, types.PriceHistoryEntry{}, phStore)
	phStore, err = oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, "Juno", phEntrys[1].PriceUpdateTime)
	require.Error(t, err)
	require.Equal(t, types.PriceHistoryEntry{}, phStore)

	phStores, err := oracleKeeper.getHistoryEntryBetweenTime(
		ctx,
		"Juno",
		phEntrys[0].PriceUpdateTime.Add(-time.Minute),
		phEntrys[2].PriceUpdateTime.Add(time.Minute),
	)
	require.NoError(t, err) // To discuss: should this throw error?
	require.Equal(t, []types.PriceHistoryEntry(nil), phStores)
}
