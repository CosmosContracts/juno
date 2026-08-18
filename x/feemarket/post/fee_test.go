package post_test

import (
	"fmt"
	stdmath "math"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	ibcchanneltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	junoapp "github.com/CosmosContracts/juno/v31/app"
	"github.com/CosmosContracts/juno/v31/app/ante/decorators"
	"github.com/CosmosContracts/juno/v31/testutil"
	keeper "github.com/CosmosContracts/juno/v31/x/feemarket/keeper"
	"github.com/CosmosContracts/juno/v31/x/feemarket/post"
	"github.com/CosmosContracts/juno/v31/x/feemarket/types"
	feepaytypes "github.com/CosmosContracts/juno/v31/x/feepay/types"
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

type feePayTestTx struct {
	msgs []sdk.Msg
}

func (tx feePayTestTx) GetMsgs() []sdk.Msg                 { return tx.msgs }
func (feePayTestTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (feePayTestTx) GetGas() uint64                        { return 1 }
func (feePayTestTx) GetFee() sdk.Coins                     { return nil }
func (feePayTestTx) FeePayer() []byte                      { return nil }
func (feePayTestTx) FeeGranter() []byte                    { return nil }

func TestPostTestSuite(t *testing.T) {
	suite.Run(t, new(PostTestSuite))
}

func (s *PostTestSuite) SetupTest() {
	s.Setup()
	s.TxBuilder = s.App.TxConfig().NewTxBuilder()
	s.queryServer = keeper.NewQueryServer(*s.App.AppKeepers.FeeMarketKeeper)
	s.msgServer = keeper.NewMsgServer(s.App.AppKeepers.FeeMarketKeeper)

	s.App.AppKeepers.FeeMarketKeeper.SetEnabledHeight(s.Ctx, -1)

	// register the shared test accounts so fee deduction can resolve them
	for i, addr := range s.TestAccs {
		acc := s.App.AppKeepers.AccountKeeper.NewAccountWithAddress(s.Ctx, addr)
		s.Require().NoError(acc.SetAccountNumber(uint64(i + 1000)))
		s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, acc)
	}

	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(),
		decorators.NewDeductFeeDecorator(
			s.App.AppKeepers.FeePayKeeper,
			*s.App.AppKeepers.FeeMarketKeeper,
			s.App.AppKeepers.AccountKeeper,
			s.App.AppKeepers.BankKeeper,
			s.App.AppKeepers.FeeGrantKeeper,
			// bondDenom — matches the feemarket fee denom in test genesis
			"stake",
			junoapp.GetDefaultBypassFeeMessages(),
			authante.NewDeductFeeDecorator(
				s.App.AppKeepers.AccountKeeper,
				s.App.AppKeepers.BankKeeper,
				s.App.AppKeepers.FeeGrantKeeper,
				nil,
			),
		),
		authante.NewSigGasConsumeDecorator(s.App.AppKeepers.AccountKeeper, authante.DefaultSigVerificationGasConsumer),
	}
	s.AnteHandler = sdk.ChainAnteDecorators(anteDecorators...)

	s.PostHandler = sdk.ChainPostDecorators(
		post.NewFeeMarketDeductDecorator(
			s.App.AppKeepers.AccountKeeper,
			s.App.AppKeepers.BankKeeper,
			*s.App.AppKeepers.FeeMarketKeeper,
			s.App.AppKeepers.FeePayKeeper,
			s.App.AppKeepers.StakingKeeper,
		),
	)
}

