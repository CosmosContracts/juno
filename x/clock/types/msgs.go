package types

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	globalerrors "github.com/CosmosContracts/juno/v30/app/helpers"
)

const (
	// Sudo Message called on the contracts
	EndBlockSudoMessage = `{"clock_end_block":{}}`
)

// == MsgUpdateParams ==
const (
	TypeMsgRegisterFeePayContract   = "register_clock_contract"
	TypeMsgUnregisterFeePayContract = "unregister_clock_contract"
	TypeMsgUnjailFeePayContract     = "unjail_clock_contract"
	TypeMsgUpdateParams             = "update_clock_params"
)

var (
	_ sdk.Msg = &MsgRegisterClockContract{}
	_ sdk.Msg = &MsgUnregisterClockContract{}
	_ sdk.Msg = &MsgUnjailClockContract{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// ValidateBasic runs stateless checks on the message
func (msg MsgRegisterClockContract) ValidateBasic() error {
	return validateAddresses(msg.SenderAddress, msg.ContractAddress)
}

// ValidateBasic runs stateless checks on the message
func (msg MsgUnregisterClockContract) ValidateBasic() error {
	return validateAddresses(msg.SenderAddress, msg.ContractAddress)
}

// ValidateBasic runs stateless checks on the message
func (msg MsgUnjailClockContract) ValidateBasic() error {
	return validateAddresses(msg.SenderAddress, msg.ContractAddress)
}

// ValidateBasic does a sanity check on the provided data.
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	return msg.Params.Validate()
}

// ValidateAddresses validates the provided addresses
func validateAddresses(addresses ...string) error {
	for _, address := range addresses {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return errors.Wrapf(globalerrors.ErrInvalidAddress, "invalid address: %s", address)
		}
	}

	return nil
}
