package testutil

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestAccount represents an address and its private key used in the tests.
type TestAccount struct {
	Account sdk.AccountI
	Priv    cryptotypes.PrivKey
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	Name              string
	Malleate          func(*KeeperTestHelper) TestCaseArgs
	StateUpdate       func(*KeeperTestHelper)
	RunAnte           bool
	RunPost           bool
	Simulate          bool
	ExpPass           bool
	ExpErr            error
	ExpectConsumedGas uint64
	Mock              bool
}

type TestCaseArgs struct {
	ChainID   string
	AccNums   []uint64
	AccSeqs   []uint64
	FeeAmount sdk.Coins
	GasLimit  uint64
	Msgs      []sdk.Msg
	Privs     []cryptotypes.PrivKey
}
