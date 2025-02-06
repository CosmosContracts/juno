package keeper

import (
	sdkmath "cosmossdk.io/math"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Validator struct {
	Moniker          string `json:"moniker"`
	ValidatorAddress string `json:"validator_address"`
	Commission       string `json:"commission"`
	ValidatorTokens  string `json:"validator_tokens"`
	BondedTokens     string `json:"bonded_tokens"`
	BondStatus       string `json:"bond_status"`
}

func NewValidator(val stakingtypes.ValidatorI) *Validator {
	return &Validator{
		Moniker:          val.GetMoniker(),
		ValidatorAddress: val.GetOperator(),
		Commission:       val.GetCommission().String(),
		ValidatorTokens:  val.GetTokens().String(),
		BondedTokens:     val.GetBondedTokens().String(),
		BondStatus:       val.GetStatus().String(),
	}
}

type ValidatorSlashed struct {
	Moniker          string `json:"moniker"`
	ValidatorAddress string `json:"validator_address"`
	SlashedAmount    string `json:"slashed_amount"`
}

func NewValidatorSlashed(val stakingtypes.ValidatorI, fraction sdkmath.LegacyDec) *ValidatorSlashed {
	return &ValidatorSlashed{
		Moniker:          val.GetMoniker(),
		ValidatorAddress: val.GetOperator(),
		SlashedAmount:    fraction.String(),
	}
}

type Delegation struct {
	ValidatorAddress string `json:"validator_address"`
	DelegatorAddress string `json:"delegator_address"`
	Shares           string `json:"shares"`
}

func NewDelegation(del stakingtypes.DelegationI) *Delegation {
	return &Delegation{
		ValidatorAddress: del.GetValidatorAddr(),
		DelegatorAddress: del.GetDelegatorAddr(),
		Shares:           del.GetShares().String(),
	}
}

// Validators
type SudoMsgAfterValidatorCreated struct {
	AfterValidatorCreated *Validator `json:"after_validator_created"`
}
type SudoMsgAfterValidatorRemoved struct {
	AfterValidatorRemoved *Validator `json:"after_validator_removed"`
}
type SudoMsgBeforeValidatorModified struct {
	BeforeValidatorModified *Validator `json:"before_validator_modified"`
}
type SudoMsgAfterValidatorModified struct {
	AfterValidatorModified *Validator `json:"after_validator_modified"`
}
type SudoMsgAfterValidatorBonded struct {
	AfterValidatorBonded *Validator `json:"after_validator_bonded"`
}
type SudoMsgAfterValidatorBeginUnbonding struct {
	AfterValidatorBeginUnbonding *Validator `json:"after_validator_begin_unbonding"`
}
type SudoMsgBeforeValidatorSlashed struct {
	BeforeValidatorSlashed *ValidatorSlashed `json:"before_validator_slashed"`
}

// Delegations
type SudoMsgBeforeDelegationCreated struct {
	BeforeDelegationCreated *Delegation `json:"before_delegation_created"`
}
type SudoMsgBeforeDelegationSharesModified struct {
	BeforeDelegationSharesModified *Delegation `json:"before_delegation_shares_modified"`
}
type SudoMsgAfterDelegationModified struct {
	AfterDelegationModified *Delegation `json:"after_delegation_modified"`
}

type SudoMsgBeforeDelegationRemoved struct {
	BeforeDelegationRemoved *Delegation `json:"before_delegation_removed"`
}
