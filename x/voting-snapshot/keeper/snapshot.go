package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// recordDelegatorPower writes a (delegator, height) snapshot for the
// delegator's current total bonded tokens. Idempotent within a block:
// repeated writes at the same height overwrite. LSTs short-circuit to 0.
func (k Keeper) recordDelegatorPower(ctx context.Context, del sdk.AccAddress) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()

	isLST, err := k.IsLST(ctx, del)
	if err != nil {
		return err
	}
	power := math.ZeroInt()
	if !isLST {
		power, err = k.stakingKeeper.GetDelegatorBonded(ctx, del)
		if err != nil {
			return err
		}
	}
	return k.VotingPower.Set(ctx, collections.Join[[]byte, int64](del.Bytes(), height), power)
}

// recordTotal writes a (height) snapshot of total bonded supply.
// Note: x/staking's TotalBondedTokens includes LST-held delegations.
// LST exclusion happens on the *delegator-side* read path; the
// denominator stays whole. (See planning/05-staking-snapshot.md —
// per-LST subtraction from the denominator is a v30.x refinement.)
func (k Keeper) recordTotal(ctx context.Context) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()
	total, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return err
	}
	return k.TotalPower.Set(ctx, height, total)
}

// VotingPowerAt returns the bonded-token voting power attributable to
// del as of `height`. It returns the most recent snapshot at-or-before
// the requested height. Returns zero if no snapshot exists (height is
// before backfill / module activation).
func (k Keeper) VotingPowerAt(ctx context.Context, del sdk.AccAddress, height int64) (math.Int, error) {
	rng := collections.NewPrefixedPairRange[[]byte, int64](del.Bytes()).
		EndInclusive(height).
		Descending()

	iter, err := k.VotingPower.Iterate(ctx, rng)
	if err != nil {
		return math.ZeroInt(), err
	}
	defer func() { _ = iter.Close() }()

	if !iter.Valid() {
		return math.ZeroInt(), nil
	}
	return iter.Value()
}

// TotalVotingPowerAt returns the most recent total-bonded snapshot
// at-or-before `height`. Returns zero if no snapshot exists.
func (k Keeper) TotalVotingPowerAt(ctx context.Context, height int64) (math.Int, error) {
	rng := new(collections.Range[int64]).
		EndInclusive(height).
		Descending()

	iter, err := k.TotalPower.Iterate(ctx, rng)
	if err != nil {
		return math.ZeroInt(), err
	}
	defer func() { _ = iter.Close() }()

	if !iter.Valid() {
		return math.ZeroInt(), nil
	}
	return iter.Value()
}

// HeightPower is one (height, power) pair returned by VotingPowerOverRange.
type HeightPower struct {
	Height int64
	Power  math.Int
}

// VotingPowerOverRange returns every recorded snapshot for `del` whose
// height falls in [fromHeight, toHeight] (inclusive). Useful for
// time-decay schemes (conviction voting, plural voting) that want to
// integrate power over a window rather than read at a single height.
//
// Caller-side note: pre-existing at-or-before semantics still apply
// for the boundaries — a delegator who didn't change stake within
// [fromHeight, toHeight] will produce zero rows here, and the caller
// should fall back to VotingPowerAt(fromHeight) to learn the
// constant-over-the-window value.
func (k Keeper) VotingPowerOverRange(ctx context.Context, del sdk.AccAddress, fromHeight, toHeight int64) ([]HeightPower, error) {
	if toHeight < fromHeight {
		return nil, nil
	}
	rng := collections.NewPrefixedPairRange[[]byte, int64](del.Bytes()).
		StartInclusive(fromHeight).
		EndInclusive(toHeight)

	iter, err := k.VotingPower.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer func() { _ = iter.Close() }()

	var out []HeightPower
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		val, err := iter.Value()
		if err != nil {
			return nil, err
		}
		out = append(out, HeightPower{Height: key.K2(), Power: val})
	}
	return out, nil
}
