package keeper

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

var ten = sdk.MustNewDecFromStr("10")

// Keeper of the oracle store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramstypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistributionKeeper
	StakingKeeper types.StakingKeeper

	distrName string
}

// NewKeeper constructs a new keeper for oracle
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramspace paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distrKeeper types.DistributionKeeper,
	stakingKeeper types.StakingKeeper,
	distrName string,
) Keeper {
	// ensure oracle module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramspace.HasKeyTable() {
		paramspace = paramspace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramspace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
		StakingKeeper: stakingKeeper,
		distrName:     distrName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

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

// GetExchangeRate gets the consensus exchange rate of USD denominated in the
// denom asset from the store.
func (k Keeper) GetExchangeRate(ctx sdk.Context, symbol string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	symbol = strings.ToUpper(symbol)
	b := store.Get(types.GetExchangeRateKey(symbol))
	if b == nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, symbol)
	}

	decProto := sdk.DecProto{}
	k.cdc.MustUnmarshal(b, &decProto)

	return decProto.Dec, nil
}

// GetExchangeRateBase gets the consensus exchange rate of an asset
// in the base denom (e.g. ATOM -> uatom)
func (k Keeper) GetExchangeRateBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	var symbol string
	var exponent uint64
	// Translate the base denom -> symbol
	params := k.GetParams(ctx)
	for _, listDenom := range params.AcceptList {
		if listDenom.BaseDenom == denom {
			symbol = listDenom.SymbolDenom
			exponent = uint64(listDenom.Exponent)
			break
		}
	}
	if len(symbol) == 0 {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	exchangeRate, err := k.GetExchangeRate(ctx, symbol)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	powerReduction := ten.Power(exponent)
	return exchangeRate.Quo(powerReduction), nil
}

// SetExchangeRate sets the consensus exchange rate of USD denominated in the
// denom asset to the store.
func (k Keeper) SetExchangeRate(ctx sdk.Context, denom string, exchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: exchangeRate})
	denom = strings.ToUpper(denom)
	store.Set(types.GetExchangeRateKey(denom), bz)
}

// SetExchangeRateWithEvent sets an consensus
// exchange rate to the store with ABCI event
func (k Keeper) SetExchangeRateWithEvent(ctx sdk.Context, denom string, exchangeRate sdk.Dec) error {
	k.SetExchangeRate(ctx, denom, exchangeRate)
	return ctx.EventManager().EmitTypedEvent(&types.EventSetFxRate{
		Denom: denom, Rate: exchangeRate})
}

// IterateExchangeRates iterates over USD rates in the store.
func (k Keeper) IterateExchangeRates(ctx sdk.Context, handler func(string, sdk.Dec) bool) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixExchangeRate)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		denom := string(key[len(types.KeyPrefixExchangeRate) : len(key)-1])
		dp := sdk.DecProto{}

		k.cdc.MustUnmarshal(iter.Value(), &dp)
		if handler(denom, dp.Dec) {
			break
		}
	}
}

func (k Keeper) ClearExchangeRates(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixExchangeRate)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// GetFeederDelegation gets the account address to which the validator operator
// delegated oracle vote rights.
func (k Keeper) GetFeederDelegation(ctx sdk.Context, vAddr sdk.ValAddress) (sdk.AccAddress, error) {
	// check that the given validator exists
	if val := k.StakingKeeper.Validator(ctx, vAddr); val == nil || !val.IsBonded() {
		return nil, stakingtypes.ErrNoValidatorFound.Wrapf("validator %s is not in active set", vAddr)
	}

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFeederDelegationKey(vAddr))
	if bz == nil {
		// no delegation, so validator itself must provide price feed
		return sdk.AccAddress(vAddr), nil
	}
	return sdk.AccAddress(bz), nil
}

// SetFeederDelegation sets the account address to which the validator operator
// delegated oracle vote rights.
func (k Keeper) SetFeederDelegation(ctx sdk.Context, operator sdk.ValAddress, delegatedFeeder sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetFeederDelegationKey(operator), delegatedFeeder.Bytes())
}

type IterateFeederDelegationHandler func(delegator sdk.ValAddress, delegate sdk.AccAddress) (stop bool)

// IterateFeederDelegations iterates over the feed delegates and performs a
// callback function.
func (k Keeper) IterateFeederDelegations(ctx sdk.Context, handler IterateFeederDelegationHandler) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixFeederDelegation)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		delegator := sdk.ValAddress(iter.Key()[2:])
		delegate := sdk.AccAddress(iter.Value())

		if handler(delegator, delegate) {
			break
		}
	}
}

// GetMissCounter retrieves the # of vote periods missed in this oracle slash
// window.
func (k Keeper) GetMissCounter(ctx sdk.Context, operator sdk.ValAddress) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMissCounterKey(operator))
	if bz == nil {
		// by default the counter is zero
		return 0
	}

	var missCounter gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &missCounter)

	return missCounter.Value
}

// SetMissCounter updates the # of vote periods missed in this oracle slash
// window.
func (k Keeper) SetMissCounter(ctx sdk.Context, operator sdk.ValAddress, missCounter uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: missCounter})
	store.Set(types.GetMissCounterKey(operator), bz)
}

