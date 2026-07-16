package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Query caps for VotingPowerOverRangeCapped. Applied at the wasmbinding
// and gRPC layers so a contract or client cannot induce an unbounded
// store scan.
const (
	// MaxVotingPowerRangeWidth bounds the queryable window to ~1 week of
	// blocks. Wider analytics belong off-chain (indexer), not in
	// consensus-metered queries.
	MaxVotingPowerRangeWidth int64 = 100_800
	// MaxVotingPowerRangeRows bounds the number of returned snapshots.
	// Exceeding it is an error (not a silent truncation) — callers must
	// narrow the range so they never act on partial data.
	MaxVotingPowerRangeRows = 1024
)

// delegatorBondedPower computes the stake counted toward del's voting
// power: tokens delegated to validators currently in Bonded status.
//
// This intentionally differs from staking's GetDelegatorBonded, which
// counts delegations to validators in every bond status. TotalPower is
// derived from TotalBondedTokens (bonded pool only); using the same
// bonded-only basis here keeps Σ VotingPower <= TotalPower and stops
// jailed/unbonded validators' stake from lingering in the numerator.
func (k Keeper) delegatorBondedPower(ctx context.Context, del sdk.AccAddress) (math.Int, error) {
	bonded := math.LegacyZeroDec()
	var innerErr error
	err := k.stakingKeeper.IterateDelegatorDelegations(ctx, del, func(d stakingtypes.Delegation) bool {
		valAddr, err := sdk.ValAddressFromBech32(d.ValidatorAddress)
		if err != nil {
			innerErr = err
			return true
		}
		val, err := k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
				// dangling delegation record; contributes nothing
				return false
			}
			innerErr = err
			return true
		}
		if !val.IsBonded() {
			return false
		}
		bonded = bonded.Add(val.TokensFromSharesTruncated(d.Shares))
		return false
	})
	if err != nil {
		return math.ZeroInt(), err
	}
	if innerErr != nil {
		return math.ZeroInt(), innerErr
	}
	return bonded.RoundInt(), nil
}

// recordDelegatorPower writes a (delegator, height) snapshot for the
// delegator's current bonded-validator stake. Idempotent within a block:
// repeated writes at the same height overwrite. LSTs short-circuit to 0.
func (k Keeper) recordDelegatorPower(ctx context.Context, del sdk.AccAddress) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()

	isLST, err := k.IsLST(ctx, del)
	if err != nil {
		return err
	}
	power := math.ZeroInt()
	if !isLST {
		power, err = k.delegatorBondedPower(ctx, del)
		if err != nil {
			return err
		}
	}
	return k.VotingPower.Set(ctx, collections.Join[[]byte, int64](del.Bytes(), height), power)
}

// computeTotalPower returns the chain-wide voting-power denominator:
// staking's TotalBondedTokens minus the bonded stake held by
// LST-allowlisted addresses.
//
// LST symmetry: allowlisted addresses record zero per-delegator power
// AND their stake is excluded here, so numerator and denominator share
// one basis and Σ VotingPower[d,h] <= TotalPower[h] holds (up to
// shares-rounding dust from TokensFromSharesTruncated). The allowlist
// is governance-managed and expected to stay small, so the per-LST
// delegation walk is O(|allowlist| * delegations-per-LST) and cheap.
func (k Keeper) computeTotalPower(ctx context.Context) (math.Int, error) {
	total, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return total, nil
		}
		return math.ZeroInt(), err
	}
	for _, listed := range params.LstAllowlist {
		addr, err := sdk.AccAddressFromBech32(listed)
		if err != nil {
			// Params.Validate rejects malformed entries at set time;
			// surface rather than skip if one somehow got in.
			return math.ZeroInt(), err
		}
		lstPower, err := k.delegatorBondedPower(ctx, addr)
		if err != nil {
			return math.ZeroInt(), err
		}
		total = total.Sub(lstPower)
	}
	if total.IsNegative() {
		total = math.ZeroInt()
	}
	return total, nil
}

// recordTotal writes a (height) snapshot of the LST-adjusted total
// voting power. See computeTotalPower for the basis.
func (k Keeper) recordTotal(ctx context.Context) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()
	total, err := k.computeTotalPower(ctx)
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

// TotalVotingPowerAt returns the most recent total-power snapshot
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
// Unbounded — internal use only. Externally reachable surfaces
// (wasmbindings, gRPC) must call VotingPowerOverRangeCapped.
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

// VotingPowerOverRangeCapped is the externally reachable variant of
// VotingPowerOverRange. It rejects windows wider than
// MaxVotingPowerRangeWidth and result sets larger than
// MaxVotingPowerRangeRows — errors, never silent truncation, so callers
// can't mistake a partial window for the whole one.
func (k Keeper) VotingPowerOverRangeCapped(ctx context.Context, del sdk.AccAddress, fromHeight, toHeight int64) ([]HeightPower, error) {
	if toHeight < fromHeight {
		return nil, nil
	}
	if fromHeight < 0 {
		fromHeight = 0
	}
	if toHeight-fromHeight > MaxVotingPowerRangeWidth {
		return nil, ErrRangeTooWide
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
		if len(out) >= MaxVotingPowerRangeRows {
			return nil, ErrRangeTooManyRows
		}
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
