package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// skipUntilHeight allows us to skip gentxs.
const skipUntilHeight = 2

type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k: k}
}

func (h Hooks) sendMsgToAll(ctx sdk.Context, msgBz []byte) error {
	// on errors return nil, if in a loop continue.

	// TODO: add this in the keeper, anyone can register it.
	// iter all contracts here
	contract, err := sdk.AccAddressFromBech32("juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8")
	if err != nil {
		return nil
	}

	// 100k/250k gas limit?
	gasLimitCtx := ctx.WithGasMeter(sdk.NewGasMeter(100_000))
	if _, err = h.k.contractKeeper.Sudo(gasLimitCtx, contract, msgBz); err != nil {
		return nil
	}

	// ctx.GasMeter().ConsumeGas(100_000, "cw-staking-hooks: AfterValidatorCreated")
	return nil
}

// initialize validator distribution record
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	fmt.Println("VALIDATOR_CREATED: ", val)

	msgBz, err := json.Marshal(SudoMsgAfterValidatorCreated{
		AfterValidatorCreated: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

// AfterValidatorRemoved performs clean up after a validator is removed
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	fmt.Println("AfterValidatorRemoved: ", val)

	msgBz, err := json.Marshal(SudoMsgAfterValidatorRemoved{
		AfterValidatorRemoved: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

// increment period
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del := h.k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationCreated: ", del)

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationCreated{
		BeforeDelegationCreated: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

// withdraw delegation rewards (which also increments period)
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del := h.k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationSharesModified: ", del)

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationSharesModified{
		BeforeDelegationSharesModified: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

// create new delegation period record
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// h.k.initializeDelegation(ctx, valAddr, delAddr)
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}

	del := h.k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationSharesModified: ", del)

	msgBz, err := json.Marshal(SudoMsgAfterDelegationModified{
		AfterDelegationModified: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

// record the slash event
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	fmt.Println("BeforeValidatorSlashed: ", val, fraction)

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorSlashed{
		BeforeValidatorSlashed: NewValidatorSlashed(val, fraction),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	fmt.Println("BeforeValidatorModified: ", val)

	msgBz, err := json.Marshal(SudoMsgBeforeValidatorModified{
		BeforeValidatorModified: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

func (h Hooks) AfterValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	fmt.Println("AfterValidatorBonded: ", val)

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBonded{
		AfterValidatorBonded: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	fmt.Println("AfterValidatorBeginUnbonding: ", val)

	msgBz, err := json.Marshal(SudoMsgAfterValidatorBeginUnbonding{
		AfterValidatorBeginUnbonding: NewValidator(val),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if ctx.BlockHeight() <= skipUntilHeight {
		return nil
	}
	del := h.k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	fmt.Println("BeforeDelegationRemoved: ", del)

	msgBz, err := json.Marshal(SudoMsgBeforeDelegationRemoved{
		BeforeDelegationRemoved: NewDelegation(del),
	})
	if err != nil {
		return nil
	}

	return h.sendMsgToAll(ctx, msgBz)
}

func (h Hooks) AfterUnbondingInitiated(_ sdk.Context, _ uint64) error {
	// idk what this is / does. Need to look into
	return nil
}
