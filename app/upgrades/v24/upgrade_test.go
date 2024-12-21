package v24_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v26/app/apptesting"
	v24 "github.com/CosmosContracts/juno/v26/app/upgrades/v24"
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
	s.ConfirmUpgradeSucceeded(v24.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(_ *UpgradeTestSuite) {
}

func postUpgradeChecks(_ *UpgradeTestSuite) {
}
