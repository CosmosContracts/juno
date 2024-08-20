package v23_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v24/app/apptesting"
	v23 "github.com/CosmosContracts/juno/v24/app/upgrades/v23"
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
	s.ConfirmUpgradeSucceeded(v23.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(_ *UpgradeTestSuite) {
}

func postUpgradeChecks(_ *UpgradeTestSuite) {
}
