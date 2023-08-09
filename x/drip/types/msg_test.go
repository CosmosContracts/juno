package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgsTestSuite struct {
	suite.Suite
	amount sdk.Coins
	sender sdk.AccAddress
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) SetupTest() {
	sender := "cosmos1"
	suite.sender = sdk.AccAddress([]byte(sender))
	suite.amount = sdk.NewCoins(sdk.NewCoin("ujuice", sdk.NewInt(1000000)))
}

func (suite *MsgsTestSuite) TestMsgDistributeTokensGetters() {
	msgInvalid := MsgDistributeTokens{}
	msg := NewMsgDistributeTokens(
		suite.amount,
		suite.sender,
	)
	suite.Require().Equal(RouterKey, msg.Route())
	suite.Require().Equal(TypeMsgDistributeTokens, msg.Type())
	suite.Require().NotNil(msgInvalid.GetSignBytes())
	suite.Require().NotNil(msg.GetSigners())
}

func (suite *MsgsTestSuite) TestMsgDistributeTokensNew() {
	testCases := []struct {
		msg        string
		amount     sdk.Coins
		sender     string
		expectPass bool
	}{
		{
			"pass",
			suite.amount,
			suite.sender.String(),
			true,
		},
		{
			"sender address cannot be empty",
			suite.amount,
			"",
			false,
		},
		{
			"invalid coins",
			nil,
			suite.sender.String(),
			false,
		},
	}

	for i, tc := range testCases {
		tx := MsgDistributeTokens{
			Amount:        tc.amount,
			SenderAddress: tc.sender,
		}

		err := tx.ValidateBasic()

		if tc.expectPass {
			suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
			suite.Require().Contains(err.Error(), tc.msg)
		}
	}
}
