package keeper

import (
	"context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Hooks returns a wrapper implementing the staking hook interface.
func (k Keeper) Hooks() Hooks { return Hooks{k: k} }

type Hooks struct {
	k Keeper
}

// staking lifecycle: dirty-marking only
//
// Hooks never compute or write power. Several of them fire while the
// staking store still holds the pre-mutation state:
//   - BeforeDelegationRemoved fires before the delegation record is
//     deleted, so an in-hook GetDelegatorBonded-style recompute returns
//     the un-undelegated amount (permanent phantom voting power);
//   - BeforeValidatorSlashed fires before RemoveValidatorTokens, so an
//     in-hook recompute records pre-slash values.
//
// Instead every hook marks the affected delegator(s) dirty in the
// transient store; the module EndBlocker (ordered after staking's
// EndBlocker in app/modules.go) recomputes each dirty delegator's power
// from settled state and writes the snapshot at the current height.

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	return h.k.markDirty(ctx, delAddr)
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	return h.k.markDirty(ctx, delAddr)
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, _ math.LegacyDec) error {
	// Every delegator under valAddr loses tokens once x/staking applies
	// the slash; mark them all dirty so the EndBlocker re-records their
	// post-slash power.
	return h.k.markValidatorDelegatorsDirty(ctx, valAddr)
}

// Validator bond-status transitions change whether a validator's
// delegations count toward voting power (only Bonded validators count,
// matching the TotalBondedTokens denominator). Mark all of the
// validator's delegators dirty so their numerators are re-recorded on
// the same block the denominator moves.

func (h Hooks) AfterValidatorBonded(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.markValidatorDelegatorsDirty(ctx, valAddr)
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.markValidatorDelegatorsDirty(ctx, valAddr)
}

func (h Hooks) AfterValidatorRemoved(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	// By removal time the validator has no delegations left (staking only
	// removes zero-share validators), so the walk is empty — but the
	// total may still be stale in edge cases; mark it for recompute.
	return h.k.markValidatorDelegatorsDirty(ctx, valAddr)
}

// remaining staking hooks: no-op

func (Hooks) AfterValidatorCreated(_ context.Context, _ sdk.ValAddress) error { return nil }

func (Hooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (Hooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (Hooks) BeforeDelegationSharesModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}
func (Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error { return nil }

// compile-time assertion that Hooks satisfies stakingtypes.StakingHooks.
var _ stakingtypes.StakingHooks = Hooks{}
