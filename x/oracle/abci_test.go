package oracle_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	app "github.com/CosmosContracts/juno/v12/app"
	appparams "github.com/CosmosContracts/juno/v12/app/params"
	"github.com/CosmosContracts/juno/v12/x/oracle"
	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

const (
	displayDenom string = appparams.DisplayDenom
	bondDenom    string = appparams.BondDenom
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
}

const (
	initialPower = int64(10000000000)
)

func (s *IntegrationTestSuite) SetupTest() {
	require := s.Require()
	isCheckTx := false
	app := app.Setup(s.T(), isCheckTx, 1)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
	})

	oracle.InitGenesis(ctx, app.OracleKeeper, *types.DefaultGenesisState())

	sh := teststaking.NewHelper(s.T(), ctx, app.StakingKeeper)
	sh.Denom = bondDenom
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)

	// mint and send coins to validator
	require.NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	require.NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))

	sh.CreateValidator(valAddr, valPubKey, amt, true)

	staking.EndBlocker(ctx, app.StakingKeeper)

	s.app = app
	s.ctx = ctx
}

// Test addresses
var (
	valPubKeys = simapp.CreateTestPubKeys(1)

	valPubKey = valPubKeys[0]
	pubKey    = secp256k1.GenPrivKey().PubKey()
	addr      = sdk.AccAddress(pubKey.Address())
	valAddr   = sdk.ValAddress(pubKey.Address())

	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(bondDenom, initTokens))
)

var historacleTestCases = []struct {
	exchangeRates           []string
	expectedHistoricTWAPMax sdk.Dec
	expectedHistoricTWAPMin sdk.Dec
}{
	{
		[]string{
			"1.0", "1.2", "1.1", "1.4", "1.1", "1.15",
			"1.2", "1.3", "1.2", "1.12", "1.2", "1.15",
		},
		sdk.MustNewDecFromStr("1.18"), // upper bound for TWAP[0, 11]
		sdk.MustNewDecFromStr("1.16"), // lower bound for TWAP[0, 11]
	},
	{
		[]string{
			"2.3", "2.12", "2.14", "2.24", "2.18", "2.15",
			"2.51", "2.59", "2.67", "2.76", "2.89", "2.85",
		},
		sdk.MustNewDecFromStr("2.5"), // upper bound for TWAP[0, 11]
		sdk.MustNewDecFromStr("2.3"), // lower bound for TWAP[0, 11]
	},
}

func (s *IntegrationTestSuite) TestEndblockerHistoracle() {
	app, ctx := s.app, s.ctx

	for _, tc := range historacleTestCases {
		startTimeStamp := time.Now().UTC() // store timestamp before insertions

		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(app.OracleKeeper.GetParams(ctx).VotePeriod-1))
		ctx = ctx.WithBlockTime(time.Now().UTC())

		// check if last price is updated after each vote period
		for _, exchangeRate := range tc.exchangeRates {
			var tuples types.ExchangeRateTuples
			for _, denom := range app.OracleKeeper.Whitelist(ctx) {
				tuples = append(tuples, types.ExchangeRateTuple{
					Denom:        denom.SymbolDenom,
					ExchangeRate: sdk.MustNewDecFromStr(exchangeRate),
				})
			}

			prevote := types.AggregateExchangeRatePrevote{
				Hash:        "hash",
				Voter:       valAddr.String(),
				SubmitBlock: uint64(ctx.BlockHeight()),
			}
			app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr, prevote)
			err := oracle.EndBlocker(ctx, app.OracleKeeper)
			s.Require().NoError(err)

			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(app.OracleKeeper.VotePeriod(ctx)))
			ctx = ctx.WithBlockTime(time.Now().UTC())

			vote := types.AggregateExchangeRateVote{
				ExchangeRateTuples: tuples,
				Voter:              valAddr.String(),
			}
			app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr, vote)

			err = oracle.EndBlocker(ctx, app.OracleKeeper)
			s.Require().NoError(err)

			// historical prices query
			for _, denom := range app.OracleKeeper.Whitelist(ctx) {
				readExchangeRate, err := app.OracleKeeper.GetExchangeRate(ctx, denom.SymbolDenom)
				s.Require().NoError(err)
				s.Require().Equal(sdk.MustNewDecFromStr(exchangeRate), readExchangeRate)
			}
		}

		// historical TWAP
		for _, denom := range app.OracleKeeper.Whitelist(ctx) {
			// query for TWAP of all entered exchangeRates
			twap, err := app.OracleKeeper.GetArithmetricTWAP(ctx, denom.SymbolDenom, startTimeStamp, time.Now().UTC())
			s.Require().NoError(err)
			s.Require().True(tc.expectedHistoricTWAPMax.GTE(twap))
			s.Require().True(tc.expectedHistoricTWAPMin.LTE(twap))
		}

		ctx = ctx.WithBlockHeight(0)
		ctx = ctx.WithBlockTime(time.Now().UTC())
	}
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