func (s *PostTestSuite) RunTestCase(t *testing.T, tc PostTestCase, args testutil.TestCaseArgs) {
	require.NoError(t, s.TxBuilder.SetMsgs(args.Msgs...))
	s.TxBuilder.SetFeeAmount(args.FeeAmount)
	s.TxBuilder.SetGasLimit(args.GasLimit)

	// Theoretically speaking, ante handler unit tests should only test
	// ante handlers, but here we sometimes also test the tx creation
	// process.
	testTx, txErr := s.CreateTestTx(args.Privs, args.AccNums, args.AccSeqs, args.ChainID)

	var (
		newCtx  sdk.Context
		anteErr error
		postErr error
	)

	// reset gas meter
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit()))

	if tc.RunAnte {
		newCtx, anteErr = s.AnteHandler(s.Ctx, testTx, tc.Simulate)
	}

	// perform mid-tx state update if configured
	if tc.StateUpdate != nil {
		tc.StateUpdate(s)
	}

	if tc.RunPost && anteErr == nil {
		newCtx, postErr = s.PostHandler(s.Ctx, testTx, tc.Simulate, true)
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

func (s *PostTestSuite) TestPostHandle() {
	// Same data for every test case
	const (
		baseDenom           = "stake"
		resolvableDenom     = "atom"
		expectedConsumedGas = 36650

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
				ExpectConsumedGas: 24202,
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
				ExpectConsumedGas: 24208,
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
				ExpectConsumedGas: 11736,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
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
				Name:     "signer has does not have enough funds for fee and tip - fail",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInsufficientFunds,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], validFee)

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
				ExpectConsumedGas: 11736,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], validFeeWithTip)

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
				ExpectConsumedGas: 24256,
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
				s.FundAcc(s.TestAccs[0], validResolvableFee)

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
				Name:     "fee in non-fee denom rejected in ante (v30 ErrorDenomResolver)",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInvalidRequest,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], validResolvableFee)

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
				ExpectConsumedGas: 2333,
				Mock:              false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], validResolvableFee)

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
				ExpectConsumedGas: 2333,
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
				Name:     "fee with tip in non-fee denom rejected in ante (v30 ErrorDenomResolver)",
				RunAnte:  true,
				RunPost:  true,
				Simulate: false,
				ExpPass:  false,
				ExpErr:   sdkerrors.ErrInvalidRequest,
				Mock:     false,
			},
			Malleate: func(s *PostTestSuite) testutil.TestCaseArgs {
				s.FundAcc(s.TestAccs[0], validResolvableFeeWithTip)

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
				ExpectConsumedGas: 2333,
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

// TestFeePayNoProposerTipAndRefund covers the feepay path end to end:
//   - the ante escrows price × gasLimit out of the contract's feepay balance;
//   - the post handler deducts only the CONSUMED fee;
//   - the unused remainder is refunded to the x/feepay module account and
//     re-credited to the contract's feepay balance;
//   - the proposer receives NO tip;
//   - the payout is bounded by this tx's escrow, not the whole collector
//     balance (soft-burn accumulation with DistributeFees=false stays put).
func (s *PostTestSuite) TestFeePayNoProposerTipAndRefund() {
	s.SetupTest()

	const (
		gasLimit        = uint64(300_000) // headroom: this meter also pays for the post handler's writes
		contractBalance = uint64(10_000_000)
		preExisting     = int64(500_000) // simulates DistributeFees=false accumulation
	)

	// register a feepay contract directly in the keeper store
	contractAccAddr := sdk.AccAddress([]byte("feepay_contract_addr_x")).String()
	fpc := feepaytypes.FeePayContract{
		ContractAddress: contractAccAddr,
		Balance:         contractBalance,
		WalletLimit:     100,
	}
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, fpc)

	// fund the feepay module so it can escrow, and pre-fund the collector to
	// simulate accumulated (non-distributed) fees a feepay tx must NOT touch
	s.FundModuleAcc(feepaytypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("stake", int64(contractBalance))))
	s.FundModuleAcc(types.FeeCollectorName, sdk.NewCoins(sdk.NewInt64Coin("stake", preExisting)))

	// the tx signer needs an account (but NO funds — feepay pays)
	signerPriv, _, signerAddr := testdata.KeyTestPubAddr()
	acc := s.App.AppKeepers.AccountKeeper.NewAccountWithAddress(s.Ctx, signerAddr)
	s.Require().NoError(acc.SetPubKey(signerPriv.PubKey()))
	s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, acc)

	execMsg := &wasmtypes.MsgExecuteContract{
		Sender:   signerAddr.String(),
		Contract: contractAccAddr,
		Msg:      []byte("{}"),
	}

	s.Require().NoError(s.TxBuilder.SetMsgs(execMsg))
	s.TxBuilder.SetFeeAmount(nil) // zero fee: feepay covers it
	s.TxBuilder.SetGasLimit(gasLimit)
	testTx, err := s.CreateTestTx(
		[]cryptotypes.PrivKey{signerPriv},
		[]uint64{acc.GetAccountNumber()},
		[]uint64{0},
		s.Ctx.ChainID(),
	)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit()))

	// gas price is DefaultMinBaseGasPrice = 1stake/gas in test genesis
	newCtx, err := s.AnteHandler(s.Ctx, testTx, false)
	s.Require().NoError(err)
	s.Ctx = newCtx

	escrow := int64(gasLimit) // 1stake/gas × gasLimit

	// escrow moved feepay module -> collector; contract balance decremented
	feepayAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(feepaytypes.ModuleName)
	collectorAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	s.Require().Equal(int64(contractBalance)-escrow, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, "stake").Amount.Int64())
	s.Require().Equal(preExisting+escrow, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, "stake").Amount.Int64())

	contract, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAccAddr)
	s.Require().NoError(err)
	s.Require().Equal(contractBalance-uint64(escrow), contract.Balance)

	// make the block proposer resolvable so any (erroneous) tip payout would
	// be observable on the operator account
	vals, err := s.App.AppKeepers.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(vals)
	consAddr, err := vals[0].GetConsAddr()
	s.Require().NoError(err)
	header := s.Ctx.BlockHeader()
	header.ProposerAddress = consAddr
	s.Ctx = s.Ctx.WithBlockHeader(header)

	valAddr, err := sdk.ValAddressFromBech32(vals[0].GetOperator())
	s.Require().NoError(err)
	operatorAcc := sdk.AccAddress(valAddr)
	operatorBalBefore := s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, operatorAcc, "stake").Amount

	gasConsumedBeforePost := int64(s.Ctx.GasMeter().GasConsumed())
	s.Require().Positive(gasConsumedBeforePost)
	s.Require().Less(gasConsumedBeforePost, escrow)

	_, err = s.PostHandler(s.Ctx, testTx, false, true)
	s.Require().NoError(err)

	// proposer/operator got NO tip
	s.Require().Equal(operatorBalBefore, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, operatorAcc, "stake").Amount)

	// derive the consumed fee from the contract balance (the post handler
	// itself consumes gas between our capture above and its CheckTxFee call,
	// so the exact number is not predictable from here). At 1stake/gas the
	// consumed fee equals the gas consumed at CheckTxFee time.
	contract, err = s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAccAddr)
	s.Require().NoError(err)
	consumed := int64(contractBalance) - int64(contract.Balance)
	s.Require().GreaterOrEqual(consumed, gasConsumedBeforePost)
	s.Require().Less(consumed, escrow)
	refund := escrow - consumed

	// refund flowed back to the feepay module account and the contract balance:
	// net contract charge is exactly the consumed fee, not the full gas limit
	s.Require().Equal(int64(contractBalance)-escrow+refund, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, "stake").Amount.Int64())

	// the pre-existing collector balance was NOT siphoned: only the consumed
	// fee remains on top of it (DistributeFees=false keeps it in the account)
	s.Require().Equal(preExisting+consumed, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, "stake").Amount.Int64())
}

