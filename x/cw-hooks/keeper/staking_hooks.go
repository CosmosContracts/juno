package keeper

import (
	"context"
	"encoding/json"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v29/x/cw-hooks/types"
)

// skipUntilHeight allows us to skip gentxs.
const skipUntilHeight = 2

type StakingHooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = StakingHooks{}

// Create new distribution hooks
func (k Keeper) StakingHooks() StakingHooks {
	return StakingHooks{k: k}
}

// initialize validator distribution record
func (h StakingHooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorCreated: ", val)
	if val == nil {
		return err
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorCreated{
		AfterValidatorCreated: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

// AfterValidatorRemoved performs clean up after a validator is removed
func (h StakingHooks) AfterValidatorRemoved(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorRemoved: ", val)
	if val == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorRemoved{
		AfterValidatorRemoved: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

// increment period
func (h StakingHooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	h.k.Logger(ctx).Debug("BeforeDelegationCreated: ", del)
	if del == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationCreated{
		BeforeDelegationCreated: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

// withdraw delegation rewards (which also increments period)
func (h StakingHooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	h.k.Logger(ctx).Debug("BeforeDelegationSharesModified: ", del)
	if del == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationSharesModified{
		BeforeDelegationSharesModified: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

// create new delegation period record
func (h StakingHooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	h.k.Logger(ctx).Debug("BeforeDelegationSharesModified: ", del)
	if del == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgAfterDelegationModified{
		AfterDelegationModified: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

// record the slash event
func (h StakingHooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("BeforeValidatorSlashed: ", val, fraction)
	if val == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorSlashed{
		BeforeValidatorSlashed: NewValidatorSlashed(val, fraction),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("BeforeValidatorModified: ", val)
	if val == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorModified{
		BeforeValidatorModified: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) AfterValidatorBonded(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorBonded: ", val)
	if val == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBonded{
		AfterValidatorBonded: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) AfterValidatorBeginUnbonding(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val, err := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	h.k.Logger(ctx).Debug("AfterValidatorBeginUnbonding: ", val)
	if val == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBeginUnbonding{
		AfterValidatorBeginUnbonding: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del, err := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	h.k.Logger(ctx).Debug("BeforeDelegationRemoved: ", del)
	if del == nil {
		return nil
	}
	if err != nil {
		return err
	}

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationRemoved{
		BeforeDelegationRemoved: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (StakingHooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}
