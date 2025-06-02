package ante_test

import (
	"context"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/CosmosContracts/juno/v30/app/decorators"
	"github.com/CosmosContracts/juno/v30/testutil"
)

func (s *AnteTestSuite) TestEscrowFunds() {
	cases := map[string]struct {
		fee      int64
		valid    bool
		err      error
		malleate func(*AnteTestSuite) (signer testutil.TestAccount, feeAcc sdk.AccAddress)
	}{
		"paying with insufficient fee": {
			fee:   1,
			valid: false,
			err:   sdkerrors.ErrInsufficientFee,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		"paying with good funds": {
			fee:   24497000000,
			valid: true,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[0].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(24497000000))))

				return s.fullAccs[0], s.fullAccs[0].Account.GetAddress()
			},
		},
		"paying with no account": {
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
		"valid fee grant": {
			// note: the original test said "valid fee grant with no account".
			// this is impossible given that feegrant.GrantAllowance calls
			// SetAccount for the grantee.
			fee:   36630000000,
			valid: true,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[1].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(36630000000))))
				err := s.App.AppKeepers.FeeGrantKeeper.GrantAllowance(
					s.Ctx,
					s.fullAccs[1].Account.GetAddress(),
					s.fullAccs[0].Account.GetAddress(),
					&feegrant.BasicAllowance{
						SpendLimit: sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(36630000000))),
					},
				)
				s.Require().NoError(err)
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		"no fee grant": {
			fee:   36630000000,
			valid: false,
			err:   sdkerrors.ErrNotFound,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		"allowance smaller than requested fee": {
			fee:   36630000000,
			valid: false,
			err:   feegrant.ErrFeeLimitExceeded,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[1].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(36630000000))))
				err := s.App.AppKeepers.FeeGrantKeeper.GrantAllowance(
					s.Ctx,
					s.fullAccs[1].Account.GetAddress(),
					s.fullAccs[0].Account.GetAddress(),
					&feegrant.BasicAllowance{
						SpendLimit: sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(20000000000))),
					},
				)
				s.Require().NoError(err)
				return s.fullAccs[0], s.fullAccs[1].Account.GetAddress()
			},
		},
		"granter cannot cover allowed fee grant": {
			fee:   36630000000,
			valid: false,
			err:   sdkerrors.ErrInsufficientFunds,
			malleate: func(s *AnteTestSuite) (testutil.TestAccount, sdk.AccAddress) {
				s.FundAcc(s.fullAccs[3].Account.GetAddress(), sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(20000000000))))
				err := s.App.AppKeepers.FeeGrantKeeper.GrantAllowance(
					s.Ctx,
					s.fullAccs[3].Account.GetAddress(),
					s.fullAccs[2].Account.GetAddress(),
					&feegrant.BasicAllowance{
						SpendLimit: sdk.NewCoins(sdk.NewCoin("ujuno", math.NewInt(36630000000))),
					},
				)
				s.Require().NoError(err)
				return s.fullAccs[2], s.fullAccs[3].Account.GetAddress()
			},
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		s.Run(name, func() {
			protoTxCfg := tx.NewTxConfig(codec.NewProtoCodec(s.App.InterfaceRegistry()), tx.DefaultSignModes)
			// this just tests our handler
			dfd := decorators.NewDeductFeeDecorator(
				s.App.AppKeepers.FeePayKeeper,
				*s.App.AppKeepers.FeeMarketKeeper,
				s.App.AppKeepers.AccountKeeper,
				s.App.AppKeepers.BankKeeper,
				s.App.AppKeepers.FeeGrantKeeper,
				"ujuno",
				authante.NewDeductFeeDecorator(
					s.App.AppKeepers.AccountKeeper,
					s.App.AppKeepers.BankKeeper,
					s.App.AppKeepers.FeeGrantKeeper,
					nil,
				),
			)
			feeAnteHandler := sdk.ChainAnteDecorators(dfd)

			signer, feeAcc := stc.malleate(s)

			fee := sdk.NewCoins(sdk.NewInt64Coin("ujuno", tc.fee))
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

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)
	tx.SetFeeGranter(feeGranter)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        p.PubKey(),
		}
		signBytes, err := authsign.GetSignBytesAdapter(
			context.Background(), gen.SignModeHandler(), signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}