// TestFeePayUsesConfiguredFeeDenom covers the consensus-sensitive case where
// feemarket's configured fee denom differs from the staking bond denom.
func (s *PostTestSuite) TestFeePayUsesConfiguredFeeDenom() {
	s.SetupTest()

	const (
		feeDenom       = "ufee"
		gasLimit       = uint64(300_000)
		initialBalance = uint64(10_000_000)
	)

	params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	params.FeeDenom = feeDenom
	s.Require().NoError(s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, params))

	contractAddr := sdk.AccAddress([]byte("fee_denom_contract_x")).String()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
		ContractAddress: contractAddr,
		Balance:         initialBalance,
		WalletLimit:     100,
	})
	s.FundModuleAcc(feepaytypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(feeDenom, int64(initialBalance))))

	signerPriv, _, signerAddr := testdata.KeyTestPubAddr()
	acc := s.App.AppKeepers.AccountKeeper.NewAccountWithAddress(s.Ctx, signerAddr)
	s.Require().NoError(acc.SetPubKey(signerPriv.PubKey()))
	s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, acc)

	execMsg := &wasmtypes.MsgExecuteContract{Sender: signerAddr.String(), Contract: contractAddr, Msg: []byte("{}")}
	s.Require().NoError(s.TxBuilder.SetMsgs(execMsg))
	s.TxBuilder.SetFeeAmount(nil)
	s.TxBuilder.SetGasLimit(gasLimit)
	testTx, err := s.CreateTestTx(
		[]cryptotypes.PrivKey{signerPriv},
		[]uint64{acc.GetAccountNumber()},
		[]uint64{0},
		s.Ctx.ChainID(),
	)
	s.Require().NoError(err)
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit()))

	newCtx, err := s.AnteHandler(s.Ctx, testTx, false)
	s.Require().NoError(err)
	s.Ctx = newCtx

	escrow := int64(gasLimit)
	feepayAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(feepaytypes.ModuleName)
	collectorAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	s.Require().Equal(int64(initialBalance)-escrow, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, feeDenom).Amount.Int64())
	s.Require().Equal(escrow, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, feeDenom).Amount.Int64())

	_, err = s.PostHandler(s.Ctx, testTx, false, true)
	s.Require().NoError(err)

	contract, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(err)
	consumed := int64(initialBalance) - int64(contract.Balance)
	s.Require().Positive(consumed)
	s.Require().Less(consumed, escrow)
	s.Require().Equal(int64(initialBalance)-consumed, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, feeDenom).Amount.Int64())
	s.Require().Equal(consumed, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, feeDenom).Amount.Int64())
	s.Require().True(s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, "stake").IsZero())
}

