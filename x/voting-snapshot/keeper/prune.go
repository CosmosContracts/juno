package keeper

import (
	"bytes"
	"context"
	"fmt"
	stdmath "math"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/voting-snapshot/types"
)

// MaxPruneDeletionsPerRun bounds how many snapshot keys a single prune
// sweep may delete (per collection). A run that hits the cap logs and
// leaves the remainder for the next interval — never a silent drop of
// data a read still needs (the keep-most-recent-below-cutoff guard is
// applied before the cap, so capping only defers deletion of already
// stale keys).
const MaxPruneDeletionsPerRun = 10_000

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
// same treatment for symmetry — the total only writes on staking
// events, not every block, so a quiet period can produce the same
// sparse pattern.
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
	// Params.Validate bounds both fields to int64 range at set time;
	// re-check here so a legacy/corrupt value can't overflow the casts.
	if params.RetentionWindowHeights > uint64(stdmath.MaxInt64) {
		return fmt.Errorf("%w: retention_window_heights %d", types.ErrParamOutOfRange, params.RetentionWindowHeights)
	}
	if params.PruneInterval > uint64(stdmath.MaxInt64) {
		return fmt.Errorf("%w: prune_interval %d", types.ErrParamOutOfRange, params.PruneInterval)
	}
	interval := int64(params.PruneInterval)
	if interval <= 0 {
		interval = 1
	}
	if height%interval != 0 {
		return nil
	}

	cutoff := height - int64(params.RetentionWindowHeights)
	if cutoff <= 0 {
		return nil
	}

	vpCapped, err := k.pruneVotingPower(ctx, cutoff, MaxPruneDeletionsPerRun)
	if err != nil {
		return err
	}
	tpCapped, err := k.pruneTotalPower(ctx, cutoff, MaxPruneDeletionsPerRun)
	if err != nil {
		return err
	}
	if vpCapped || tpCapped {
		k.Logger(ctx).Info(
			"prune sweep hit per-run deletion cap; remaining stale snapshots defer to the next interval",
			"cap", MaxPruneDeletionsPerRun,
			"voting_power_capped", vpCapped,
			"total_power_capped", tpCapped,
			"cutoff", cutoff,
		)
	}
	return nil
}

// pruneVotingPower deletes snapshots with height < cutoff except for the
// single most-recent below-cutoff snapshot per delegator, which is the
// at-or-before value any read at cutoff would resolve to. The collection
// is sorted by (delegator, height) ascending, so within each delegator
// group the below-cutoff snapshots are a contiguous prefix; we keep the
// last one we see in that prefix and stage the prior ones for deletion.
//
// At most maxDeletions keys are staged per run (capped=true when the
// budget is exhausted). All deletes happen after the iterator is closed
// — the KVStore contract forbids mutating under an open iterator.
func (k Keeper) pruneVotingPower(ctx context.Context, cutoff int64, maxDeletions int) (capped bool, err error) {
	var stale []collections.Pair[[]byte, int64]

	collect := func() error {
		iter, err := k.VotingPower.Iterate(ctx, new(collections.Range[collections.Pair[[]byte, int64]]))
		if err != nil {
			return err
		}
		defer func() { _ = iter.Close() }()

		var (
			currentDel []byte
			pendingKey collections.Pair[[]byte, int64]
			hasPending bool
		)
		for ; iter.Valid(); iter.Next() {
			if len(stale) >= maxDeletions {
				capped = true
				return nil
			}
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
		return nil
	}
	if err := collect(); err != nil {
		return false, err
	}

	// iterator is closed; safe to delete
	for _, key := range stale {
		if err := k.VotingPower.Remove(ctx, key); err != nil {
			return capped, err
		}
	}
	return capped, nil
}

// pruneTotalPower deletes total-power snapshots with height < cutoff
// except for the single most-recent below-cutoff snapshot, which is the
// at-or-before value any read at cutoff would resolve to. Same shape as
// pruneVotingPower without the per-delegator grouping.
func (k Keeper) pruneTotalPower(ctx context.Context, cutoff int64, maxDeletions int) (capped bool, err error) {
	var stale []int64

	collect := func() error {
		iter, err := k.TotalPower.Iterate(ctx, new(collections.Range[int64]).EndExclusive(cutoff))
		if err != nil {
			return err
		}
		defer func() { _ = iter.Close() }()

		var (
			pendingKey int64
			hasPending bool
		)
		for ; iter.Valid(); iter.Next() {
			if len(stale) >= maxDeletions {
				capped = true
				return nil
			}
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
		return nil
	}
	if err := collect(); err != nil {
		return false, err
	}

	// iterator is closed; safe to delete
	for _, key := range stale {
		if err := k.TotalPower.Remove(ctx, key); err != nil {
			return capped, err
		}
	}
	return capped, nil
}
