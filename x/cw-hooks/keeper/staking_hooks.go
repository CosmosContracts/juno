package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/CosmosContracts/juno/v17/x/cw-hooks/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: return nil on unmarshal or err?

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
func (h StakingHooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	fmt.Println("VALIDATOR_CREATED: ", val)
	if val == nil {
		return nil
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
func (h StakingHooks) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	fmt.Println("AfterValidatorRemoved: ", val)
	if val == nil {
		return nil
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
func (h StakingHooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationCreated: ", del)
	if del == nil {
		return nil
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
func (h StakingHooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationSharesModified: ", del)
	if del == nil {
		return nil
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
func (h StakingHooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// h.k.initializeDelegation(ctx, valAddr, delAddr)
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationSharesModified: ", del)
	if del == nil {
		return nil
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
func (h StakingHooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	fmt.Println("BeforeValidatorSlashed: ", val, fraction)
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorSlashed{
		BeforeValidatorSlashed: NewValidatorSlashed(val, fraction),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	fmt.Println("BeforeValidatorModified: ", val)
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorModified{
		BeforeValidatorModified: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) AfterValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	fmt.Println("AfterValidatorBonded: ", val)
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBonded{
		AfterValidatorBonded: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) AfterValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.GetStakingKeeper().Validator(ctx, valAddr)
	fmt.Println("AfterValidatorBeginUnbonding: ", val)
	if val == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBeginUnbonding{
		AfterValidatorBeginUnbonding: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	del := h.k.GetStakingKeeper().Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationRemoved: ", del)
	if del == nil {
		return nil
	}

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationRemoved{
		BeforeDelegationRemoved: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.k.ExecuteMessageOnContracts(ctx, types.KeyPrefixStaking, msgBz)
}

func (h StakingHooks) AfterUnbondingInitiated(_ sdk.Context, _ uint64) error {
	return nil
}
