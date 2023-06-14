package antetest

import (
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v16/app"
	gaiafeeante "github.com/CosmosContracts/juno/v16/x/globalfee/ante"

	appparams "github.com/CosmosContracts/juno/v16/app/params"
	globfeetypes "github.com/CosmosContracts/juno/v16/x/globalfee/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app       *app.App
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

func (s *IntegrationTestSuite) SetupTest() {
	app := app.Setup(s.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: "testing",
		Height:  1,
	})

	encodingConfig := appparams.MakeEncodingConfig()
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	s.app = app
	s.ctx = ctx
	s.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
}

func (s *IntegrationTestSuite) SetupTestGlobalFeeStoreAndMinGasPrice(minGasPrice []sdk.DecCoin, globalFeeParams *globfeetypes.Params) (gaiafeeante.FeeDecorator, sdk.AnteHandler) {
	keeper := s.app.AppKeepers.GlobalFeeKeeper
	keeper.SetParams(s.ctx, *globalFeeParams)

	s.ctx = s.ctx.WithMinGasPrices(minGasPrice).WithIsCheckTx(true)

	// set staking params
	stakingParam := stakingtypes.DefaultParams()
	stakingParam.BondDenom = "uatom"
	sKeeper := s.app.AppKeepers.StakingKeeper
	sKeeper.SetParams(s.ctx, stakingParam)

	// build fee decorator
	feeDecorator := gaiafeeante.NewFeeDecorator(app.GetDefaultBypassFeeMessages(), keeper, *sKeeper, uint64(1_000_000))

	// chain fee decorator to antehandler
	antehandler := sdk.ChainAnteDecorators(feeDecorator)

	return feeDecorator, antehandler
}

// SetupTestStakingSubspace sets uatom as bond denom for the fee tests.
func (s *IntegrationTestSuite) SetupTestStakingKeeper(params stakingtypes.Params) *stakingkeeper.Keeper {
	sKeeper := s.app.AppKeepers.StakingKeeper
	sKeeper.SetParams(s.ctx, params)
	return sKeeper
}

func (s *IntegrationTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	if err := s.txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
			signerData,
			s.txBuilder,
			priv,
			s.clientCtx.TxConfig,
			accSeqs[i],
		)
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	if err := s.txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}