func (s *PostTestSuite) TestFeePayMissingConfiguredFeeDenomFundsHasNoMutation() {
	s.SetupTest()

	const (
		feeDenom       = "ufee"
		gasLimit       = uint64(300_000)
		initialBalance = uint64(10_000_000)
	)

	params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	params.FeeDenom = feeDenom
	s.Require().NoError(s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, params))

	contractAddr := sdk.AccAddress([]byte("missing_fee_funds_xx")).String()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
		ContractAddress: contractAddr,
		Balance:         initialBalance,
		WalletLimit:     100,
	})
	// Deliberately fund only the bond denom. Accounting claims sufficient
	// funds, but the configured fee-denom escrow is absent.
	s.FundModuleAcc(feepaytypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("stake", int64(initialBalance))))

	signerPriv, _, signerAddr := testdata.KeyTestPubAddr()
	acc := s.App.AppKeepers.AccountKeeper.NewAccountWithAddress(s.Ctx, signerAddr)
	s.Require().NoError(acc.SetPubKey(signerPriv.PubKey()))
	s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, acc)

	execMsg := &wasmtypes.MsgExecuteContract{Sender: signerAddr.String(), Contract: contractAddr, Msg: []byte("{}")}
	s.Require().NoError(s.TxBuilder.SetMsgs(execMsg))
	s.TxBuilder.SetFeeAmount(nil)
	s.TxBuilder.SetGasLimit(gasLimit)
	testTx, err := s.CreateTestTx(
		[]cryptotypes.PrivKey{signerPriv},
		[]uint64{acc.GetAccountNumber()},
		[]uint64{0},
		s.Ctx.ChainID(),
	)
	s.Require().NoError(err)
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit()))

	_, err = s.AnteHandler(s.Ctx, testTx, false)
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
	s.Require().ErrorContains(err, "error transferring funds from FeePay to FeeCollector")

	contract, getErr := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(getErr)
	s.Require().Equal(initialBalance, contract.Balance)
	uses, getErr := s.App.AppKeepers.FeePayKeeper.GetContractUses(s.Ctx, contract, signerAddr.String())
	s.Require().NoError(getErr)
	s.Require().Zero(uses)
	feepayAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(feepaytypes.ModuleName)
	collectorAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	s.Require().Equal(int64(initialBalance), s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, "stake").Amount.Int64())
	s.Require().True(s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, feeDenom).IsZero())
	s.Require().True(s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, feeDenom).IsZero())
}

