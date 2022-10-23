package keeper

import (
	"sort"

	"github.com/CosmosContracts/juno/v11/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OrganizeBallotByDenom collects all oracle votes for the current vote period,
// categorized by the votes' denom parameter.
func (k Keeper) OrganizeBallotByDenom(
	ctx sdk.Context,
	validatorClaimMap map[string]types.Claim,
) []types.BallotDenom {
	votes := map[string]types.ExchangeRateBallot{}

	// collect aggregate votes
	aggregateHandler := func(voterAddr sdk.ValAddress, vote types.AggregateExchangeRateVote) bool {
		// organize ballot only for the active validators
		claim, ok := validatorClaimMap[vote.Voter]
		if ok {
			power := claim.Power

			for _, tuple := range vote.ExchangeRateTuples {
				tmpPower := power

				votes[tuple.Denom] = append(
					votes[tuple.Denom],
					types.NewVoteForTally(tuple.ExchangeRate, tuple.Denom, voterAddr, tmpPower),
				)
			}
		}

		return false
	}

	k.IterateAggregateExchangeRateVotes(ctx, aggregateHandler)

	// sort created ballots
	for denom, ballot := range votes {
		sort.Sort(ballot)
		votes[denom] = ballot
	}
	return types.BallotMapToSlice(votes)
}

// ClearBallots clears all tallied prevotes and votes from the store.
func (k Keeper) ClearBallots(ctx sdk.Context, votePeriod uint64) {
	// clear all aggregate prevotes
	k.IterateAggregateExchangeRatePrevotes(
		ctx,
		func(voterAddr sdk.ValAddress, aggPrevote types.AggregateExchangeRatePrevote) bool {
			if ctx.BlockHeight() > int64(aggPrevote.SubmitBlock+votePeriod) {
				k.DeleteAggregateExchangeRatePrevote(ctx, voterAddr)
			}

			return false
		},
	)

	// clear all aggregate votes
	k.IterateAggregateExchangeRateVotes(
		ctx,
		func(voterAddr sdk.ValAddress, _ types.AggregateExchangeRateVote) bool {
			k.DeleteAggregateExchangeRateVote(ctx, voterAddr)
			return false
		},
	)
}
