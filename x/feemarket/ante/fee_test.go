package ante_test

import (
	"fmt"

	_ "github.com/cosmos/cosmos-sdk/x/auth"

	ibcchanneltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v30/app/ante/decorators"
	"github.com/CosmosContracts/juno/v30/testutil"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

// newBypassMsg returns an IBC relayer message that is in the default
// bypass-min-fee allow-list.
func newBypassMsg(signer sdk.AccAddress) sdk.Msg {
	return &ibcchanneltypes.MsgRecvPacket{Signer: signer.String()}
}

func (s *AnteTestSuite) TestAnteHandle() {
	// Same data for every test case
	gasLimit := NewTestGasLimit()

	// "stake" is the feemarket fee denom in test genesis (types.DefaultParams).
	// Any other denom must be REJECTED: v30 wires the ErrorDenomResolver, so
	// fees are payable only in the fee (bond) denom.
	validFeeAmount := feemarkettypes.DefaultMinBaseGasPrice.MulInt64(int64(gasLimit))
	validFee := sdk.NewCoins(sdk.NewCoin("stake", validFeeAmount.TruncateInt()))
	validFeeDifferentDenom := sdk.NewCoins(sdk.NewCoin("uatom", validFeeAmount.TruncateInt()))
	// permissionless tokenfactory-style denom — the C2 fee-bypass vector
	tokenfactoryFee := sdk.NewCoins(sdk.NewCoin("factory/juno12creator/fakefee", validFeeAmount.TruncateInt()))

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
					Msgs:      []sdk.Msg{NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "0 gas given should fail with non-fee denom",
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
				s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))))

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "fee in non-fee denom is rejected even with funds",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInvalidRequest,
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
				Name:     "fee in permissionless tokenfactory denom is rejected",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInvalidRequest,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], tokenfactoryFee)

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: tokenfactoryFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "zero-fee tx of only bypass msgs passes",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  true,
				ExpErr:   nil,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{newBypassMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: nil,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "zero-fee bypass tx over the gas cap fails",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInvalidGasLimit,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{newBypassMsg(s.TestAccs[0])},
					GasLimit:  decorators.MaxBypassMinFeeMsgGasUsage + 1,
					FeeAmount: nil,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "zero-fee tx mixing bypass and normal msgs fails",
				RunAnte:  true,
				RunPost:  false,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   feemarkettypes.ErrNoFeeCoins,
				Mock:     false,
			},
			Malleate: func(s *AnteTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{newBypassMsg(s.TestAccs[0]), testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: nil,
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