func (s *PostTestSuite) TestFeePayAnteRejectsRequiredFeeAboveUint64Atomically() {
	s.SetupTest()
	state, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
	s.Require().NoError(err)
	state.BaseGasPrice = math.LegacyNewDec(3)
	s.Require().NoError(s.App.AppKeepers.FeeMarketKeeper.SetState(s.Ctx, state))

	contractAddr := sdk.AccAddress([]byte("12345678901234567890")).String()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
		ContractAddress: contractAddr, Balance: stdmath.MaxUint64, WalletLimit: 10,
	})
	signerPriv, _, signerAddr := testdata.KeyTestPubAddr()
	acc := s.App.AppKeepers.AccountKeeper.NewAccountWithAddress(s.Ctx, signerAddr)
	s.Require().NoError(acc.SetPubKey(signerPriv.PubKey()))
	s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, acc)
	s.Require().NoError(s.TxBuilder.SetMsgs(&wasmtypes.MsgExecuteContract{
		Sender: signerAddr.String(), Contract: contractAddr, Msg: []byte("{}"),
	}))
	s.TxBuilder.SetFeeAmount(nil)
	s.TxBuilder.SetGasLimit(uint64(stdmath.MaxInt64))
	testTx, err := s.CreateTestTx([]cryptotypes.PrivKey{signerPriv},
		[]uint64{acc.GetAccountNumber()}, []uint64{0}, s.Ctx.ChainID())
	s.Require().NoError(err)

	_, err = s.AnteHandler(s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit())), testTx, false)
	s.Require().ErrorIs(err, feepaytypes.ErrFeePayAmountOutOfRange)
	contract, getErr := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(getErr)
	s.Require().Equal(uint64(stdmath.MaxUint64), contract.Balance)
	uses, getErr := s.App.AppKeepers.FeePayKeeper.GetContractUses(s.Ctx, contract, signerAddr.String())
	s.Require().NoError(getErr)
	s.Require().Zero(uses)
	collectorAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	s.Require().True(s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, "stake").IsZero())
}

func (s *PostTestSuite) TestFeePayRefundOverflowIsAtomic() {
	s.SetupTest()
	contractAddr := sdk.AccAddress([]byte("12345678901234567890")).String()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
		ContractAddress: contractAddr,
		Balance:         stdmath.MaxUint64,
	})
	s.FundModuleAcc(types.FeeCollectorName, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))

	dfd := post.NewFeeMarketDeductDecorator(
		s.App.AppKeepers.AccountKeeper,
		s.App.AppKeepers.BankKeeper,
		*s.App.AppKeepers.FeeMarketKeeper,
		s.App.AppKeepers.FeePayKeeper,
		s.App.AppKeepers.StakingKeeper,
	)
	tx := feePayTestTx{msgs: []sdk.Msg{&wasmtypes.MsgExecuteContract{Contract: contractAddr}}}
	collectorAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	feepayAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(feepaytypes.ModuleName)
	err := dfd.PayOutFeeAndRefundFeePay(s.Ctx, tx,
		sdk.NewCoin("stake", math.ZeroInt()), sdk.NewInt64Coin("stake", 1))
	s.Require().ErrorIs(err, feepaytypes.ErrFeePayBalanceOverflow)

	contract, getErr := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(getErr)
	s.Require().Equal(uint64(stdmath.MaxUint64), contract.Balance)
	s.Require().Equal(math.NewInt(1), s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, "stake").Amount)
	s.Require().True(s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, feepayAddr, "stake").IsZero())
}

