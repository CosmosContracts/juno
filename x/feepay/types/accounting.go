package types

import (
	"math"

	sdkmath "cosmossdk.io/math"
)

// MaxFeePayContractBalance is the largest amount representable by the
// consensus FeePay contract balance field.
const MaxFeePayContractBalance = uint64(math.MaxUint64)

// ContractBalanceAfterAddition converts amount without truncation and returns
// the resulting balance only when it fits the FeePay uint64 ledger.
func ContractBalanceAfterAddition(balance uint64, amount sdkmath.Int) (uint64, error) {
	if !amount.BigInt().IsUint64() {
		return 0, ErrFeePayAmountOutOfRange.Wrapf("amount %s is outside uint64 range", amount)
	}

	addition := amount.Uint64()
	if addition > MaxFeePayContractBalance-balance {
		return 0, ErrFeePayBalanceOverflow.Wrapf("balance %d plus amount %s exceeds %d", balance, amount, MaxFeePayContractBalance)
	}

	return balance + addition, nil
}

// ContractBalanceAfterSubtraction converts amount without truncation and
// returns the resulting balance only when the ledger has sufficient funds.
func ContractBalanceAfterSubtraction(balance uint64, amount sdkmath.Int) (uint64, error) {
	if !amount.BigInt().IsUint64() {
		return 0, ErrFeePayAmountOutOfRange.Wrapf("amount %s is outside uint64 range", amount)
	}

	deduction := amount.Uint64()
	if deduction > balance {
		return 0, ErrContractNotEnoughFunds.Wrapf("expected: %s, got: %d", amount, balance)
	}

	return balance - deduction, nil
}
