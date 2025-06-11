package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	"github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper
	encCfg           moduletestutil.TestEncodingConfig
	authorityAccount sdk.AccAddress

	msgServer   types.MsgServer
	queryServer types.QueryServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.encCfg = moduletestutil.MakeTestEncodingConfig()
	s.authorityAccount = authtypes.NewModuleAddress(govtypes.ModuleName)

	s.queryServer = keeper.NewQueryServer(*s.App.AppKeepers.FeeMarketKeeper)
	s.msgServer = keeper.NewMsgServer(s.App.AppKeepers.FeeMarketKeeper)

	s.App.AppKeepers.FeeMarketKeeper.SetEnabledHeight(s.Ctx, -1)
}

func (s *KeeperTestSuite) TestState() {
	s.Run("set and get default eip1559 state", func() {
		state := types.DefaultState()

		err := s.App.AppKeepers.FeeMarketKeeper.SetState(s.Ctx, state)
		s.Require().NoError(err)

		gotState, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(state, gotState)
	})

	s.Run("set and get aimd eip1559 state", func() {
		state := types.DefaultAIMDState()

		err := s.App.AppKeepers.FeeMarketKeeper.SetState(s.Ctx, state)
		s.Require().NoError(err)

		gotState, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
		s.Require().NoError(err)

		s.Require().Equal(state, gotState)
	})
}

func (s *KeeperTestSuite) TestParams() {
	s.Run("set and get default params", func() {
		params := types.DefaultParams()

		err := s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, params)
		s.Require().NoError(err)

		gotParams, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(params, gotParams)
	})

	s.Run("set and get custom params", func() {
		params := types.Params{
			Alpha:               math.LegacyMustNewDecFromStr("0.1"),
			Beta:                math.LegacyMustNewDecFromStr("0.1"),
			Gamma:               math.LegacyMustNewDecFromStr("0.1"),
			Delta:               math.LegacyMustNewDecFromStr("0.1"),
			MinBaseGasPrice:     math.LegacyNewDec(10),
			MinLearningRate:     math.LegacyMustNewDecFromStr("0.1"),
			MaxLearningRate:     math.LegacyMustNewDecFromStr("0.1"),
			MaxBlockUtilization: 10,
			Window:              1,
			Enabled:             true,
		}

		err := s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, params)
		s.Require().NoError(err)

		gotParams, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(params, gotParams)
	})
}

func (s *KeeperTestSuite) TestEnabledHeight() {
	s.Run("get and set values", func() {
		s.App.AppKeepers.FeeMarketKeeper.SetEnabledHeight(s.Ctx, 10)

		got, err := s.App.AppKeepers.FeeMarketKeeper.GetEnabledHeight(s.Ctx)
		s.Require().NoError(err)
		s.Require().Equal(int64(10), got)
	})
}
