package ante_test

import (
	"context"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	junoante "github.com/CosmosContracts/juno/v30/app/ante"
	"github.com/CosmosContracts/juno/v30/app/ante/decorators"
	"github.com/CosmosContracts/juno/v30/testutil"
)

func (s *AnteTestSuite) TestNewAnteHandlerUsesEmbeddedFeegrantKeeper() {
	s.SetupTest()

	grantee := s.fullAccs[0]
	granter := s.fullAccs[1]
	fee := sdk.NewInt64Coin("stake", 36_630_000_000)
	s.FundAcc(granter.Account.GetAddress(), sdk.NewCoins(fee))
	s.Require().NoError(s.App.AppKeepers.FeeGrantKeeper.GrantAllowance(
		s.Ctx,
		granter.Account.GetAddress(),
		grantee.Account.GetAddress(),
		&feegrant.BasicAllowance{SpendLimit: sdk.NewCoins(fee)},
	))

	handler, err := junoante.NewAnteHandler(junoante.HandlerOptions{
		HandlerOptions: authante.HandlerOptions{
			FeegrantKeeper:  s.App.AppKeepers.FeeGrantKeeper,
			SignModeHandler: s.App.TxConfig().SignModeHandler(),
		},
		AccountKeeper:         s.App.AppKeepers.AccountKeeper,
		BankKeeper:            s.App.AppKeepers.BankKeeper,
		StakingKeeper:         *s.App.AppKeepers.StakingKeeper,
		BondDenom:             "stake",
		IBCKeeper:             s.App.AppKeepers.IBCKeeper,
		TXCounterStoreService: runtime.NewKVStoreService(s.App.AppKeepers.GetKey(wasmtypes.StoreKey)),
		NodeConfig:            &wasmtypes.NodeConfig{},
		WasmKeeper:            &s.App.AppKeepers.WasmKeeper,
		FeemarketKeeper:       *s.App.AppKeepers.FeeMarketKeeper,
		FeepayKeeper:          s.App.AppKeepers.FeePayKeeper,
		FeeshareKeeper:        s.App.AppKeepers.FeeShareKeeper,
	})
	s.Require().NoError(err)

	txConfig := tx.NewTxConfig(codec.NewProtoCodec(s.App.InterfaceRegistry()), tx.DefaultSignModes)
	account := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, grantee.Account.GetAddress())
	signedTx, err := genTxWithFeeGranter(
		txConfig,
		[]sdk.Msg{testdata.NewTestMsg(grantee.Account.GetAddress())},
		sdk.NewCoins(fee),
		200_000,
		s.Ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		granter.Account.GetAddress(),
		grantee.Priv,
	)
	s.Require().NoError(err)

	before := s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, granter.Account.GetAddress(), fee.Denom)
	s.Require().NotPanics(func() {
		_, err = handler(s.Ctx, signedTx, false)
	})
	s.Require().NoError(err)
	after := s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, granter.Account.GetAddress(), fee.Denom)
	s.Require().True(after.Amount.Equal(before.Amount.Sub(fee.Amount)))
}

func (s *AnteTestSuite) TestFeegranterWithoutKeeperReturnsError() {
	s.SetupTest()

	grantee := s.fullAccs[0]
	granter := s.fullAccs[1]
	fee := sdk.NewInt64Coin("stake", 36_630_000_000)
	s.FundAcc(granter.Account.GetAddress(), sdk.NewCoins(fee))

	dfd := decorators.NewDeductFeeDecorator(
		s.App.AppKeepers.FeePayKeeper,
		*s.App.AppKeepers.FeeMarketKeeper,
		s.App.AppKeepers.AccountKeeper,
		s.App.AppKeepers.BankKeeper,
		nil,
		"stake",
		nil,
		nil,
	)
	handler := sdk.ChainAnteDecorators(dfd)

	txConfig := tx.NewTxConfig(codec.NewProtoCodec(s.App.InterfaceRegistry()), tx.DefaultSignModes)
	account := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, grantee.Account.GetAddress())
	signedTx, err := genTxWithFeeGranter(
		txConfig,
		[]sdk.Msg{testdata.NewTestMsg(grantee.Account.GetAddress())},
		sdk.NewCoins(fee),
		200_000,
		s.Ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		granter.Account.GetAddress(),
		grantee.Priv,
	)
	s.Require().NoError(err)

	s.Require().NotPanics(func() {
		_, err = handler(s.Ctx, signedTx, false)
	})
	s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)
}

