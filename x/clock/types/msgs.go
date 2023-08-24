package types

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/drip/types"
)

const (
	// Sudo Message called on the contracts
	EndBlockSudoMessage = `{"end_block":{}}`
)

// == MsgUpdateParams ==
const TypeMsgUpdateParams = "update_clock_params"

var _ sdk.Msg = &MsgUpdateParams{}

// NewMsgUpdateParams creates new instance of MsgUpdateParams
func NewMsgUpdateParams(
	sender sdk.Address,
	contracts []string,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: sender.String(),
		Params:    Params{ContractAddresses: contracts},
	}
}

// Route returns the name of the module
func (msg MsgUpdateParams) Route() string { return types.RouterKey }

// Type returns the the action
func (msg MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	return msg.Params.Validate()
}
