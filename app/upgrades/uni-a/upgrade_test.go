package unia_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/app/apptesting"
	unia "github.com/CosmosContracts/juno/v17/app/upgrades/uni-a"
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
	s.ConfirmUpgradeSucceeded(unia.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
	bal := getAccBalance(s, unia.ReeceBech32)
	s.Require().Equal(bal, sdk.NewCoins())
}

func postUpgradeChecks(s *UpgradeTestSuite) {
	bal := getAccBalance(s, unia.ReeceBech32)
	// 100m tokens
	s.Require().Equal(bal, sdk.NewCoins(sdk.NewCoin("ujunox", sdk.NewInt(100_000_000*1_000_000))))
}

func getAccBalance(s *UpgradeTestSuite, bech32 string) sdk.Coins {
	addr := sdk.MustAccAddressFromBech32(bech32)
	return s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, addr)
}
