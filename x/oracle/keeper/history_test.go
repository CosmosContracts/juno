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

func Test_getHistoryEntryAtOrBeforeTime(t *testing.T) {
	timeNow := time.Now().UTC()
	ctx, keepers := CreateTestInput(t, false)
	oracleKeeper := keepers.OracleKeeper

	phEntry := types.PriceHistoryEntry{
		Price:           sdk.OneDec(),
		VotePeriodCount: 10,
		PriceUpdateTime: timeNow,
	}
	oracleKeeper.storeHistoricalData(ctx, "Denom", phEntry)

	for _, tc := range []struct {
		desc      string
		denom     string
		timeGet   time.Time
		res       types.PriceHistoryEntry
		shouldErr bool
	}{
		{
			desc:      "Success - timeGet equal PriceUpdateTime",
			denom:     "Denom",
			timeGet:   timeNow,
			res:       phEntry,
			shouldErr: false,
		},
		{
			desc:      "Success - timeGet after PriceUpdateTime",
			denom:     "Denom",
			timeGet:   timeNow.Add(time.Minute),
			res:       phEntry,
			shouldErr: false,
		},
		{
			desc:      "Fail - timeGet before PriceUpdateTime",
			denom:     "Denom",
			timeGet:   timeNow.Add(-time.Minute),
			shouldErr: true,
		},
		{
			desc:      "Fail - Invalid denom",
			denom:     "Invalid",
			timeGet:   timeNow,
			shouldErr: true,
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			if !tc.shouldErr {
				res, err := oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, tc.denom, tc.timeGet)
				require.NoError(t, err)
				require.Equal(t, tc.res, res)
			} else {
				_, err := oracleKeeper.getHistoryEntryAtOrBeforeTime(ctx, tc.denom, tc.timeGet)
				require.Error(t, err)
			}
		})
	}
}

func Test_getHistoryEntryBetweenTime(t *testing.T) {
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

	for _, tc := range []struct {
		desc         string
		denom        string
		startTimeGet time.Time
		endTimeGet   time.Time
		res          []types.PriceHistoryEntry
		shouldErr    bool
	}{
		{
			desc:         "Success - startTime < entry 1 < entry 2 < entry 3 < endTime",
			denom:        "Denom",
			startTimeGet: timeNow.Add(-time.Minute),
			endTimeGet:   timeNow.Add(time.Minute * 5),
			res:          phEntrys,
			shouldErr:    false,
		},
		{
			desc:         "Success - startTime = entry 1 < entry 2 < entry 3 = endTime",
			denom:        "Denom",
			startTimeGet: timeNow,
			endTimeGet:   timeNow.Add(time.Minute * 4),
			res:          phEntrys,
			shouldErr:    false,
		},
		{
			desc:         "Success - entry 1 < startTime < entry 2 < entry 3 < endTime",
			denom:        "Denom",
			startTimeGet: timeNow.Add(time.Minute),
			endTimeGet:   timeNow.Add(time.Minute * 5),
			res: []types.PriceHistoryEntry{
				phEntrys[1],
				phEntrys[2],
			},
			shouldErr: false,
		},
		{
			desc:         "Success - entry 1 < entry 2 < startTime < entry 3 < endTime",
			denom:        "Denom",
			startTimeGet: timeNow.Add(time.Minute * 3),
			endTimeGet:   timeNow.Add(time.Minute * 5),
			res: []types.PriceHistoryEntry{
				phEntrys[2],
			},
			shouldErr: false,
		},
		{
			desc:         "Success - entry 1 < entry 2 < startTime < entry 3 < endTime",
			denom:        "Denom",
			startTimeGet: timeNow.Add(time.Minute * 3),
			endTimeGet:   timeNow.Add(time.Minute * 5),
			res: []types.PriceHistoryEntry{
				phEntrys[2],
			},
			shouldErr: false,
		},
		{
			desc:         "Success - entry 1 < entry 2 < startTime < entry 3 < endTime",
			denom:        "Denom",
			startTimeGet: timeNow.Add(time.Minute * 3),
			endTimeGet:   timeNow.Add(time.Minute * 5),
			res: []types.PriceHistoryEntry{
				phEntrys[2],
			},
			shouldErr: false,
		},
		{
			desc:         "Fail - entry 1 < entry 2 < entry 3 < startTime < endTime - No Value in range",
			denom:        "Denom",
			startTimeGet: timeNow.Add(time.Minute * 5),
			endTimeGet:   timeNow.Add(time.Minute * 6),
			shouldErr:    true,
		},
		{
			desc:         "Fail - Invalid denom",
			denom:        "Invalid",
			startTimeGet: timeNow.Add(-time.Minute),
			endTimeGet:   timeNow.Add(time.Minute * 5),
			shouldErr:    true,
		},
		{
			desc:         "Fail - Invalid startTime after endTime",
			denom:        "Denom",
			startTimeGet: timeNow.Add(time.Minute * 5),
			endTimeGet:   timeNow,
			shouldErr:    true,
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			if !tc.shouldErr {
				res, err := oracleKeeper.getHistoryEntryBetweenTime(ctx, tc.denom, tc.startTimeGet, tc.endTimeGet)
				require.NoError(t, err)
				require.Equal(t, tc.res, res)
			} else {
				_, err := oracleKeeper.getHistoryEntryBetweenTime(ctx, tc.denom, tc.startTimeGet, tc.endTimeGet)
				require.Error(t, err)
			}
		})

	}
}

func TestRemoveHistoryEntryBeforeTime(t *testing.T) {
	timeNow := time.Now().UTC()
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

	for _, tc := range []struct {
		desc           string
		deleteTime     time.Time
		phEntryStorage []types.PriceHistoryEntry
	}{
		{
			desc:           "deleteTime before entryTime",
			deleteTime:     timeNow.Add(-time.Minute),
			phEntryStorage: phEntrys,
		},
		{
			desc:           "deleteTime equal entryTime",
			deleteTime:     timeNow,
			phEntryStorage: phEntrys,
		},
		{
			desc:       "deleteTime after entryTime (1 element - delete phEntrys[0])",
			deleteTime: timeNow.Add(time.Minute),
			phEntryStorage: []types.PriceHistoryEntry{
				phEntrys[1],
				phEntrys[2],
			},
		},
		{
			desc:           "deleteTime after entryTime (all)",
			deleteTime:     timeNow.Add(time.Minute * 5),
			phEntryStorage: []types.PriceHistoryEntry{},
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			ctx, keepers := CreateTestInput(t, false)
			oracleKeeper := keepers.OracleKeeper
			for _, phEntry := range phEntrys {
				oracleKeeper.storeHistoricalData(ctx, "Denom", phEntry)
			}
			oracleKeeper.RemoveHistoryEntryBeforeTime(ctx, "Denom", tc.deleteTime)
			phStores, _ := oracleKeeper.getHistoryEntryBetweenTime(
				ctx,
				"Denom",
				phEntrys[0].PriceUpdateTime,
				phEntrys[2].PriceUpdateTime,
			)
			require.Equal(t, tc.phEntryStorage, phStores)
		})
	}
}
