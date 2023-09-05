package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type AfterValidatorCreated struct {
	// TODO: What else?
	ValidatorAddress string `json:"validator_address"`
	Moniker          string `json:"moniker"`
	GetCommission    string `json:"commission"`
}

type SudoMsgAfterValidatorCreated struct {
	AfterValidatorCreated AfterValidatorCreated `json:"after_validator_created"`
}

// == Wrapper struct ==
type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k: k}
}

// initialize validator distribution record
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	// ignore gentxs
	if ctx.BlockHeight() <= 2 {
		return nil
	}

	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	fmt.Println("VALIDATOR_CREATED: ", val)

	msgBz, err := json.Marshal(SudoMsgAfterValidatorCreated{
		AfterValidatorCreated: AfterValidatorCreated{
			ValidatorAddress: val.GetOperator().String(),
			Moniker:          val.GetMoniker(),
			GetCommission:    val.GetCommission().String(),
		},
	})
	if err != nil {
		return nil
	}
	fmt.Println("MSG_BYTES: " + string(msgBz))

	// TODO: add this in the keeper, anyone can register it.
	contract, err := sdk.AccAddressFromBech32("juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8")
	if err != nil {
		return nil
	}

	// 100k gas limit
	gasLimitCtx := ctx.WithGasMeter(sdk.NewGasMeter(100_000))
	if _, err = h.k.contractKeeper.Sudo(gasLimitCtx, contract, msgBz); err != nil {
		return nil
	}

	// ctx.GasMeter().ConsumeGas(100_000, "juno-staking-hooks: AfterValidatorCreated")

	return nil
}

// AfterValidatorRemoved performs clean up after a validator is removed
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// increment period
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// val := h.k.stakingKeeper.Validator(ctx, valAddr)
	return nil
}

// withdraw delegation rewards (which also increments period)
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// val := h.k.stakingKeeper.Validator(ctx, valAddr)
	// del := h.k.stakingKeeper.Delegation(ctx, delAddr, valAddr)

	return nil
}

// create new delegation period record
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// h.k.initializeDelegation(ctx, valAddr, delAddr)
	return nil
}

// record the slash event
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	// h.k.updateValidatorSlashFraction(ctx, valAddr, fraction)
	return nil
}

func (h Hooks) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ sdk.Context, _ uint64) error {
	return nil
}