// TestTipPaidToProposerOperatorAccount asserts the proposer tip goes to the
// validator OPERATOR account resolved via the consensus address — not to the
// raw consensus address cast to an AccAddress (an unspendable account).
func (s *PostTestSuite) TestTipPaidToProposerOperatorAccount() {
	s.SetupTest()

	fee := sdk.NewInt64Coin("stake", 1000)
	tip := sdk.NewInt64Coin("stake", 250)
	s.FundModuleAcc(types.FeeCollectorName, sdk.NewCoins(fee.Add(tip)))

	// Create a validator whose consensus key differs from its operator key so
	// the raw cons-address cast and the operator account are distinct — the
	// exact confusion the fix addresses. (The testutil genesis validator uses
	// the same key for both, which would mask the bug.)
	consPriv := ed25519.GenPrivKey()
	// exactly 20 bytes — the app's address verifier enforces 20/32-byte addresses
	operatorAddr := sdk.ValAddress([]byte("distinct_operator_20"))
	validator, err := stakingtypes.NewValidator(operatorAddr.String(), consPriv.PubKey(), stakingtypes.Description{Moniker: "tip-test"})
	s.Require().NoError(err)
	s.Require().NoError(s.App.AppKeepers.StakingKeeper.SetValidator(s.Ctx, validator))
	s.Require().NoError(s.App.AppKeepers.StakingKeeper.SetValidatorByConsAddr(s.Ctx, validator))

	consAddr := sdk.ConsAddress(consPriv.PubKey().Address())

	header := s.Ctx.BlockHeader()
	header.ProposerAddress = consAddr
	s.Ctx = s.Ctx.WithBlockHeader(header)

	operatorAcc := sdk.AccAddress(operatorAddr)
	rawConsAcc := sdk.AccAddress(consAddr)
	s.Require().False(operatorAcc.Equals(rawConsAcc))

	dfd := post.NewFeeMarketDeductDecorator(
		s.App.AppKeepers.AccountKeeper,
		s.App.AppKeepers.BankKeeper,
		*s.App.AppKeepers.FeeMarketKeeper,
		s.App.AppKeepers.FeePayKeeper,
		s.App.AppKeepers.StakingKeeper,
	)

	s.Require().NoError(dfd.PayOutFeeAndTip(s.Ctx, fee, tip))

	// operator account received the tip; the raw-cast consensus address got nothing
	s.Require().Equal(tip.Amount, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, operatorAcc, "stake").Amount)
	s.Require().True(s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, rawConsAcc, "stake").IsZero())
}

// TestZeroFeeBypassTxSkipsDeduction asserts the post handler deducts nothing
// for a zero-fee non-feepay tx (a bypass-min-fee relayer tx) instead of
// erroring or reading the collector balance.
func (s *PostTestSuite) TestZeroFeeBypassTxSkipsDeduction() {
	s.SetupTest()

	// pre-fund the collector: a bypass tx must not move any of it
	preExisting := sdk.NewInt64Coin("stake", 123_456)
	s.FundModuleAcc(types.FeeCollectorName, sdk.NewCoins(preExisting))

	recvMsg := &ibcchanneltypes.MsgRecvPacket{Signer: s.TestAccs[0].String()}
	s.Require().NoError(s.TxBuilder.SetMsgs(recvMsg))
	s.TxBuilder.SetFeeAmount(nil)
	s.TxBuilder.SetGasLimit(150_000)
	testTx := s.TxBuilder.GetTx()

	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit()))
	s.Ctx.GasMeter().ConsumeGas(50_000, "simulated execution")

	stateBefore, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
	s.Require().NoError(err)

	_, err = s.PostHandler(s.Ctx, testTx, false, true)
	s.Require().NoError(err)

	// collector untouched
	collectorAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	s.Require().Equal(preExisting.Amount, s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, collectorAddr, "stake").Amount)

	// but the gas was still recorded in the fee market window (the recorded
	// value includes the post handler's own reads on top of the 50k)
	stateAfter, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(stateAfter.Window[stateAfter.Index], stateBefore.Window[stateBefore.Index]+50_000)
}
