package keeper

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CosmosContracts/juno/v13/x/oracle/types"
	"github.com/CosmosContracts/juno/v13/x/oracle/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// storeHistorical data writes to the store, in all needed indexing.
func (k Keeper) storeHistoricalData(ctx sdk.Context, denom string, entry types.PriceHistoryEntry) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatHistoricalDenomIndexKey(entry.PriceUpdateTime, denom)

	bz, err := k.cdc.Marshal(&entry)
	if err != nil {
		panic(err)
	}

	store.Set(key, bz)
}

// getHistoryEntryAtOrBeforeTime on a given input (denom, t)
// returns the PriceHistoryEntry from state for (denom, t'),
// where t' is such that:
// * t' <= t
// * there exists no `t” <= t` in state, where `t' < t”`
func (k Keeper) getHistoryEntryAtOrBeforeTime(ctx sdk.Context, denom string, t time.Time) (types.PriceHistoryEntry, error) {
	store := ctx.KVStore(k.storeKey)

	startKey := types.FormatHistoricalDenomIndexPrefix(denom)
	endKey := types.FormatHistoricalDenomIndexKey(t, denom)
	reverseIterate := true

	entry, err := util.GetFirstValueInRange(store, startKey, endKey, reverseIterate, k.ParseTwapFromBz)
	if err != nil {
		return types.PriceHistoryEntry{}, err
	}

	return entry, nil
}

// getFirstEntryBetweenTime on a given input (denom, start, end)
// returns the first PriceHistoryEntry from state for (denom, t'),
// where start <= t' and t' <= end
func (k Keeper) getFirstEntryBetweenTime(ctx sdk.Context, denom string, start time.Time, end time.Time) (entry types.PriceHistoryEntry, err error) {
	if start.After(end) {
		return types.PriceHistoryEntry{}, errors.New("start time after end time")
	}
	store := ctx.KVStore(k.storeKey)

	startKey := types.FormatHistoricalDenomIndexKey(start, denom)
	endKey := types.FormatHistoricalDenomIndexKey(end, denom)

	reverseIterate := false

	entry, err = util.GetFirstValueInRange(store, startKey, endKey, reverseIterate, k.ParseTwapFromBz)

	if err != nil {
		return types.PriceHistoryEntry{}, err
	}

	return entry, nil
}

// getHistoryEntryBetweenTime on a given input (denom, start, end)
// returns all PriceHistoryEntry values from state for (denom, t'),
// where start <= t' and t' <= end
func (k Keeper) getHistoryEntryBetweenTime(ctx sdk.Context, denom string, start time.Time, end time.Time) (entries []types.PriceHistoryEntry, err error) {
	if start.After(end) {
		return []types.PriceHistoryEntry{}, errors.New("start time after end time")
	}
	store := ctx.KVStore(k.storeKey)

	startKey := types.FormatHistoricalDenomIndexKey(start, denom)
	endKey := types.FormatHistoricalDenomIndexKey(end, denom)

	reverseIterate := false

	entries, err = util.GetValueInRange(store, startKey, endKey, reverseIterate, k.ParseTwapFromBz)

	if err != nil {
		return []types.PriceHistoryEntry{}, err
	}

	return entries, nil
}

func (k Keeper) ParseTwapFromBz(bz []byte) (entry types.PriceHistoryEntry, err error) {
	if len(bz) == 0 {
		return types.PriceHistoryEntry{}, errors.New("history entry not found")
	}
	err = k.cdc.Unmarshal(bz, &entry)
	return entry, err
}

// RemoveHistoryEntryBeforeTime remove all history entry
// that had UpdatePriceTime before t
func (k Keeper) RemoveHistoryEntryBeforeTime(ctx sdk.Context, denom string, t time.Time) {
	store := ctx.KVStore(k.storeKey)

	startKey := types.FormatHistoricalDenomIndexPrefix(denom)
	endKey := types.FormatHistoricalDenomIndexKey(t, denom)
	reverseIterate := true

	util.RemoveValueInRange(store, startKey, endKey, reverseIterate)
}

func (k Keeper) SetPriceHistoryEntry(ctx sdk.Context, denom string, t time.Time, exchangeRate sdk.Dec, votingPeriodCount uint64) {
	entry := types.PriceHistoryEntry{
		Price:           exchangeRate,
		VotePeriodCount: votingPeriodCount,
		PriceUpdateTime: t,
	}

	k.storeHistoricalData(ctx, denom, entry)
}

// GetArithmetricTWAP get the arithmetic TWAP
// of specific denom between startTime and endTime
func (k Keeper) GetArithmetricTWAP(
	ctx sdk.Context,
	denom string,
	startTime time.Time,
	endTime time.Time,
) (sdk.Dec, error) {
	startEntry, err := k.getHistoryEntryAtOrBeforeTime(ctx, denom, startTime)
	if err != nil {
		if strings.Contains(err.Error(), "no values in range") {
			// pick value right after startTime if no value before startTime
			startEntryAfter, error := k.getFirstEntryBetweenTime(ctx, denom, startTime, endTime)
			if error != nil {
				return sdk.Dec{}, error
			}
			startEntry = startEntryAfter
		} else {
			return sdk.Dec{}, err
		}
	}
	endEntry, err := k.getHistoryEntryAtOrBeforeTime(ctx, denom, endTime)
	if err != nil {
		return sdk.Dec{}, err
	}

	startEntry.PriceUpdateTime = startTime
	endEntry.PriceUpdateTime = endTime

	entryList, err := k.getHistoryEntryBetweenTime(ctx, denom, startTime, endTime)
	if err != nil {
		return sdk.Dec{}, err
	}

	twapPrice, err := k.calculateTWAP(startEntry, entryList, endEntry)
	if err != nil {
		return sdk.Dec{}, err
	}

	return twapPrice, nil
}

// calculateTWAP calculate TWAP between startEntry and endEntry
func (k Keeper) calculateTWAP(startEntry types.PriceHistoryEntry, entries []types.PriceHistoryEntry, endEntry types.PriceHistoryEntry) (sdk.Dec, error) {
	var allEntries []types.PriceHistoryEntry
	allEntries = append(allEntries, startEntry)
	allEntries = append(allEntries, entries...)
	allEntries = append(allEntries, endEntry)

	// Calculate the total time weight multiple with price
	total := sdk.ZeroDec()
	for i := 0; i < len(allEntries)-1; i++ {
		fl64TW := allEntries[i+1].PriceUpdateTime.Sub(allEntries[i].PriceUpdateTime).Seconds()
		decTW, err := sdk.NewDecFromStr(fmt.Sprintf("%f", fl64TW))
		if err != nil {
			return sdk.Dec{}, nil
		}
		total = total.Add(allEntries[i].Price.Mul(decTW))
	}

	// Calculate the time weight average price
	fl64TotalTW := endEntry.PriceUpdateTime.Sub(startEntry.PriceUpdateTime).Seconds()
	decTotalTW, err := sdk.NewDecFromStr(fmt.Sprintf("%f", fl64TotalTW))
	if err != nil {
		return sdk.Dec{}, err
	}
	twapPrice := total.Quo(decTotalTW)

	return twapPrice, nil
}
