package keeper

import (
	"context"
	"errors"
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
// Walks all delegations once and caches validator bond status so the upgrade
// block does not do a second per-delegator delegation walk. Writes remain
// sorted by delegator address for deterministic IAVL write order.
func (k Keeper) BackfillFromStaking(ctx context.Context) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()

	powers := map[string]math.LegacyDec{}
	validatorBonded := map[string]bool{}
	bondedValidators := map[string]stakingtypes.Validator{}
	var innerErr error
	err := k.stakingKeeper.IterateAllDelegations(ctx, func(d stakingtypes.Delegation) bool {
		if _, ok := powers[d.DelegatorAddress]; !ok {
			powers[d.DelegatorAddress] = math.LegacyZeroDec()
		}

		bonded, ok := validatorBonded[d.ValidatorAddress]
		if !ok {
			valAddr, err := sdk.ValAddressFromBech32(d.ValidatorAddress)
			if err != nil {
				innerErr = err
				return true
			}
			val, err := k.stakingKeeper.GetValidator(ctx, valAddr)
			if err != nil {
				if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
					validatorBonded[d.ValidatorAddress] = false
					return false
				}
				innerErr = err
				return true
			}
			bonded = val.IsBonded()
			validatorBonded[d.ValidatorAddress] = bonded
			if bonded {
				bondedValidators[d.ValidatorAddress] = val
				powers[d.DelegatorAddress] = powers[d.DelegatorAddress].Add(val.TokensFromSharesTruncated(d.Shares))
			}
			return false
		}
		if bonded {
			val, ok := bondedValidators[d.ValidatorAddress]
			if !ok {
				innerErr = stakingtypes.ErrNoValidatorFound
				return true
			}
			powers[d.DelegatorAddress] = powers[d.DelegatorAddress].Add(val.TokensFromSharesTruncated(d.Shares))
		}
		return false
	})
	if err != nil {
		return err
	}
	if innerErr != nil {
		return innerErr
	}

	delegators := make([]string, 0, len(powers))
	for delStr := range powers {
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
			power = powers[delStr].RoundInt()
		}

		if err := k.VotingPower.Set(ctx, collections.Join[[]byte, int64](addr.Bytes(), height), power); err != nil {
			return err
		}
	}

	return k.recordTotal(ctx)
}
