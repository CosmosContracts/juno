package util

import sdk "github.com/cosmos/cosmos-sdk/types"

// Signers converts signer bech32 addresses to sdk.AccAddress list. The function
// ignores errors. It is supposed to be used within Msg.GetSigners implementation.
func Signers(signers ...string) []sdk.AccAddress {
	as := make([]sdk.AccAddress, len(signers))
	for i := range signers {
		a, _ := sdk.AccAddressFromBech32(signers[i])
		as[i] = a
	}
	return as
}
