package v16_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v17/app/apptesting"
	v17 "github.com/CosmosContracts/juno/v17/app/upgrades/v17"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// Ensures the test does not error out.
func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	preUpgradeChecks(s)

	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v17.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
	mp := s.App.AppKeepers.MintKeeper.GetParams(s.Ctx)
	s.Require().Equal(mp.BlocksPerYear, uint64(6311520))

	sp := s.App.AppKeepers.SlashingKeeper.GetParams(s.Ctx)
	s.Require().Equal(sp.SignedBlocksWindow, int64(100))
}

func postUpgradeChecks(s *UpgradeTestSuite) {
	// Ensure the mint params have doubled
	mp := s.App.AppKeepers.MintKeeper.GetParams(s.Ctx)
	s.Require().Equal(mp.BlocksPerYear, uint64(6311520*2))

	// Ensure the slashing params have doubled
	sp := s.App.AppKeepers.SlashingKeeper.GetParams(s.Ctx)
	s.Require().Equal(sp.SignedBlocksWindow, int64(100*2))
}
