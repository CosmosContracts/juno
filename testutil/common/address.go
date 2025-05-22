package common

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateRandomAccountsAndPrivKeys is a function return a list of randomly generated AccAddresses and their according private keys.
func CreateRandomAccountsAndPrivKeys(numAccts int) ([]sdk.AccAddress, []cryptotypes.PrivKey) {
	testAddrs := make([]sdk.AccAddress, numAccts)
	testPrivKeys := make([]cryptotypes.PrivKey, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := secp256k1.GenPrivKey()
		testAddrs[i] = sdk.AccAddress(pk.PubKey().Address())
		testPrivKeys[i] = pk
	}

	return testAddrs, testPrivKeys
}

// These are for testing msg.ValidateBasic() functions
// which need to validate for valid/invalid addresses.
// Should not be used for anything else because these addresses
// are totally uninterpretable (100% random).
func GenerateTestAddrs() (string, string) {
	pk1 := secp256k1.GenPrivKey().PubKey()
	validAddr := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("").String()
	return validAddr, invalidAddr
}
