package keeper

import (
	"bytes"
	"context"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxDelegatorsPerSnapshotWalk is the advisory bound on a single
// validator-delegation walk (slash / bond-status change). Juno's largest
// validator carries ~6k delegators, so 10k is headroom. The walk is
// never truncated — dropping a delegator would freeze stale voting
// power in state — but crossing this threshold is logged so operators
// can see the block-time cost coming.
const MaxDelegatorsPerSnapshotWalk = 10_000

// markDirty stages a delegator for end-of-block snapshot recompute and
// flags the chain-wide total as dirty. Idempotent within a block.
func (k Keeper) markDirty(ctx context.Context, del sdk.AccAddress) error {
	if err := k.DirtyDelegators.Set(ctx, del.Bytes()); err != nil {
		return err
	}
	return k.TotalDirty.Set(ctx, true)
}

// markValidatorDelegatorsDirty stages every delegator of valAddr for
// end-of-block recompute. This is the one place a validator-delegation
// walk is acceptable: it only writes cheap transient-store flags, and it
// fires on rare events (slash, validator bond-status change), not on the
// per-tx delegation path.
func (k Keeper) markValidatorDelegatorsDirty(ctx context.Context, valAddr sdk.ValAddress) error {
	dels, err := k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return err
	}
	if len(dels) > MaxDelegatorsPerSnapshotWalk {
		// Never truncate — a skipped delegator would keep phantom power.
		// Log so the anomaly is visible.
		k.Logger(ctx).Error(
			"validator delegation walk exceeds advisory cap; processing all delegators",
			"validator", valAddr.String(),
			"delegations", len(dels),
			"cap", MaxDelegatorsPerSnapshotWalk,
		)
	}
	for _, d := range dels {
		addr, err := sdk.AccAddressFromBech32(d.DelegatorAddress)
		if err != nil {
			return err
		}
		if err := k.DirtyDelegators.Set(ctx, addr.Bytes()); err != nil {
			return err
		}
	}
	return k.TotalDirty.Set(ctx, true)
}

// EndBlocker drains the transient dirty-delegator set: for each dirty
// delegator it recomputes bonded power from (now fully settled) staking
// state and writes the snapshot at the current height; if the total is
// dirty it recomputes and writes TotalPower. Runs after staking's
// EndBlocker (see orderEndBlockers in app/modules.go), so slashes,
// undelegations, and validator bond-status transitions from this block
// are all reflected.
//
// Determinism: dirty keys are collected from an ordered store iterator
// and explicitly sorted before any state write (mirrors backfill.go) —
// no Go-map iteration feeds state.
func (k Keeper) EndBlocker(ctx context.Context) error {
	var dirty [][]byte
	if err := k.DirtyDelegators.Walk(ctx, nil, func(key []byte) (bool, error) {
		dirty = append(dirty, append([]byte(nil), key...))
		return false, nil
	}); err != nil {
		return err
	}

	sort.Slice(dirty, func(i, j int) bool { return bytes.Compare(dirty[i], dirty[j]) < 0 })

	for _, del := range dirty {
		if err := k.recordDelegatorPower(ctx, sdk.AccAddress(del)); err != nil {
			return err
		}
	}
	// Drain explicitly (the transient store also auto-resets on commit,
	// but tests and any same-block re-entry rely on the explicit clear).
	if err := k.DirtyDelegators.Clear(ctx, nil); err != nil {
		return err
	}

	totalDirty, err := k.TotalDirty.Has(ctx)
	if err != nil {
		return err
	}
	if totalDirty {
		if err := k.recordTotal(ctx); err != nil {
			return err
		}
		if err := k.TotalDirty.Remove(ctx); err != nil {
			return err
		}
	}
	return nil
}
