package keeper_test

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	junoApp "github.com/CosmosContracts/juno/v12/app"
	appparams "github.com/CosmosContracts/juno/v12/app/params"
	"github.com/CosmosContracts/juno/v12/x/oracle/keeper"
	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

const (
	displayDenom string = appparams.DisplayDenom
	bondDenom    string = appparams.BondDenom
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *junoApp.App
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

const (
	initialPower = int64(10000000000)
)

func (s *IntegrationTestSuite) SetupTest() {
	require := s.Require()
	isCheckTx := false
	junoApp := junoApp.Setup(s.T(), isCheckTx, 1)

	/*  `Height:  9` because this check :
	if (uint64(ctx.BlockHeight())/params.VotePeriod)-(aggregatePrevote.SubmitBlock/params.VotePeriod) != 1 {
		return nil, types.ErrRevealPeriodMissMatch
	}
	*/
	ctx := junoApp.BaseApp.NewContext(isCheckTx, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  9,
	})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, junoApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(junoApp.OracleKeeper))

	sh := teststaking.NewHelper(s.T(), ctx, junoApp.StakingKeeper)
	sh.Denom = bondDenom
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)

	// mint and send coins to validators
	require.NoError(junoApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	require.NoError(junoApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))
	require.NoError(junoApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	require.NoError(junoApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr2, initCoins))

	sh.CreateValidator(valAddr, valPubKey, amt, true)
	sh.CreateValidator(valAddr2, valPubKey2, amt, true)

	staking.EndBlocker(ctx, junoApp.StakingKeeper)

	s.app = junoApp
	s.ctx = ctx
	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(junoApp.OracleKeeper)
}

// Test addresses
var (
	valPubKeys = simapp.CreateTestPubKeys(2)

	valPubKey = valPubKeys[0]
	pubKey    = secp256k1.GenPrivKey().PubKey()
	addr      = sdk.AccAddress(pubKey.Address())
	valAddr   = sdk.ValAddress(pubKey.Address())

	valPubKey2 = valPubKeys[1]
	pubKey2    = secp256k1.GenPrivKey().PubKey()
	addr2      = sdk.AccAddress(pubKey2.Address())
	valAddr2   = sdk.ValAddress(pubKey2.Address())

	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, initTokens))
)

// NewTestMsgCreateValidator test msg creator
func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amt sdk.Int) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	msg, _ := stakingtypes.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(types.JunoDenom, amt),
		stakingtypes.Description{}, commission, sdk.OneInt(),
	)

	return msg
}

