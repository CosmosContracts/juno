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
	for i := range numAccts {
		pk := secp256k1.GenPrivKey()
		testAddrs[i] = sdk.AccAddress(pk.PubKey().Address())
		testPrivKeys[i] = pk
	}

	return testAddrs, testPrivKeys
}
