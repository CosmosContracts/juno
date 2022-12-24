package keeper

import (
	"errors"
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
// TODO : testing
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

// getHistoryEntryAtOrAfterTime on a given input (denom, t)
// returns the PriceHistoryEntry from state for (denom, t'),
// where t' is such that:
// * t' => t
// * there exists no `t” => t` in state, where `t' > t”`
// TODO : testing
func (k Keeper) getHistoryEntryAtOrAfterTime(ctx sdk.Context, denom string, t time.Time) (types.PriceHistoryEntry, error) {
	store := ctx.KVStore(k.storeKey)

	startKey := types.FormatHistoricalDenomIndexKey(t, denom)
	endKey := types.FormatHistoricalDenomIndexPrefix(denom)
	reverseIterate := true

	entry, err := util.GetFirstValueInRange(store, startKey, endKey, reverseIterate, k.ParseTwapFromBz)

	if err != nil {
		return types.PriceHistoryEntry{}, err

	}

	return entry, nil
}

// getHistoryEntryBetweenTime on a given input (denom, t)
// returns the PriceHistoryEntry from state for (denom, t'),
// TODO : testing
func (k Keeper) getHistoryEntryBetweenTime(ctx sdk.Context, denom string, start time.Time, end time.Time) (entries []types.PriceHistoryEntry, err error) {
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

// TODO : testing
func (k Keeper) RemoveHistoryEntryAtOrBeforeTime(ctx sdk.Context, denom string, t time.Time) {
	store := ctx.KVStore(k.storeKey)

	startKey := types.FormatHistoricalDenomIndexPrefix(denom)
	endKey := types.FormatHistoricalDenomIndexKey(t, denom)
	reverseIterate := true

	util.RemoveValueInRange(store, startKey, endKey, reverseIterate)
}

// TODO : testing
func (k Keeper) SetPriceHistoryEntry(ctx sdk.Context, denom string, t time.Time, exchangeRate sdk.Dec, votingPeriodCount uint64) {
	entry := types.PriceHistoryEntry{
		Price:           exchangeRate,
		VotePeriodCount: votingPeriodCount,
		PriceUpdateTime: t,
	}

	k.storeHistoricalData(ctx, denom, entry)
}

func (k Keeper) GetArithmetricTWAP(
	ctx sdk.Context,
	poolID uint64,
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

// TODO: complete this
func (k Keeper) calculateTWAP(starEntry types.PriceHistoryEntry, entries []types.PriceHistoryEntry, endEntry types.PriceHistoryEntry) (sdk.Dec, error) {
	var allEntries []types.PriceHistoryEntry
	allEntries = append(allEntries, starEntry)
	allEntries = append(allEntries, entries...)
	allEntries = append(allEntries, endEntry)

	return sdk.Dec{}, nil
}
