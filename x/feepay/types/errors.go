package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrContractNotEnoughFunds   = errorsmod.Register(ModuleName, 1, "contract does not have enough funds")
	ErrWalletExceededUsageLimit = errorsmod.Register(ModuleName, 2, "wallet exceeded usage limit")
	ErrInvalidWalletLimit       = errorsmod.Register(ModuleName, 3, "invalid wallet limit; must be between 0 and 1,000,000")
	ErrInvalidJunoFundAmount    = errorsmod.Register(ModuleName, 4, "fee pay contracts only accept juno funds")
	ErrFeePayDisabled           = errorsmod.Register(ModuleName, 5, "the FeePay module is disabled")
	ErrDeductFees               = errorsmod.Register(ModuleName, 6, "error deducting fees")
)