func (s *IntegrationTestSuite) TestSetFeederDelegation() {
	app, ctx := s.app, s.ctx

	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := app.AccountKeeper.NewAccountWithAddress(ctx, feederAddr)
	app.AccountKeeper.SetAccount(ctx, feederAcc)

	err := s.app.OracleKeeper.ValidateFeeder(ctx, addr, valAddr)
	s.Require().NoError(err)
	err = s.app.OracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
	s.Require().Error(err)

	s.app.OracleKeeper.SetFeederDelegation(ctx, valAddr, feederAddr)

	err = s.app.OracleKeeper.ValidateFeeder(ctx, addr, valAddr)
	s.Require().Error(err)
	err = s.app.OracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestGetFeederDelegation() {
	app, ctx := s.app, s.ctx

	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := app.AccountKeeper.NewAccountWithAddress(ctx, feederAddr)
	app.AccountKeeper.SetAccount(ctx, feederAcc)

	s.app.OracleKeeper.SetFeederDelegation(ctx, valAddr, feederAddr)
	resp, err := app.OracleKeeper.GetFeederDelegation(ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Equal(resp, feederAddr)
}

func (s *IntegrationTestSuite) TestMissCounter() {
	app, ctx := s.app, s.ctx
	missCounter := uint64(rand.Intn(100))

	s.Require().Equal(app.OracleKeeper.GetMissCounter(ctx, valAddr), uint64(0))
	app.OracleKeeper.SetMissCounter(ctx, valAddr, missCounter)
	s.Require().Equal(app.OracleKeeper.GetMissCounter(ctx, valAddr), missCounter)

	app.OracleKeeper.DeleteMissCounter(ctx, valAddr)
	s.Require().Equal(app.OracleKeeper.GetMissCounter(ctx, valAddr), uint64(0))
}

func (s *IntegrationTestSuite) TestAggregateExchangeRatePrevote() {
	app, ctx := s.app, s.ctx

	prevote := types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}
	app.OracleKeeper.SetAggregateExchangeRatePrevote(ctx, valAddr, prevote)

	_, err := app.OracleKeeper.GetAggregateExchangeRatePrevote(ctx, valAddr)
	s.Require().NoError(err)

	app.OracleKeeper.DeleteAggregateExchangeRatePrevote(ctx, valAddr)

	_, err = app.OracleKeeper.GetAggregateExchangeRatePrevote(ctx, valAddr)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestAggregateExchangeRatePrevoteError() {
	app, ctx := s.app, s.ctx

	_, err := app.OracleKeeper.GetAggregateExchangeRatePrevote(ctx, valAddr)
	s.Require().Errorf(err, types.ErrNoAggregatePrevote.Error())
}

func (s *IntegrationTestSuite) TestAggregateExchangeRateVote() {
	app, ctx := s.app, s.ctx

	var tuples types.ExchangeRateTuples
	tuples = append(tuples, types.ExchangeRateTuple{
		Denom:        displayDenom,
		ExchangeRate: sdk.ZeroDec(),
	})

	vote := types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}
	app.OracleKeeper.SetAggregateExchangeRateVote(ctx, valAddr, vote)

	_, err := app.OracleKeeper.GetAggregateExchangeRateVote(ctx, valAddr)
	s.Require().NoError(err)

	app.OracleKeeper.DeleteAggregateExchangeRateVote(ctx, valAddr)

	_, err = app.OracleKeeper.GetAggregateExchangeRateVote(ctx, valAddr)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestAggregateExchangeRateVoteError() {
	app, ctx := s.app, s.ctx

	_, err := app.OracleKeeper.GetAggregateExchangeRateVote(ctx, valAddr)
	s.Require().Errorf(err, types.ErrNoAggregateVote.Error())
}

func (s *IntegrationTestSuite) TestSetExchangeRateWithEvent() {
	app, ctx := s.app, s.ctx
	err := app.OracleKeeper.SetExchangeRateWithEvent(ctx, displayDenom, sdk.OneDec())
	s.Require().NoError(err)
	rate, err := app.OracleKeeper.GetExchangeRate(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.OneDec())
}

func (s *IntegrationTestSuite) TestGetExchangeRate_InvalidDenom() {
	app, ctx := s.app, s.ctx

	_, err := app.OracleKeeper.GetExchangeRate(ctx, "uxyz")
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetExchangeRate_NotSet() {
	app, ctx := s.app, s.ctx

	_, err := app.OracleKeeper.GetExchangeRate(ctx, displayDenom)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestGetExchangeRate_Valid() {
	app, ctx := s.app, s.ctx

	app.OracleKeeper.SetExchangeRate(ctx, displayDenom, sdk.OneDec())
	rate, err := app.OracleKeeper.GetExchangeRate(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.OneDec())

	app.OracleKeeper.SetExchangeRate(ctx, strings.ToLower(displayDenom), sdk.OneDec())
	rate, err = app.OracleKeeper.GetExchangeRate(ctx, displayDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate, sdk.OneDec())
}

func (s *IntegrationTestSuite) TestGetExchangeRateBase() {
	oracleParams := s.app.OracleKeeper.GetParams(s.ctx)

	var exponent uint64
	for _, denom := range oracleParams.AcceptList {
		if denom.BaseDenom == bondDenom {
			exponent = uint64(denom.Exponent)
		}
	}

	power := sdk.MustNewDecFromStr("10").Power(exponent)

	s.app.OracleKeeper.SetExchangeRate(s.ctx, displayDenom, sdk.OneDec())
	rate, err := s.app.OracleKeeper.GetExchangeRateBase(s.ctx, bondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate.Mul(power), sdk.OneDec())

	s.app.OracleKeeper.SetExchangeRate(s.ctx, strings.ToLower(displayDenom), sdk.OneDec())
	rate, err = s.app.OracleKeeper.GetExchangeRateBase(s.ctx, bondDenom)
	s.Require().NoError(err)
	s.Require().Equal(rate.Mul(power), sdk.OneDec())
}

func (s *IntegrationTestSuite) TestClearExchangeRate() {
	app, ctx := s.app, s.ctx

	app.OracleKeeper.SetExchangeRate(ctx, displayDenom, sdk.OneDec())
	app.OracleKeeper.ClearExchangeRates(ctx)
	_, err := app.OracleKeeper.GetExchangeRate(ctx, displayDenom)
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestSetDenomPriceHistory() {
	exchangeRate := sdk.NewDec(1000)

	for _, tc := range []struct {
		desc         string
		symbolDenom  string
		exchangeRate sdk.Dec
		blockHeight  uint64
		shouldErr    bool
	}{
		{
			desc:         "Success case",
			symbolDenom:  types.JunoSymbol,
			exchangeRate: exchangeRate,
			blockHeight:  10,
			shouldErr:    false,
		},
		{
			desc:         "Invalid block height",
			symbolDenom:  types.JunoSymbol,
			exchangeRate: exchangeRate,
			blockHeight:  0,
			shouldErr:    true,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			if !tc.shouldErr {
				err := s.app.OracleKeeper.SetDenomPriceHistory(s.ctx, tc.symbolDenom, tc.exchangeRate, time.Now().UTC(), tc.blockHeight)
				s.Require().NoError(err)
			} else {
				err := s.app.OracleKeeper.SetDenomPriceHistory(s.ctx, tc.symbolDenom, tc.exchangeRate, time.Now().UTC(), tc.blockHeight)
				s.Require().Error(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetDenomPriceHistoryWithBlockHeight() {
	params := s.app.OracleKeeper.GetParams(s.ctx)

	var blockHeight uint64 = 10
	votePeriodCount := blockHeight / params.VotePeriod
	symbolDenom := types.JunoSymbol
	time := time.Now().UTC()
	err := s.app.OracleKeeper.SetDenomPriceHistory(s.ctx, symbolDenom, sdk.OneDec(), time, blockHeight)
	s.Require().NoError(err)
	price, err := s.app.OracleKeeper.GetDenomPriceHistoryWithBlockHeight(s.ctx, symbolDenom, 11)
	s.Require().NoError(err)
	s.Require().Equal(price.Price, sdk.OneDec())
	s.Require().Equal(price.PriceUpdateTime, time)
	s.Require().Equal(price.VotePeriodCount, votePeriodCount)
}

func (s *IntegrationTestSuite) TestIterateDenomPriceHistory() {
	params := s.app.OracleKeeper.GetParams(s.ctx)
	source := rand.NewSource(10)
	r := rand.New(source)
	symbolDenom := types.JunoSymbol
	blockHeights := randUInt64Array(10, r)

	var votePeriodCounts []uint64
	for _, blockHeight := range blockHeights {
		err := s.app.OracleKeeper.SetDenomPriceHistory(s.ctx, symbolDenom, sdk.OneDec(), time.Now().Add(time.Second*time.Duration(blockHeight)), blockHeight)
		s.Require().NoError(err)
		votePeriod := blockHeight / params.VotePeriod
		votePeriodCounts = append(votePeriodCounts, votePeriod)
	}

	var keys []uint64
	s.app.OracleKeeper.IterateDenomPriceHistory(s.ctx, symbolDenom, func(key uint64, priceHistoryEntry types.PriceHistoryEntry) bool {
		keys = append(keys, key)
		s.Require().Equal(priceHistoryEntry.Price, sdk.OneDec())
		return false
	})

	s.Require().Equal(len(votePeriodCounts), len(keys))
	for i, key := range keys {
		s.Require().Equal(key, votePeriodCounts[i])
	}
}

func (s *IntegrationTestSuite) TestDeleteDenomPriceHistory() {
	params := s.app.OracleKeeper.GetParams(s.ctx)

	var blockHeight uint64 = 10
	votePeriodCount := blockHeight / params.VotePeriod
	symbolDenom := types.JunoSymbol
	time := time.Now().UTC()
	err := s.app.OracleKeeper.SetDenomPriceHistory(s.ctx, symbolDenom, sdk.OneDec(), time, blockHeight)
	s.Require().NoError(err)
	price, err := s.app.OracleKeeper.GetDenomPriceHistoryWithBlockHeight(s.ctx, symbolDenom, 11)
	s.Require().NoError(err)
	s.Require().Equal(price.Price, sdk.OneDec())
	s.Require().Equal(price.PriceUpdateTime, time)
	s.Require().Equal(price.VotePeriodCount, votePeriodCount)

	s.app.OracleKeeper.DeleteDenomPriceHistory(s.ctx, symbolDenom, votePeriodCount)
	price, err = s.app.OracleKeeper.GetDenomPriceHistoryWithBlockHeight(s.ctx, symbolDenom, 11)
	s.Require().Error(err)
}

func randUInt64Array(number uint64, r *rand.Rand) []uint64 {
	var result []uint64
	for number > 0 {
		temp := r.Uint64()
		result = append(result, temp)
		number--
	}

	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
