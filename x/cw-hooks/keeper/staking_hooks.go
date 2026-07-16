package keeper

import (
	"context"
	"encoding/json"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
)

// skipUntilHeight allows us to skip gentxs.
const skipUntilHeight = 2

type StakingHooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = StakingHooks{}

// StakingHooks creates new hooks for the staking module
func (k Keeper) StakingHooks() StakingHooks {
	return StakingHooks{k: k}
}

// AfterValidatorCreated is a hook that runs after anyone registers as a new validator
func (h StakingHooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorCreated: ", val)
	if err != nil {
		// A best-effort notification hook must never fail a staking state
		// transition — log and continue rather than propagating a halt.
		h.k.Logger(ctx).Error("AfterValidatorCreated: failed to read validator", "error", err)
		return nil
	}
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorCreated{
		AfterValidatorCreated: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "AfterValidatorCreated")
}

// AfterValidatorRemoved is a hook that runs after anyone deletes their validator
func (h StakingHooks) AfterValidatorRemoved(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorRemoved: ", val)
	if err != nil {
		h.k.Logger(ctx).Error("AfterValidatorRemoved: failed to read validator", "error", err)
		return nil
	}
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorRemoved{
		AfterValidatorRemoved: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "AfterValidatorRemoved")
}

// BeforeDelegationCreated is a hook that runs BEFORE any user stakes some tokens
func (h StakingHooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	// The delegation object does not exist yet at the "Before" hook for a first
	// delegation, so the previous Delegation() lookup always returned nil and
	// the create event never fired. Build the payload from the addresses
	// directly (zero shares) so the create notification always dispatches.
	msgBz, err := json.Marshal(SudoMsgBeforeDelegationCreated{
		BeforeDelegationCreated: &Delegation{
			ValidatorAddress: valAddr.String(),
			DelegatorAddress: delAddr.String(),
			Shares:           "0",
		},
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "BeforeDelegationCreated")
}

// BeforeDelegationSharesModified that runs BEFORE we update the staked amount for a user in a validator
func (h StakingHooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	h.k.Logger(ctx).Debug("BeforeDelegationSharesModified: ", del)
	if err != nil {
		// Best-effort notification: never fail a staking state transition.
		h.k.Logger(ctx).Error("staking hook: failed to read delegation", "error", err)
		return nil
	}
	if del == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationSharesModified{
		BeforeDelegationSharesModified: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "BeforeDelegationSharesModified")
}

// AfterDelegationModified is a hook that runs AFTER any user redelegates/unstakes from a validator
func (h StakingHooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	h.k.Logger(ctx).Debug("BeforeDelegationSharesModified: ", del)
	if err != nil {
		// Best-effort notification: never fail a staking state transition.
		h.k.Logger(ctx).Error("staking hook: failed to read delegation", "error", err)
		return nil
	}
	if del == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgAfterDelegationModified{
		AfterDelegationModified: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "AfterDelegationModified")
}

// BeforeValidatorSlashed is a hook that runs right BEFORE a validator is slashed for misbehaviour
func (h StakingHooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("BeforeValidatorSlashed: ", val, fraction)
	if err != nil {
		// Best-effort notification: never fail a staking state transition.
		h.k.Logger(ctx).Error("staking hook: failed to read validator", "error", err)
		return nil
	}
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorSlashed{
		BeforeValidatorSlashed: NewValidatorSlashed(val, fraction),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "BeforeValidatorSlashed")
}

// BeforeValidatorModified is a hook that runs BEFORE a validator updates their validator configuration
func (h StakingHooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("BeforeValidatorModified: ", val)
	if err != nil {
		// Best-effort notification: never fail a staking state transition.
		h.k.Logger(ctx).Error("staking hook: failed to read validator", "error", err)
		return nil
	}
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorModified{
		BeforeValidatorModified: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "BeforeValidatorModified")
}

func (h StakingHooks) AfterValidatorBonded(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorBonded: ", val)
	if err != nil {
		// Best-effort notification: never fail a staking state transition.
		h.k.Logger(ctx).Error("staking hook: failed to read validator", "error", err)
		return nil
	}
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBonded{
		AfterValidatorBonded: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "AfterValidatorBonded")
}

func (h StakingHooks) AfterValidatorBeginUnbonding(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorBeginUnbonding: ", val)
	if err != nil {
		// Best-effort notification: never fail a staking state transition.
		h.k.Logger(ctx).Error("staking hook: failed to read validator", "error", err)
		return nil
	}
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBeginUnbonding{
		AfterValidatorBeginUnbonding: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "AfterValidatorBeginUnbonding")
}

// BeforeDelegationRemoved is a hook that runs BEFORE a user claims their unstaked tokens back
func (h StakingHooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	h.k.Logger(ctx).Debug("BeforeDelegationRemoved: ", del)
	if err != nil {
		// Best-effort notification: never fail a staking state transition.
		h.k.Logger(ctx).Error("staking hook: failed to read delegation", "error", err)
		return nil
	}
	if del == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationRemoved{
		BeforeDelegationRemoved: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.dispatchHookMessage(ctx, types.StakingPrefixKey, msgBz, "BeforeDelegationRemoved")
}

func (StakingHooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}