// DeleteMissCounter removes miss counter for the validator.
func (k Keeper) DeleteMissCounter(ctx sdk.Context, operator sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetMissCounterKey(operator))
}

// IterateMissCounters iterates over the miss counters and performs a callback
// function.
func (k Keeper) IterateMissCounters(ctx sdk.Context, handler func(sdk.ValAddress, uint64) bool) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixMissCounter)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		operator := sdk.ValAddress(iter.Key()[2:])

		var missCounter gogotypes.UInt64Value
		k.cdc.MustUnmarshal(iter.Value(), &missCounter)

		if handler(operator, missCounter.Value) {
			break
		}
	}
}

// GetAggregateExchangeRatePrevote retrieves an oracle prevote from the store.
func (k Keeper) GetAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
) (types.AggregateExchangeRatePrevote, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetAggregateExchangeRatePrevoteKey(voter))
	if bz == nil {
		return types.AggregateExchangeRatePrevote{}, sdkerrors.Wrap(types.ErrNoAggregatePrevote, voter.String())
	}

	var aggregatePrevote types.AggregateExchangeRatePrevote
	k.cdc.MustUnmarshal(bz, &aggregatePrevote)

	return aggregatePrevote, nil
}

// HasAggregateExchangeRatePrevote checks if a validator has an existing prevote.
func (k Keeper) HasAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetAggregateExchangeRatePrevoteKey(voter))
}

// SetAggregateExchangeRatePrevote set an oracle aggregate prevote to the store.
func (k Keeper) SetAggregateExchangeRatePrevote(
	ctx sdk.Context,
	voter sdk.ValAddress,
	prevote types.AggregateExchangeRatePrevote,
) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&prevote)
	store.Set(types.GetAggregateExchangeRatePrevoteKey(voter), bz)
}

// DeleteAggregateExchangeRatePrevote deletes an oracle prevote from the store.
func (k Keeper) DeleteAggregateExchangeRatePrevote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAggregateExchangeRatePrevoteKey(voter))
}

// IterateAggregateExchangeRatePrevotes iterates rate over prevotes in the store
func (k Keeper) IterateAggregateExchangeRatePrevotes(
	ctx sdk.Context,
	handler func(sdk.ValAddress, types.AggregateExchangeRatePrevote) bool,
) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixAggregateExchangeRatePrevote)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		voterAddr := sdk.ValAddress(iter.Key()[2:])

		var aggregatePrevote types.AggregateExchangeRatePrevote
		k.cdc.MustUnmarshal(iter.Value(), &aggregatePrevote)

		if handler(voterAddr, aggregatePrevote) {
			break
		}
	}
}

// GetAggregateExchangeRateVote retrieves an oracle prevote from the store.
func (k Keeper) GetAggregateExchangeRateVote(
	ctx sdk.Context,
	voter sdk.ValAddress,
) (types.AggregateExchangeRateVote, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetAggregateExchangeRateVoteKey(voter))
	if bz == nil {
		return types.AggregateExchangeRateVote{}, sdkerrors.Wrap(types.ErrNoAggregateVote, voter.String())
	}

	var aggregateVote types.AggregateExchangeRateVote
	k.cdc.MustUnmarshal(bz, &aggregateVote)

	return aggregateVote, nil
}

// SetAggregateExchangeRateVote adds an oracle aggregate prevote to the store.
func (k Keeper) SetAggregateExchangeRateVote(
	ctx sdk.Context,
	voter sdk.ValAddress,
	vote types.AggregateExchangeRateVote,
) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&vote)
	store.Set(types.GetAggregateExchangeRateVoteKey(voter), bz)
}

// DeleteAggregateExchangeRateVote deletes an oracle prevote from the store.
func (k Keeper) DeleteAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAggregateExchangeRateVoteKey(voter))
}

type IterateExchangeRateVote = func(
	voterAddr sdk.ValAddress,
	aggregateVote types.AggregateExchangeRateVote,
) (stop bool)

// IterateAggregateExchangeRateVotes iterates rate over prevotes in the store.
func (k Keeper) IterateAggregateExchangeRateVotes(
	ctx sdk.Context,
	handler IterateExchangeRateVote,
) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixAggregateExchangeRateVote)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		voterAddr := sdk.ValAddress(iter.Key()[2:])

		var aggregateVote types.AggregateExchangeRateVote
		k.cdc.MustUnmarshal(iter.Value(), &aggregateVote)

		if handler(voterAddr, aggregateVote) {
			break
		}
	}
}

// ValidateFeeder returns error if the given feeder is not allowed to feed the message.
func (k Keeper) ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	delegate, err := k.GetFeederDelegation(ctx, valAddr)
	if err != nil {
		return err
	}
	if !delegate.Equals(feederAddr) {
		return sdkerrors.Wrap(types.ErrNoVotingPermission, feederAddr.String())
	}

	return nil
}

func (k Keeper) isInTrackingList(ctx sdk.Context, symbolDenom string) (bool, types.Denom) {
	var denom types.Denom
	upperSymbolDenom := strings.ToUpper(symbolDenom)
	params := k.GetParams(ctx)
	for _, trackingDenom := range params.PriceTrackingList {
		if trackingDenom.SymbolDenom == upperSymbolDenom {
			denom = trackingDenom
			return true, denom
		}
	}

	return false, denom
}
