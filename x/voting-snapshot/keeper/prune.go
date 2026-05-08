package keeper

import (
	"context"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// pruneInterval bounds how often the EndBlocker actually walks the
// retention boundary. Pruning every block would burn cycles for no
// gain since old snapshots only become eligible to drop one-per-block.
// Once-per-block pruning is also acceptable; this constant lets us
// later batch prunes if the iteration cost becomes meaningful.
const pruneInterval int64 = 1

// Prune drops snapshots with height < (current_height -
// RetentionWindowHeights). Called from the module's EndBlocker.
// No-op if the retention window is 0 (disabled).
func (k Keeper) Prune(ctx context.Context) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()
	if pruneInterval > 0 && height%pruneInterval != 0 {
		return nil
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return err
	}
	if params.RetentionWindowHeights == 0 {
		return nil
	}

	cutoff := height - int64(params.RetentionWindowHeights) //nolint:gosec // RetentionWindowHeights is governance-bounded; overflow not realistic
	if cutoff <= 0 {
		return nil
	}

	if err := k.pruneVotingPower(ctx, cutoff); err != nil {
		return err
	}
	return k.pruneTotalPower(ctx, cutoff)
}

func (k Keeper) pruneVotingPower(ctx context.Context, cutoff int64) error {
	rng := new(collections.Range[collections.Pair[[]byte, int64]])
	// Iterate everything; filter by height in the loop. Range filtering
	// over the second key of a Pair isn't directly expressible —
	// per-delegator iteration with EndExclusive on height would require
	// knowing the delegator set up front, which we don't. The total
	// row count is bounded by `delegators × snapshots-per-delegator`,
	// so this scan is cheap in practice.
	iter, err := k.VotingPower.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()

	var stale []collections.Pair[[]byte, int64]
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		if key.K2() < cutoff {
			stale = append(stale, key)
		}
	}
	for _, key := range stale {
		if err := k.VotingPower.Remove(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) pruneTotalPower(ctx context.Context, cutoff int64) error {
	rng := new(collections.Range[int64]).EndExclusive(cutoff)
	iter, err := k.TotalPower.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()

	var stale []int64
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		stale = append(stale, key)
	}
	for _, key := range stale {
		if err := k.TotalPower.Remove(ctx, key); err != nil {
			return err
		}
	}
	return nil
}
