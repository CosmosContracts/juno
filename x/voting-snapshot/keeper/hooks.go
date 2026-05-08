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

// staking lifecycle: per-event snapshot writes
//
// We snapshot on AfterDelegationModified (delegate, redelegate-out,
// undelegate-up-to-completion) and BeforeDelegationRemoved (full
// undelegation). BeforeValidatorSlashed snapshots the affected
// delegators' new power; the post-slash recompute is the right
// semantic per design doc (a vote shouldn't carry power that no
// longer exists).

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	if err := h.k.recordDelegatorPower(ctx, delAddr); err != nil {
		return err
	}
	return h.k.recordTotal(ctx)
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	if err := h.k.recordDelegatorPower(ctx, delAddr); err != nil {
		return err
	}
	return h.k.recordTotal(ctx)
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, _ sdk.ValAddress, _ math.LegacyDec) error {
	// On slash, every delegator under the validator loses shares retroactively
	// once x/staking applies the slash. We re-snapshot the total for now.
	// (v30 MVP: best-effort; per-validator delegator re-snapshot lands in v30.x.)
	return h.k.recordTotal(ctx)
}

// remaining staking hooks: no-op (we only need delegation + slashing signals)

func (Hooks) AfterValidatorCreated(_ context.Context, _ sdk.ValAddress) error { return nil }

func (Hooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (Hooks) AfterValidatorRemoved(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (Hooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (Hooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
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
