package keeper

import (
	"context"
	"sort"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BackfillFromStaking writes a snapshot of every active delegator's bonded
// power and the total voting power at the current block height. Used in the
// v30 upgrade handler so contracts querying VotingPowerAt(height >= upgrade)
// get real data immediately rather than zeros until the next delegation event.
//
// Walks all delegations once to collect the unique delegator set, then
// resolves each delegator's bonded-validator power (same basis as the
// hook-driven snapshots — see delegatorBondedPower). Linear in the number
// of delegations — acceptable for a one-shot upgrade migration.
func (k Keeper) BackfillFromStaking(ctx context.Context) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()

	totals := map[string]math.Int{}
	err := k.stakingKeeper.IterateAllDelegations(ctx, func(d stakingtypes.Delegation) bool {
		// Collect unique delegator addresses here; power is resolved per
		// delegator below via delegatorBondedPower.
		totals[d.DelegatorAddress] = math.ZeroInt() // sentinel; resolved below
		return false
	})
	if err != nil {
		return err
	}

	delegators := make([]string, 0, len(totals))
	for delStr := range totals {
		delegators = append(delegators, delStr)
	}
	sort.Strings(delegators)

	for _, delStr := range delegators {
		addr, err := sdk.AccAddressFromBech32(delStr)
		if err != nil {
			return err
		}

		isLST, err := k.IsLST(ctx, addr)
		if err != nil {
			return err
		}

		var power math.Int
		if isLST {
			power = math.ZeroInt()
		} else {
			power, err = k.delegatorBondedPower(ctx, addr)
			if err != nil {
				return err
			}
		}

		if err := k.VotingPower.Set(ctx, collections.Join[[]byte, int64](addr.Bytes(), height), power); err != nil {
			return err
		}
	}

	return k.recordTotal(ctx)
}
