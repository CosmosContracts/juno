package ante_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v30/app/decorators"
	"github.com/CosmosContracts/juno/v30/testutil"
	keeper "github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	feemarketpost "github.com/CosmosContracts/juno/v30/x/feemarket/post"
	"github.com/CosmosContracts/juno/v30/x/feemarket/types"
	testdata "github.com/cosmos/cosmos-sdk/testutil/testdata"
)

type AnteTestSuite struct {
	testutil.KeeperTestHelper

	fullAccs []testutil.TestAccount

	AnteHandler sdk.AnteHandler
	PostHandler sdk.PostHandler

	TxBuilder client.TxBuilder

	msgServer   types.MsgServer
	queryServer types.QueryServer
}

type AnteTestCase struct {
	testutil.TestCase
	Malleate    func(*AnteTestSuite) testutil.TestCaseArgs
	StateUpdate func(*AnteTestSuite)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (s *AnteTestSuite) SetupTest() {
	s.Setup()
	s.TxBuilder = s.App.TxConfig().NewTxBuilder()
	s.queryServer = keeper.NewQueryServer(*s.App.AppKeepers.FeeMarketKeeper)
	s.msgServer = keeper.NewMsgServer(s.App.AppKeepers.FeeMarketKeeper)

	testAccs := s.CreateTestAccounts(4)
	s.TestAccs = make([]sdk.AccAddress, len(testAccs))
	for i, acc := range testAccs {
		s.TestAccs[i] = acc.Account.GetAddress()
	}

	s.fullAccs = testAccs

	s.App.AppKeepers.FeeMarketKeeper.SetEnabledHeight(s.Ctx, -1)

	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(),
		decorators.NewDeductFeeDecorator(
			s.App.AppKeepers.FeePayKeeper,
			*s.App.AppKeepers.FeeMarketKeeper,
			s.App.AppKeepers.AccountKeeper,
			s.App.AppKeepers.BankKeeper,
			s.App.AppKeepers.FeeGrantKeeper,
			"ujuno",
			ante.NewDeductFeeDecorator(
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
		feemarketpost.NewFeeMarketDeductDecorator(
			s.App.AppKeepers.AccountKeeper,
			s.App.AppKeepers.BankKeeper,
			*s.App.AppKeepers.FeeMarketKeeper,
		),
	)
}

func (s *AnteTestSuite) RunTestCase(t *testing.T, tc AnteTestCase, args testutil.TestCaseArgs) {
	require.NoError(t, s.TxBuilder.SetMsgs(args.Msgs...))
	s.TxBuilder.SetFeeAmount(args.FeeAmount)
	s.TxBuilder.SetGasLimit(args.GasLimit)

	// Theoretically speaking, ante handler unit tests should only test
	// ante handlers, but here we also test the tx creation process.
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
func (s *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (authsigning.Tx, error) {
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

// NewTestFeeAmount is a test fee amount.
func NewTestFeeAmount() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("stake", 150))
}

// NewTestGasLimit is a test fee gas limit.
func NewTestGasLimit() uint64 {
	return 200000
}

// NewTestMsg creates a message for testing with the given signers.
func NewTestMsg(t *testing.T, addrs ...sdk.AccAddress) *testdata.TestMsg {
	var accAddresses []string

	for _, addr := range addrs {
		accAddresses = append(accAddresses, addr.String())
	}

	return &testdata.TestMsg{
		Signers: accAddresses,
	}
}

func (s *AnteTestSuite) CreateTestAccounts(numAccs int) []testutil.TestAccount {
	s.T().Helper()

	var accounts []testutil.TestAccount

	for i := range numAccs {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := s.App.AppKeepers.AccountKeeper.NewAccountWithAddress(s.Ctx, addr)
		err := acc.SetAccountNumber(uint64(i + 1000))
		if err != nil {
			panic(err)
		}
		s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, acc)
		accounts = append(accounts, testutil.TestAccount{Account: acc, Priv: priv})
	}

	return accounts
}
