package oracle

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v12/x/oracle/keeper"
	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

// InitGenesis initializes the x/oracle module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, genState types.GenesisState) {
	for _, d := range genState.FeederDelegations {
		voter, err := sdk.ValAddressFromBech32(d.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		feeder, err := sdk.AccAddressFromBech32(d.FeederAddress)
		if err != nil {
			panic(err)
		}

		keeper.SetFeederDelegation(ctx, voter, feeder)
	}

	for _, ex := range genState.ExchangeRates {
		keeper.SetExchangeRate(ctx, ex.Denom, ex.ExchangeRate)
	}

	for _, mc := range genState.MissCounters {
		operator, err := sdk.ValAddressFromBech32(mc.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		keeper.SetMissCounter(ctx, operator, mc.MissCounter)
	}

	for _, ap := range genState.AggregateExchangeRatePrevotes {
		valAddr, err := sdk.ValAddressFromBech32(ap.Voter)
		if err != nil {
			panic(err)
		}

		keeper.SetAggregateExchangeRatePrevote(ctx, valAddr, ap)
	}

	for _, av := range genState.AggregateExchangeRateVotes {
		valAddr, err := sdk.ValAddressFromBech32(av.Voter)
		if err != nil {
			panic(err)
		}

		keeper.SetAggregateExchangeRateVote(ctx, valAddr, av)
	}

	keeper.SetParams(ctx, genState.Params)

	// check if the module account exists
	moduleAcc := keeper.GetOracleAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
}

// ExportGenesis returns the x/oracle module's exported genesis.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	params := keeper.GetParams(ctx)

	feederDelegations := []types.FeederDelegation{}
	keeper.IterateFeederDelegations(ctx, func(valAddr sdk.ValAddress, feederAddr sdk.AccAddress) (stop bool) {
		feederDelegations = append(feederDelegations, types.FeederDelegation{
			ValidatorAddress: valAddr.String(),
			FeederAddress:    feederAddr.String(),
		})

		return false
	})

	exchangeRates := []types.ExchangeRateTuple{}
	keeper.IterateExchangeRates(ctx, func(denom string, rate sdk.Dec) (stop bool) {
		exchangeRates = append(exchangeRates, types.ExchangeRateTuple{
			Denom:        denom,
			ExchangeRate: rate,
		})

		return false
	})

	missCounters := []types.MissCounter{}
	keeper.IterateMissCounters(ctx, func(operator sdk.ValAddress, missCounter uint64) (stop bool) {
		missCounters = append(missCounters, types.MissCounter{
			ValidatorAddress: operator.String(),
			MissCounter:      missCounter,
		})

		return false
	})

	aggregateExchangeRatePrevotes := []types.AggregateExchangeRatePrevote{}
	keeper.IterateAggregateExchangeRatePrevotes(
		ctx,
		func(_ sdk.ValAddress, aggregatePrevote types.AggregateExchangeRatePrevote) (stop bool) {
			aggregateExchangeRatePrevotes = append(aggregateExchangeRatePrevotes, aggregatePrevote)
			return false
		},
	)

	aggregateExchangeRateVotes := []types.AggregateExchangeRateVote{}
	keeper.IterateAggregateExchangeRateVotes(
		ctx,
		func(_ sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) bool {
			aggregateExchangeRateVotes = append(aggregateExchangeRateVotes, aggregateVote)
			return false
		},
	)

	return types.NewGenesisState(
		params,
		exchangeRates,
		feederDelegations,
		missCounters,
		aggregateExchangeRatePrevotes,
		aggregateExchangeRateVotes,
	)
}
