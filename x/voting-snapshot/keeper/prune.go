package keeper

import (
	"bytes"
	"context"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Prune drops snapshots with height < (current_height -
// RetentionWindowHeights). Called from the module's EndBlocker.
// No-op if the retention window is 0 (disabled) or if the current
// block isn't on a prune-interval boundary.
//
// Per-delegator the pruner preserves the most recent snapshot that
// is still below the cutoff (h_max_below). Without that guard a
// set-and-forget delegator whose last staking event predates the
// retention window would lose every snapshot and read zero voting
// power even though their stake is unchanged. TotalPower gets the
// same treatment for symmetry — TotalBondedTokens only writes on
// staking events, not every block, so a quiet period can produce
// the same sparse pattern.
func (k Keeper) Prune(ctx context.Context) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()

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
	interval := int64(params.PruneInterval) //nolint:gosec // governance-bounded
	if interval <= 0 {
		interval = 1
	}
	if height%interval != 0 {
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

// pruneVotingPower deletes snapshots with height < cutoff except for the
// single most-recent below-cutoff snapshot per delegator, which is the
// at-or-before value any read at cutoff would resolve to. The collection
// is sorted by (delegator, height) ascending, so within each delegator
// group the below-cutoff snapshots are a contiguous prefix; we keep the
// last one we see in that prefix and stage the prior ones for deletion.
func (k Keeper) pruneVotingPower(ctx context.Context, cutoff int64) error {
	rng := new(collections.Range[collections.Pair[[]byte, int64]])
	iter, err := k.VotingPower.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()

	var (
		stale      []collections.Pair[[]byte, int64]
		currentDel []byte
		pendingKey collections.Pair[[]byte, int64]
		hasPending bool
	)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		if !bytes.Equal(key.K1(), currentDel) {
			currentDel = append(currentDel[:0], key.K1()...)
			hasPending = false
		}
		if key.K2() >= cutoff {
			continue
		}
		if hasPending {
			stale = append(stale, pendingKey)
		}
		pendingKey = key
		hasPending = true
	}
	for _, key := range stale {
		if err := k.VotingPower.Remove(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// pruneTotalPower deletes total-supply snapshots with height < cutoff
// except for the single most-recent below-cutoff snapshot, which is the
// at-or-before value any read at cutoff would resolve to. Same shape as
// pruneVotingPower without the per-delegator grouping.
func (k Keeper) pruneTotalPower(ctx context.Context, cutoff int64) error {
	rng := new(collections.Range[int64]).EndExclusive(cutoff)
	iter, err := k.TotalPower.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer func() { _ = iter.Close() }()

	var (
		stale      []int64
		pendingKey int64
		hasPending bool
	)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		if hasPending {
			stale = append(stale, pendingKey)
		}
		pendingKey = key
		hasPending = true
	}
	for _, key := range stale {
		if err := k.TotalPower.Remove(ctx, key); err != nil {
			return err
		}
	}
	return nil
}
