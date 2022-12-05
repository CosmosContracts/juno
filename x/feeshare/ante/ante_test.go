package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	ante "github.com/CosmosContracts/juno/v12/x/feeshare/ante"
)

type AnteTestSuite struct {
	suite.Suite
}

func TestAnteSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (suite *AnteTestSuite) TestFeeLogic() {

	// We expect all to pass
	feeCoins := sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(500)))

	testCases := []struct {
		name               string
		incomingFee        sdk.Coins
		govPercent         sdk.Dec
		numContracts       int
		expectedFeePayment sdk.Coin
	}{
		{
			"100% fee",
			feeCoins,
			sdk.NewDecWithPrec(100, 2),
			1,
			sdk.NewCoin("ujuno", sdk.NewInt(500)),
		},
		{
			"100% fee / 2 contracts @ 500 fee = 250 per",
			feeCoins,
			sdk.NewDecWithPrec(100, 2),
			2,
			sdk.NewCoin("ujuno", sdk.NewInt(250)),
		},
		{
			"100% fee / 10 contracts @ 500 fee = 50 per",
			feeCoins,
			sdk.NewDecWithPrec(100, 2),
			10,
			sdk.NewCoin("ujuno", sdk.NewInt(50)),
		},
		{
			"67% fee / 7 contracts @ 500 fee = 48 per",
			feeCoins,
			sdk.NewDecWithPrec(67, 2),
			7,
			sdk.NewCoin("ujuno", sdk.NewInt(48)),
		},
		{
			"50% fee / 1 contracts @ 500 fee = 250 per",
			feeCoins,
			sdk.NewDecWithPrec(50, 2),
			1,
			sdk.NewCoin("ujuno", sdk.NewInt(250)),
		},
		{
			"50% fee / 2 contracts @ 500 fee = 125 per",
			feeCoins,
			sdk.NewDecWithPrec(50, 2),
			2,
			sdk.NewCoin("ujuno", sdk.NewInt(125)),
		},
		{
			"50% fee / 3 contracts @ 500 fee = 83 per",
			feeCoins,
			sdk.NewDecWithPrec(50, 2),
			3,
			sdk.NewCoin("ujuno", sdk.NewInt(83)),
		},
		{
			"25% fee / 2 contracts @ 500 fee = 62 per",
			feeCoins,
			sdk.NewDecWithPrec(25, 2),
			2,
			sdk.NewCoin("ujuno", sdk.NewInt(62)),
		},
		{
			"15% fee / 3 contracts @ 500 fee = 25 per",
			feeCoins,
			sdk.NewDecWithPrec(15, 2),
			3,
			sdk.NewCoin("ujuno", sdk.NewInt(25)),
		},
		{
			"1% fee / 2 contracts @ 500 fee = 0 per",
			feeCoins,
			sdk.NewDecWithPrec(1, 2),
			2,
			sdk.NewCoin("ujuno", sdk.NewInt(2)),
		},
	}

	for _, tc := range testCases {
		coins := ante.FeePayLogic(tc.incomingFee, tc.govPercent, tc.numContracts)

		suite.Require().Equal(tc.expectedFeePayment.Amount.Int64(), coins[0].Amount.Int64(), tc.name)
	}
}
