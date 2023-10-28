package types

import (
	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidAddress            = sdkerrors.ErrInvalidAddress
	ErrContractNotRegistered     = errorsmod.Register(ModuleName, 1, "contract not registered")
	ErrContractAlreadyRegistered = errorsmod.Register(ModuleName, 2, "contract already registered")
	ErrContractRegisterNotAdmin  = errorsmod.Register(ModuleName, 3, "this address is not the contract admin, cannot register")
	ErrContractNotEnoughFunds    = errorsmod.Register(ModuleName, 4, "contract does not have enough funds")
	ErrWalletExceededUsageLimit  = errorsmod.Register(ModuleName, 5, "wallet exceeded usage limit")
	ErrContractNotAdmin          = errorsmod.Register(ModuleName, 6, "sender is not the contract admin")
	ErrContractNotCreator        = errorsmod.Register(ModuleName, 7, "sender is not the contract creator")
	ErrInvalidWalletLimit        = errorsmod.Register(ModuleName, 8, "invalid wallet limit; must be between 0 and 1,000,000")
	ErrInvalidJunoFundAmount     = errorsmod.Register(ModuleName, 9, "fee pay contracts only accept juno funds")
	ErrInvalidCWContract         = errorsmod.Register(ModuleName, 10, "invalid CosmWasm contract")
	ErrFeePayDisabled            = errorsmod.Register(ModuleName, 11, "the FeePay module is disabled")
)
