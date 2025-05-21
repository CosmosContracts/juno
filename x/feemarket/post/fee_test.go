package post_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v30/testutil"
	keeper "github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	"github.com/CosmosContracts/juno/v30/x/feemarket/post"
	"github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

type PostTestSuite struct {
	testutil.KeeperTestHelper

	AnteHandler sdk.AnteHandler
	PostHandler sdk.PostHandler

	TxBuilder client.TxBuilder

	msgServer   types.MsgServer
	queryServer types.QueryServer
}

type PostTestCase struct {
	testutil.TestCase
	Malleate    func(*PostTestSuite) testutil.TestCaseArgs
	StateUpdate func(*PostTestSuite)
}

func (s *PostTestSuite) SetupTest() {
	s.Setup()
	s.TxBuilder = s.App.TxConfig().NewTxBuilder()
	s.queryServer = keeper.NewQueryServer(*s.App.AppKeepers.FeeMarketKeeper)
	s.msgServer = keeper.NewMsgServer(s.App.AppKeepers.FeeMarketKeeper)
}

func (s *PostTestSuite) SetAccountBalances(accounts []testutil.TestAccountBalance) {
	s.T().Helper()

	oldState := s.App.AppKeepers.BankKeeper.ExportGenesis(s.Ctx)

	balances := make([]banktypes.Balance, len(accounts))
	for i, acc := range accounts {
		balances[i] = banktypes.Balance{
			Address: acc.Account.String(),
			Coins:   acc.Coins,
		}
	}

	oldState.Balances = balances
	s.App.AppKeepers.BankKeeper.InitGenesis(s.Ctx, oldState)
}

func (s *PostTestSuite) RunTestCase(t *testing.T, tc PostTestCase, args testutil.TestCaseArgs) {
	require.NoError(t, s.TxBuilder.SetMsgs(args.Msgs...))
	s.TxBuilder.SetFeeAmount(args.FeeAmount)
	s.TxBuilder.SetGasLimit(args.GasLimit)

	// Theoretically speaking, ante handler unit tests should only test
	// ante handlers, but here we sometimes also test the tx creation
	// process.
	tx, txErr := s.CreateTestTx(args.Privs, args.AccNums, args.AccSeqs, args.ChainID)

	var (
		newCtx  sdk.Context
		anteErr error
		postErr error
	)

	// reset gas meter
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit()))

	if tc.RunAnte {
		newCtx, anteErr = s.AnteHandler(s.Ctx, tx, tc.Simulate)
	}

	// perform mid-tx state update if configured
	if tc.StateUpdate != nil {
		tc.StateUpdate(s)
	}

	if tc.RunPost && anteErr == nil {
		newCtx, postErr = s.PostHandler(s.Ctx, tx, tc.Simulate, true)
	}

	if tc.ExpPass {
		require.NoError(t, txErr)
		require.NoError(t, anteErr)
		require.NoError(t, postErr)
		require.NotNil(t, newCtx)

		s.Ctx = newCtx
		if tc.RunPost {
			consumedGas := newCtx.GasMeter().GasConsumed()
			require.Equal(t, tc.ExpectConsumedGas, consumedGas)
		}

	} else {
		switch {
		case txErr != nil:
			require.Error(t, txErr)
			require.ErrorIs(t, txErr, tc.ExpErr)

		case anteErr != nil:
			require.Error(t, anteErr)
			require.NoError(t, postErr)
			require.ErrorIs(t, anteErr, tc.ExpErr)

		case postErr != nil:
			require.NoError(t, anteErr)
			require.Error(t, postErr)
			require.ErrorIs(t, postErr, tc.ExpErr)

		default:
			t.Fatal("expected one of txErr, handleErr to be an error")
		}
	}
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *PostTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (authsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signing.SignMode(s.App.TxConfig().SignModeHandler().DefaultMode()),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := s.TxBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			s.Ctx,
			signing.SignMode(s.App.TxConfig().SignModeHandler().DefaultMode()), signerData,
			s.TxBuilder, priv, s.App.TxConfig(), accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.TxBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.TxBuilder.GetTx(), nil
}

// NewTestGasLimit is a test fee gas limit.
func NewTestGasLimit() uint64 {
	return 200000
}

