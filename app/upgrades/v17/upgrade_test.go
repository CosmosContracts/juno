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

	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v17.UpgradeName, upgradeHeight)
}
