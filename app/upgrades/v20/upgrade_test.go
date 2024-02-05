package v19_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v19/app/apptesting"
	v20 "github.com/CosmosContracts/juno/v19/app/upgrades/v20"
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

	// == CREATE MOCK CORE-1 BASE ACCOUNT ==
	acc := sdk.MustAccAddressFromBech32(v20.Core1MultisigVestingAccount)

	// set account and mint it some balance
	s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, authtypes.NewBaseAccount(acc, nil, 0, 0))

	amt := int64(9406347457268)
	s.App.AppKeepers.BankKeeper.MintCoins(s.Ctx, "mint", sdk.NewCoins(sdk.NewInt64Coin("ujuno", amt)))
	s.App.AppKeepers.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, "mint", acc, sdk.NewCoins(sdk.NewInt64Coin("ujuno", amt)))

	accBal := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, acc)
	fmt.Printf("Core1 bal: %s\n", accBal)

	// == UPGRADE ==
	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v20.UpgradeName, upgradeHeight)
	postUpgradeChecks(s)

	// == POST VERIFICATION ==
	charterBal := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(v20.CharterCouncil))
	fmt.Printf("Council Post Upgrade Balance: %s\n", charterBal)

	s.Require().True(charterBal.AmountOf("ujuno").GTE(accBal.AmountOf("ujuno")))
}

func preUpgradeChecks(s *UpgradeTestSuite) {
}

func postUpgradeChecks(s *UpgradeTestSuite) {
}
