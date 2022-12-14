package keeper

import (
	"fmt"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the x/oracle module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	for _, d := range genState.FeederDelegations {
		voter, err := sdk.ValAddressFromBech32(d.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		feeder, err := sdk.AccAddressFromBech32(d.FeederAddress)
		if err != nil {
			panic(err)
		}

		k.SetFeederDelegation(ctx, voter, feeder)
	}

	for _, ex := range genState.ExchangeRates {
		k.SetExchangeRate(ctx, ex.Denom, ex.ExchangeRate)
	}

	for _, mc := range genState.MissCounters {
		operator, err := sdk.ValAddressFromBech32(mc.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		k.SetMissCounter(ctx, operator, mc.MissCounter)
	}

	for _, ap := range genState.AggregateExchangeRatePrevotes {
		valAddr, err := sdk.ValAddressFromBech32(ap.Voter)
		if err != nil {
			panic(err)
		}

		k.SetAggregateExchangeRatePrevote(ctx, valAddr, ap)
	}

	for _, av := range genState.AggregateExchangeRateVotes {
		valAddr, err := sdk.ValAddressFromBech32(av.Voter)
		if err != nil {
			panic(err)
		}

		k.SetAggregateExchangeRateVote(ctx, valAddr, av)
	}

	for _, historyPrice := range genState.PriceHistory {
		err := k.appendPriceHistory(ctx, historyPrice.Denom.SymbolDenom, historyPrice.PriceHistoryEntry...)
		if err != nil {
			panic(err)
		}
	}

	k.SetParams(ctx, genState.Params)

	// check if the module account exists
	moduleAcc := k.GetOracleAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
}

// ExportGenesis returns the x/oracle module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	feederDelegations := []types.FeederDelegation{}
	k.IterateFeederDelegations(ctx, func(valAddr sdk.ValAddress, feederAddr sdk.AccAddress) (stop bool) {
		feederDelegations = append(feederDelegations, types.FeederDelegation{
			ValidatorAddress: valAddr.String(),
			FeederAddress:    feederAddr.String(),
		})

		return false
	})

	exchangeRates := []types.ExchangeRateTuple{}
	k.IterateExchangeRates(ctx, func(denom string, rate sdk.Dec) (stop bool) {
		exchangeRates = append(exchangeRates, types.ExchangeRateTuple{
			Denom:        denom,
			ExchangeRate: rate,
		})

		return false
	})

	missCounters := []types.MissCounter{}
	k.IterateMissCounters(ctx, func(operator sdk.ValAddress, missCounter uint64) (stop bool) {
		missCounters = append(missCounters, types.MissCounter{
			ValidatorAddress: operator.String(),
			MissCounter:      missCounter,
		})

		return false
	})

	aggregateExchangeRatePrevotes := []types.AggregateExchangeRatePrevote{}
	k.IterateAggregateExchangeRatePrevotes(
		ctx,
		func(_ sdk.ValAddress, aggregatePrevote types.AggregateExchangeRatePrevote) (stop bool) {
			aggregateExchangeRatePrevotes = append(aggregateExchangeRatePrevotes, aggregatePrevote)
			return false
		},
	)

	aggregateExchangeRateVotes := []types.AggregateExchangeRateVote{}
	k.IterateAggregateExchangeRateVotes(
		ctx,
		func(_ sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) bool {
			aggregateExchangeRateVotes = append(aggregateExchangeRateVotes, aggregateVote)
			return false
		},
	)

	priceHistory := []types.PriceHistory{}
	for _, denom := range params.PriceTrackingList {
		// Get price history entry list
		var priceHistoryEntryLists []types.PriceHistoryEntry
		k.IterateDenomPriceHistory(ctx, denom.SymbolDenom, func(votingPeriodCount uint64, priceHistoryEntry types.PriceHistoryEntry) bool {
			priceHistoryEntryLists = append(priceHistoryEntryLists, priceHistoryEntry)
			return false
		})
		// Add to priceHistory
		denomPriceHistory := types.PriceHistory{
			Denom:             denom,
			PriceHistoryEntry: priceHistoryEntryLists,
		}
		priceHistory = append(priceHistory, denomPriceHistory)
	}

	return types.NewGenesisState(
		params,
		exchangeRates,
		feederDelegations,
		missCounters,
		aggregateExchangeRatePrevotes,
		aggregateExchangeRateVotes,
		priceHistory,
	)
}
