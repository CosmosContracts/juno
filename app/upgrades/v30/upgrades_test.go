package v30_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	v30 "github.com/CosmosContracts/juno/v30/app/upgrades/v30"
	"github.com/CosmosContracts/juno/v30/testutil"
)

type UpgradeTestSuite struct {
	testutil.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestConfigureFeemarketParamsSeedsStateFromConsensusAndStaking() {
	s.Setup()
	s.Ctx = s.Ctx.WithBlockHeight(123)

	// Deliberately different from fallbackMaxBlockUtilization (25M) so this
	// test fails if the handler silently falls back instead of reading the
	// consensus params.
	const maxGas int64 = 30_000_000
	s.Require().NoError(s.App.AppKeepers.ConsensusParamsKeeper.ParamsStore.Set(s.Ctx, cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{MaxGas: maxGas},
	}))

	_, err := s.runV30Handler()
	s.Require().NoError(err)

	stakingParams, err := s.App.AppKeepers.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)

	params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().True(params.Enabled)
	s.Require().Equal(stakingParams.BondDenom, params.FeeDenom)
	s.Require().Equal(uint64(maxGas), params.MaxBlockUtilization)
	s.Require().Equal(math.LegacyMustNewDecFromStr("0.075"), params.MinBaseGasPrice)

	state, err := s.App.AppKeepers.FeeMarketKeeper.GetState(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(params.Window, uint64(len(state.Window)))
	s.Require().Equal(params.MinBaseGasPrice, state.BaseGasPrice)
	s.Require().Equal(params.MinLearningRate, state.LearningRate)

	enabledHeight, err := s.App.AppKeepers.FeeMarketKeeper.GetEnabledHeight(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(s.Ctx.BlockHeight(), enabledHeight)
}

func (s *UpgradeTestSuite) TestV30HandlerFallsBackForUnboundedMaxGas() {
	s.Setup()
	s.Require().NoError(s.App.AppKeepers.ConsensusParamsKeeper.ParamsStore.Set(s.Ctx, cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{MaxGas: -1},
	}))

	_, err := s.runV30Handler()
	s.Require().NoError(err)

	params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(25_000_000), params.MaxBlockUtilization)
}

func (s *UpgradeTestSuite) TestConfigureCWHooksParamsSetsFailureThreshold() {
	s.Setup()

	_, err := s.runV30Handler()
	s.Require().NoError(err)

	params, err := s.App.AppKeepers.CWHooksKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), params.ContractFailureRemovalThreshold)
}

func (s *UpgradeTestSuite) runV30Handler() (module.VersionMap, error) {
	handler := v30.CreateV30UpgradeHandler(
		s.App.ModuleManager,
		module.NewConfigurator(s.App.AppCodec(), s.App.MsgServiceRouter(), s.App.GRPCQueryRouter()),
		&s.App.AppKeepers,
	)
	return handler(s.Ctx, upgradetypes.Plan{Name: v30.UpgradeName}, s.App.ModuleManager.GetVersionMap())
}
