package v31_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v31/app"
	v31 "github.com/CosmosContracts/juno/v31/app/upgrades/v31"
	"github.com/CosmosContracts/juno/v31/testutil"
)

type UpgradeTestSuite struct {
	testutil.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgradeDefinition() {
	s.Require().Equal("v31", v31.UpgradeName)
	s.Require().Equal(v31.UpgradeName, v31.Upgrade.UpgradeName)
	s.Require().NotNil(v31.Upgrade.CreateUpgradeHandler)

	s.Require().Empty(v31.Upgrade.StoreUpgrades.Added)
	s.Require().Empty(v31.Upgrade.StoreUpgrades.Renamed)
	s.Require().Empty(v31.Upgrade.StoreUpgrades.Deleted)
}

func (s *UpgradeTestSuite) TestUpgradeIsRegistered() {
	names := make(map[string]int)
	for _, u := range app.Upgrades {
		names[u.UpgradeName]++
	}
	s.Require().Equal(1, names[v31.UpgradeName], "v31 must be registered exactly once")
	s.Require().Equal(1, names["v30"], "historical v30 upgrade must stay registered")
}

func (s *UpgradeTestSuite) TestHandlerReturnsCurrentVersionMap() {
	s.Setup()

	versionMap, err := s.runV31Handler()
	s.Require().NoError(err)
	s.Require().Equal(s.App.ModuleManager.GetVersionMap(), versionMap)
}

func (s *UpgradeTestSuite) TestHandlerLeavesParamsUntouched() {
	s.Setup()
	k := s.App.AppKeepers

	feemarketBefore, err := k.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	cwHooksBefore, err := k.CWHooksKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	votingSnapshotBefore, err := k.VotingSnapshotKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	wasmBefore := k.WasmKeeper.GetParams(s.Ctx)

	_, err = s.runV31Handler()
	s.Require().NoError(err)

	feemarketAfter, err := k.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	cwHooksAfter, err := k.CWHooksKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	votingSnapshotAfter, err := k.VotingSnapshotKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)

	s.Require().Equal(feemarketBefore, feemarketAfter)
	s.Require().Equal(cwHooksBefore, cwHooksAfter)
	s.Require().Equal(votingSnapshotBefore, votingSnapshotAfter)
	s.Require().Equal(wasmBefore, k.WasmKeeper.GetParams(s.Ctx))
}

func (s *UpgradeTestSuite) TestHandlerIsIdempotent() {
	s.Setup()

	first, err := s.runV31Handler()
	s.Require().NoError(err)
	second, err := s.runV31Handler()
	s.Require().NoError(err)
	s.Require().Equal(first, second)
}

func (s *UpgradeTestSuite) runV31Handler() (module.VersionMap, error) {
	handler := v31.CreateV31UpgradeHandler(
		s.App.ModuleManager,
		module.NewConfigurator(s.App.AppCodec(), s.App.MsgServiceRouter(), s.App.GRPCQueryRouter()),
		&s.App.AppKeepers,
	)
	return handler(s.Ctx, upgradetypes.Plan{Name: v31.UpgradeName}, s.App.ModuleManager.GetVersionMap())
}