func (s *AnteTestSuite) TestEscrowFunds() {
	// Slice (not map) for deterministic ordering. Several subtests
	// mutate FeeGrantKeeper state (GrantAllowance) and downstream cases
	// expect prior cases not to have run — under non-deterministic map
	// iteration that bug surfaces as an intermittent FAIL on
	// "no fee grant" finding the allowance from "valid fee grant".
	// Plus: SetupTest is invoked at the top of each subtest below
	// so state is fresh per case.
	type tcDef struct {
		name     string
		fee      int64
		valid    bool
		err      error
		malleate func(*AnteTestSuite) (signer testutil.TestAccount, feeAcc sdk.AccAddress)
	}
	cases := []tcDef{
		{
			name:  "paying with insufficient fee",
			fee:   1,
			valid: false,
			err:   sdkerrors.ErrInsufficientFee,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		{
			name:  "paying with good funds",
			fee:   24497000000,
			valid: true,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[0].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(24497000000))))

				return s.fullAccs[0], s.fullAccs[0].Account.GetAddress()
			},
		},
		{
			name:  "paying with no account",
			fee:   24497000000,
			valid: false,
			err:   sdkerrors.ErrUnknownAddress,
			malleate: func(_ *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				// Do not register the account
				priv, _, addr := testdata.KeyTestPubAddr()
				return testutil.TestAccount{
					Account: authtypes.NewBaseAccountWithAddress(addr),
					Priv:    priv,
				}, nil
			},
		},
		{
			name: "valid fee grant",
			// note: the original test said "valid fee grant with no account".
			// this is impossible given that feegrant.GrantAllowance calls
			// SetAccount for the grantee.
			fee:   36630000000,
			valid: true,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[1].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(36630000000))))
				err := s.App.AppKeepers.FeeGrantKeeper.GrantAllowance(
					s.Ctx,
					s.fullAccs[1].Account.GetAddress(),
					s.fullAccs[0].Account.GetAddress(),
					&feegrant.BasicAllowance{
						SpendLimit: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(36630000000))),
					},
				)
				s.Require().NoError(err)
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		{
			name:  "no fee grant",
			fee:   36630000000,
			valid: false,
			err:   sdkerrors.ErrNotFound,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		{
			name:  "allowance smaller than requested fee",
			fee:   36630000000,
			valid: false,
			err:   feegrant.ErrFeeLimitExceeded,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[1].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(36630000000))))
				err := s.App.AppKeepers.FeeGrantKeeper.GrantAllowance(
					s.Ctx,
					s.fullAccs[1].Account.GetAddress(),
					s.fullAccs[0].Account.GetAddress(),
					&feegrant.BasicAllowance{
						SpendLimit: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(20000000000))),
					},
				)
				s.Require().NoError(err)
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		{
			name:  "granter cannot cover allowed fee grant",
			fee:   36630000000,
			valid: false,
			err:   sdkerrors.ErrInsufficientFunds,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[3].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(20000000000))))
				err := s.App.AppKeepers.FeeGrantKeeper.GrantAllowance(
					s.Ctx,
					s.fullAccs[3].Account.GetAddress(),
					s.fullAccs[2].Account.GetAddress(),
					&feegrant.BasicAllowance{
						SpendLimit: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(36630000000))),
					},
				)
				s.Require().NoError(err)
				return s.fullAccs[2], s.fullAccs[3].Account.GetAddress()
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			// Reset suite state before each subtest so prior FeeGrantKeeper
			// mutations (or any other suite-scoped state) don't leak in.
			s.SetupTest()
			protoTxCfg := tx.NewTxConfig(codec.NewProtoCodec(s.App.InterfaceRegistry()), tx.DefaultSignModes)
			// this just tests our handler
			dfd := decorators.NewDeductFeeDecorator(
				s.App.AppKeepers.FeePayKeeper,
				*s.App.AppKeepers.FeeMarketKeeper,
				s.App.AppKeepers.AccountKeeper,
				s.App.AppKeepers.BankKeeper,
				s.App.AppKeepers.FeeGrantKeeper,
				"stake",
				nil,
				authante.NewDeductFeeDecorator(
					s.App.AppKeepers.AccountKeeper,
					s.App.AppKeepers.BankKeeper,
					s.App.AppKeepers.FeeGrantKeeper,
					nil,
				),
			)
			feeAnteHandler := sdk.ChainAnteDecorators(dfd)

			signer, feeAcc := tc.malleate(s)

			fee := sdk.NewCoins(sdk.NewInt64Coin("stake", tc.fee))
			msgs := []sdk.Msg{testdata.NewTestMsg(signer.Account.GetAddress())}

			acc := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, signer.Account.GetAddress())
			privs, accNums, seqs := []cryptotypes.PrivKey{signer.Priv}, []uint64{0}, []uint64{0}

			if acc != nil {
				accNums, seqs = []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}
			}

			var defaultGenTxGas uint64 = 10
			tx, err := genTxWithFeeGranter(protoTxCfg, msgs, fee, defaultGenTxGas, s.Ctx.ChainID(), accNums, seqs, feeAcc, privs...)
			s.Require().NoError(err)
			_, err = feeAnteHandler(s.Ctx, tx, false)
			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().ErrorIs(err, tc.err)
			}
		})
	}
}

func genTxWithFeeGranter(gen client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums,
	accSeqs []uint64, feeGranter sdk.AccAddress, priv ...cryptotypes.PrivKey,
) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode := signing.SignMode_SIGN_MODE_DIRECT

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	testTx := gen.NewTxBuilder()
	err := testTx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = testTx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	testTx.SetMemo(memo)
	testTx.SetFeeAmount(feeAmt)
	testTx.SetGasLimit(gas)
	testTx.SetFeeGranter(feeGranter)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        p.PubKey(),
		}
		signBytes, err := authsign.GetSignBytesAdapter(
			context.Background(), gen.SignModeHandler(), signMode, signerData, testTx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = testTx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return testTx.GetTx(), nil
}
