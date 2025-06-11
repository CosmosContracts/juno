package ante_test

import (
	"fmt"

	"github.com/CosmosContracts/juno/v30/testutil"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
)

func (s *AnteTestSuite) TestAnteHandle() {
	// Same data for every test case
	gasLimit := NewTestGasLimit()

	validFeeAmount := feemarkettypes.DefaultMinBaseGasPrice.MulInt64(int64(gasLimit))
	validFee := sdk.NewCoins(sdk.NewCoin("ujuno", validFeeAmount.TruncateInt()))
	validFeeDifferentDenom := sdk.NewCoins(sdk.NewCoin("uatom", math.Int(validFeeAmount)))

	testCases := []AnteTestCase{
		{
			TestCase: testutil.TestCase{
				Name:              "0 gas given should fail",
				RunAnte:           true,
				RunPost:           false,
				Simulate:          false,
				ExpPass:           false,
				ExpErr:            sdkerrors.ErrOutOfGas,
				ExpectConsumedGas: 0,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		// test --gas=auto flag settings
		// when --gas=auto is set, cosmos-sdk sets gas=0 and simulate=true
		{
			TestCase: testutil.TestCase{
				Name:     "--gas=auto behaviour test - no balance",
				RunAnte:  true,
				RunPost:  false,
				Simulate: true,
				ExpPass:  true,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{NewTestMsg(s.T(), s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "0 gas given should fail with resolvable denom",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrOutOfGas,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFeeDifferentDenom,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "0 gas given should pass in simulate - no fee",
				RunAnte:  true,
				RunPost:  false,
				Simulate: true,
				ExpPass:  true,
				ExpErr:   nil,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: nil,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "0 gas given should pass in simulate - fee",
				RunAnte:  true,
				RunPost:  false,
				Simulate: true,
				ExpPass:  true,
				ExpErr:   nil,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "signer has enough funds, should pass",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  true,
				ExpErr:   nil,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {

				s.FundAcc(s.TestAccs[0], validFee)

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "signer has insufficient funds, should fail",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInsufficientFunds,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(100))))

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "signer has enough funds in resolvable denom, should pass",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  true,
				ExpErr:   nil,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], validFeeDifferentDenom)

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFeeDifferentDenom,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "no fee - fail",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   feemarkettypes.ErrNoFeeCoins,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  1000000000,
					FeeAmount: nil,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "no gas limit - fail",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrOutOfGas,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: nil,
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.Name), func() {
			args := tc.Malleate(s)

			s.RunTestCase(s.T(), tc, args)
		})
	}
}