func (s *PostTestSuite) TestDeductCoins() {
	tests := []struct {
		name           string
		coins          sdk.Coins
		distributeFees bool
		wantErr        bool
	}{
		{
			name:           "valid",
			coins:          sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10))),
			distributeFees: false,
			wantErr:        false,
		},
		{
			name:           "valid no coins",
			coins:          sdk.NewCoins(),
			distributeFees: false,
			wantErr:        false,
		},
		{
			name:           "valid zero coin",
			coins:          sdk.NewCoins(sdk.NewCoin("test", math.ZeroInt())),
			distributeFees: false,
			wantErr:        false,
		},
		{
			name:           "valid - distribute",
			coins:          sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10))),
			distributeFees: true,
			wantErr:        false,
		},
		{
			name:           "valid no coins - distribute",
			coins:          sdk.NewCoins(),
			distributeFees: true,
			wantErr:        false,
		},
		{
			name:           "valid zero coin - distribute",
			coins:          sdk.NewCoins(sdk.NewCoin("test", math.ZeroInt())),
			distributeFees: true,
			wantErr:        false,
		},
	}
	for _, tc := range tests {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			if err := post.DeductCoins(s.App.AppKeepers.BankKeeper, s.Ctx, tc.coins, tc.distributeFees); (err != nil) != tc.wantErr {
				s.Errorf(err, "DeductCoins() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func (s *PostTestSuite) TestDeductCoinsAndDistribute() {
	tests := []struct {
		name    string
		coins   sdk.Coins
		wantErr bool
	}{
		{
			name:    "valid",
			coins:   sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10))),
			wantErr: false,
		},
		{
			name:    "valid no coins",
			coins:   sdk.NewCoins(),
			wantErr: false,
		},
		{
			name:    "valid zero coin",
			coins:   sdk.NewCoins(sdk.NewCoin("test", math.ZeroInt())),
			wantErr: false,
		},
	}
	for _, tc := range tests {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			if err := post.DeductCoins(s.App.AppKeepers.BankKeeper, s.Ctx, tc.coins, true); (err != nil) != tc.wantErr {
				s.Errorf(err, "DeductCoins() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func (s *PostTestSuite) TestSendTip() {
	tests := []struct {
		name    string
		coins   sdk.Coins
		wantErr bool
	}{
		{
			name:    "valid",
			coins:   sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10))),
			wantErr: false,
		},
		{
			name:    "valid no coins",
			coins:   sdk.NewCoins(),
			wantErr: false,
		},
		{
			name:    "valid zero coin",
			coins:   sdk.NewCoins(sdk.NewCoin("test", math.ZeroInt())),
			wantErr: false,
		},
	}
	for _, tc := range tests {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			if err := post.SendTip(s.App.AppKeepers.BankKeeper, s.Ctx, s.TestAccs[1], tc.coins); (err != nil) != tc.wantErr {
				s.Errorf(err, "SendTip() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func (s *PostTestSuite) TestPostHandleMock() {
	// Same data for every test case
	const (
		baseDenom              = "stake"
		resolvableDenom        = "atom"
		expectedConsumedGas    = 10631
		expectedConsumedSimGas = expectedConsumedGas + post.BankSendGasConsumption
		gasLimit               = expectedConsumedSimGas
	)

	validFeeAmount := types.DefaultMinBaseGasPrice.MulInt64(int64(gasLimit))
	validFeeAmountWithTip := validFeeAmount.Add(math.LegacyNewDec(100))
	validFee := sdk.NewCoins(sdk.NewCoin(baseDenom, validFeeAmount.TruncateInt()))
	validFeeWithTip := sdk.NewCoins(sdk.NewCoin(baseDenom, validFeeAmountWithTip.TruncateInt()))
	validResolvableFee := sdk.NewCoins(sdk.NewCoin(resolvableDenom, validFeeAmount.TruncateInt()))
	validResolvableFeeWithTip := sdk.NewCoins(sdk.NewCoin(resolvableDenom, validFeeAmountWithTip.TruncateInt()))

	testCases := []PostTestCase{
		{
			TestCase: testutil.TestCase{
				Name:     "signer has no funds",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInsufficientFunds,
				Mock:     true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "signer has no funds - simulate",
				RunAnte:  true,
				RunPost:  true,
				Simulate: true,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInsufficientFunds,
				Mock:     true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "0 gas given should fail",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrOutOfGas,
				Mock:     true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "0 gas given should pass - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedSimGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass, no tip",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedSimGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "fee market is enabled during the transaction - should pass and skip deduction until next block",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: 15340, // extra gas consumed because msg server is run, but deduction is skipped
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				// disable fee market before tx
				s.Ctx = s.Ctx.WithBlockHeight(10)
				disabledParams := types.DefaultParams()
				disabledParams.Enabled = false
				err := s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, disabledParams)
				s.Require().NoError(err)

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFee,
				}
			},
			StateUpdate: func(s *PostTestSuite) {
				// enable the fee market
				enabledParams := types.DefaultParams()
				req := &types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    enabledParams,
				}

				_, err := s.msgServer.UpdateParams(s.Ctx, req)
				s.Require().NoError(err)

				height, err := s.App.AppKeepers.FeeMarketKeeper.GetEnabledHeight(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(int64(10), height)
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass, no tip - resolvable denom",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass, no tip - resolvable denom - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedSimGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip - resolvable denom",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip - resolvable denom - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedSimGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "0 gas given should pass in simulate - no fee",
				RunAnte:           true,
				RunPost:           false,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedSimGas,
				Mock:              true,
			},

			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: nil,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "0 gas given should pass in simulate - fee",
				RunAnte:           true,
				RunPost:           false,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedSimGas,
				Mock:              true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
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
				ExpErr:   types.ErrNoFeeCoins,
				Mock:     true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
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
				Mock:     true,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
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

func (s *PostTestSuite) TestPostHandle() {
	// Same data for every test case
	const (
		baseDenom           = "stake"
		resolvableDenom     = "atom"
		expectedConsumedGas = 36650

		expectedConsumedGasResolve = 36524 // slight difference due to denom resolver

		gasLimit = 100000
	)

	validFeeAmount := types.DefaultMinBaseGasPrice.MulInt64(int64(gasLimit))
	validFeeAmountWithTip := validFeeAmount.Add(math.LegacyNewDec(100))
	validFee := sdk.NewCoins(sdk.NewCoin(baseDenom, validFeeAmount.TruncateInt()))
	validFeeWithTip := sdk.NewCoins(sdk.NewCoin(baseDenom, validFeeAmountWithTip.TruncateInt()))
	validResolvableFee := sdk.NewCoins(sdk.NewCoin(resolvableDenom, validFeeAmount.TruncateInt()))
	validResolvableFeeWithTip := sdk.NewCoins(sdk.NewCoin(resolvableDenom, validFeeAmountWithTip.TruncateInt()))

	testCases := []PostTestCase{
		{
			TestCase: testutil.TestCase{
				Name:     "signer has no funds",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInsufficientFunds,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has no funds - simulate - pass",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				Mock:              false,
				ExpectConsumedGas: expectedConsumedGas,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "0 gas given should fail",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrOutOfGas,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "0 gas given should pass - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass, no tip",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: 36650,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validFee,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "signer has does not have enough funds for fee and tip - fail",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInsufficientFunds,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validFee,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: 36650,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validFeeWithTip,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "fee market is enabled during the transaction - should pass and skip deduction until next block",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: 15340, // extra gas consumed because msg server is run, but bank keepers are skipped
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validResolvableFee,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				// disable fee market before tx
				s.Ctx = s.Ctx.WithBlockHeight(10)
				disabledParams := types.DefaultParams()
				disabledParams.Enabled = false
				err := s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, disabledParams)
				s.Require().NoError(err)

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFee,
				}
			},
			StateUpdate: func(s *PostTestSuite) {
				// enable the fee market
				enabledParams := types.DefaultParams()
				req := &types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    enabledParams,
				}

				_, err := s.msgServer.UpdateParams(s.Ctx, req)
				s.Require().NoError(err)

				height, err := s.App.AppKeepers.FeeMarketKeeper.GetEnabledHeight(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(int64(10), height)
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass, no tip - resolvable denom",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGasResolve,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validResolvableFee,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass, no tip - resolvable denom - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validResolvableFee,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has no balance, should pass, no tip - resolvable denom - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFee,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:     "signer does not have enough funds, fail - resolvable denom",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInsufficientFunds,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validResolvableFee,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip - resolvable denom",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          false,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGasResolve,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				testAcc := testutil.TestAccount{
					Account: s.TestAccs[0],
					Priv:    s.TestPrivKeys[0],
				}
				balance := testutil.TestAccountBalance{
					TestAccount: testAcc,
					Coins:       validResolvableFeeWithTip,
				}
				s.SetAccountBalances([]testutil.TestAccountBalance{balance})

				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "signer has enough funds, should pass with tip - resolvable denom - simulate",
				RunAnte:           true,
				RunPost:           true,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  gasLimit,
					FeeAmount: validResolvableFeeWithTip,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "0 gas given should pass in simulate - no fee",
				RunAnte:           true,
				RunPost:           false,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: nil,
				}
			},
		},
		{
			TestCase: testutil.TestCase{
				Name:              "0 gas given should pass in simulate - fee",
				RunAnte:           true,
				RunPost:           false,
				Simulate:          true,
				ExpPass:           true,
				ExpErr:            nil,
				ExpectConsumedGas: expectedConsumedGas,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				return testutil.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(s.TestAccs[0])},
					GasLimit:  0,
					FeeAmount: validFee,
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
				ExpErr:   types.ErrNoFeeCoins,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
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
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
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
