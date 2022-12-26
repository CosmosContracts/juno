package keeper

import (
	"errors"
	"fmt"
	"time"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	"github.com/CosmosContracts/juno/v12/x/oracle/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// storeHistorical data writes to the store, in all needed indexing.
// TODO : testing
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
	// reverseIterator not catch end key => Need this scope to catch if the value is in end key
	key := types.FormatHistoricalDenomIndexKey(t, denom)
	bz := store.Get(key)
	if bz != nil {
		var entry types.PriceHistoryEntry
		err := k.cdc.Unmarshal(bz, &entry)
		if err != nil {
			return types.PriceHistoryEntry{}, err
		}
		return entry, nil
	}

	startKey := types.FormatHistoricalDenomIndexPrefix(denom)
	endKey := types.FormatHistoricalDenomIndexKey(t, denom)
	reverseIterate := true

	entry, err := util.GetFirstValueInRange(store, startKey, endKey, reverseIterate, k.ParseTwapFromBz)
	if err != nil {
		return types.PriceHistoryEntry{}, err
	}

	return entry, nil
}

// getHistoryEntryBetweenTime on a given input (denom, t)
// returns the PriceHistoryEntry from state for (denom, t'),
func (k Keeper) getHistoryEntryBetweenTime(ctx sdk.Context, denom string, start time.Time, end time.Time) (entries []types.PriceHistoryEntry, err error) {
	store := ctx.KVStore(k.storeKey)

	startKey := types.FormatHistoricalDenomIndexKey(start, denom)
	endKey := types.FormatHistoricalDenomIndexKey(end, denom)

	reverseIterate := false

	entries, err = util.GetValueInRange(store, startKey, endKey, reverseIterate, k.ParseTwapFromBz)

	if err != nil {
		return []types.PriceHistoryEntry{}, err
	}

	// Check if the end have entry
	key := types.FormatHistoricalDenomIndexKey(end, denom)
	bz := store.Get(key)
	if bz != nil {
		var entry types.PriceHistoryEntry
		err := k.cdc.Unmarshal(bz, &entry)
		if err != nil {
			return entries, err
		}
		entries = append(entries, entry)
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
		return sdk.Dec{}, err
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

	total := sdk.ZeroDec()
	for i := 0; i < len(allEntries)-1; i++ {
		fl64TW := allEntries[i+1].PriceUpdateTime.Sub(allEntries[i].PriceUpdateTime).Seconds()
		decTW, err := sdk.NewDecFromStr(fmt.Sprintf("%f", fl64TW))
		if err != nil {
			return sdk.Dec{}, nil
		}
		total = total.Add(allEntries[i].Price.Mul(decTW))
	}

	fl64TotalTW := endEntry.PriceUpdateTime.Sub(startEntry.PriceUpdateTime).Seconds()
	decTotalTW, err := sdk.NewDecFromStr(fmt.Sprintf("%f", fl64TotalTW))
	if err != nil {
		return sdk.Dec{}, err
	}

	twapPrice := total.Quo(decTotalTW)

	return twapPrice, nil
}
