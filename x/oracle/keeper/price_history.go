package keeper

import (
	"strings"
	"time"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Set Price history
func (k Keeper) SetDenomPriceHistory(ctx sdk.Context, symbolDenom string, exchangeRate sdk.Dec, time time.Time, blockHeight uint64) error {
	// Check if not in tracking list
	upperSymbolDenom := strings.ToUpper(symbolDenom)
	found, _ := k.isInTrackingList(ctx, upperSymbolDenom)
	if !found {
		// if not in tracking list, doing nothing => just return nil
		return nil
	}
	// Calculate voting Period count
	params := k.GetParams(ctx)
	votingPeriodCount := (blockHeight + 1) / params.VotePeriod
	if votingPeriodCount == 0 {
		return sdkerrors.Wrap(types.ErrInvalidVotePeriod, "Voting period must be positive")
	}

	// Get store
	store := ctx.KVStore(k.storeKey)
	priceHistoryStore := prefix.NewStore(store, types.GetPriceHistoryKey(upperSymbolDenom))

	// Store data to store
	priceHistoryEntry := &types.PriceHistoryEntry{
		Price:           exchangeRate,
		VotePeriodCount: votingPeriodCount,
		PriceUpdateTime: time,
	}
	bz, err := k.cdc.Marshal(priceHistoryEntry)
	if err != nil {
		return err
	}
	key := sdk.Uint64ToBigEndian(votingPeriodCount)
	priceHistoryStore.Set(key, bz)

	return nil
}

// Get History Price from symbol denom
func (k Keeper) GetDenomPriceHistoryWithBlockHeight(ctx sdk.Context, symbolDenom string, blockHeight uint64) (types.PriceHistoryEntry, error) {
	var priceHistoryEntry types.PriceHistoryEntry
	// Check if in tracking list
	upperSymbolDenom := strings.ToUpper(symbolDenom)
	found, _ := k.isInTrackingList(ctx, upperSymbolDenom)
	if !found {
		return priceHistoryEntry, sdkerrors.Wrapf(types.ErrUnknownDenom, "denom %s not in tracking list", upperSymbolDenom)
	}

	// Calculate votingPeriodCount
	params := k.GetParams(ctx)
	votingPeriodCount := (uint64)(blockHeight) / params.VotePeriod
	if votingPeriodCount == 0 {
		return priceHistoryEntry, sdkerrors.Wrap(types.ErrInvalidVotePeriod, "Voting period must be positive")
	}

	// Get store
	store := ctx.KVStore(k.storeKey)
	priceHistoryStore := prefix.NewStore(store, types.GetPriceHistoryKey(upperSymbolDenom))
	// Get data from store
	key := sdk.Uint64ToBigEndian(votingPeriodCount)
	bz := priceHistoryStore.Get(key)
	if bz == nil {
		return priceHistoryEntry, sdkerrors.Wrapf(types.ErrInvalidVotePeriod, "Voting period have no exchange price %d", votingPeriodCount)
	}
	k.cdc.MustUnmarshal(bz, &priceHistoryEntry)

	return priceHistoryEntry, nil
}

// Iterate over history price
func (k Keeper) IterateDenomPriceHistory(ctx sdk.Context, symbolDenom string, cb func(uint64, types.PriceHistoryEntry) bool) {
	// Get store
	upperSymbolDenom := strings.ToUpper(symbolDenom)
	store := ctx.KVStore(k.storeKey)
	priceHistoryStore := prefix.NewStore(store, types.GetPriceHistoryKey(upperSymbolDenom))
	iter := sdk.KVStorePrefixIterator(priceHistoryStore, []byte{})

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var priceHistoryEntry types.PriceHistoryEntry
		k.cdc.MustUnmarshal(iter.Value(), &priceHistoryEntry)
		if cb(sdk.BigEndianToUint64(iter.Key()), priceHistoryEntry) {
			break
		}
	}
}

// Delete denom history price
func (k Keeper) DeleteDenomPriceHistory(ctx sdk.Context, symbolDenom string, votingPeriodCount uint64) {
	// Get store
	upperSymbolDenom := strings.ToUpper(symbolDenom)
	store := ctx.KVStore(k.storeKey)
	priceHistoryStore := prefix.NewStore(store, types.GetPriceHistoryKey(upperSymbolDenom))
	// Delete
	key := sdk.Uint64ToBigEndian(votingPeriodCount)
	priceHistoryStore.Delete(key)
}

// appendPriceHistory
func (k Keeper) appendPriceHistory(ctx sdk.Context, symbolDenom string, priceHistoryEntrys ...types.PriceHistoryEntry) error {
	// Get store
	upperSymbolDenom := strings.ToUpper(symbolDenom)
	store := ctx.KVStore(k.storeKey)
	priceHistoryStore := prefix.NewStore(store, types.GetPriceHistoryKey(upperSymbolDenom))

	for _, priceHistoryEntry := range priceHistoryEntrys {
		key := sdk.Uint64ToBigEndian(priceHistoryEntry.VotePeriodCount)
		bz, err := k.cdc.Marshal(&priceHistoryEntry)
		if err != nil {
			return err
		}
		priceHistoryStore.Set(key, bz)
	}

	return nil
}
