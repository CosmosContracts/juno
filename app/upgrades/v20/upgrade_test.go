package v20_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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

	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v20.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
	// Setup mainnet account
	_, err := v20.CreateMainnetVestingAccount(s.Ctx, s.App.AppKeepers)
	s.Require().NoError(err)

	// Create 3 generic validators
	val1 := s.SetupValidator(stakingtypes.Bonded)
	val2 := s.SetupValidator(stakingtypes.Bonded)
	val3 := s.SetupValidator(stakingtypes.Bonded)

	// Get last validator, set as mock jack validator
	val, found := s.App.AppKeepers.StakingKeeper.GetValidator(s.Ctx, val3)
	s.Require().True(found)
	v20.JackValidatorAddress = val.OperatorAddress

	validators := []sdk.ValAddress{val1, val2, val3}

	// Should equal 4, including default validator
	s.Require().Equal(4, len(s.App.AppKeepers.StakingKeeper.GetAllValidators(s.Ctx)))

	// Create delegations to each validator 2x, ensuring multiple delegations
	// are created for each validator and combined in state.
	for i := 0; i < 2; i++ {
		for _, delegator := range v20.Core1VestingAccounts {
			delegatorAddr := sdk.MustAccAddressFromBech32(delegator.Address)

			for _, validator := range validators {
				s.StakingHelper.Delegate(delegatorAddr, validator, sdk.NewInt(1_000_000))
			}
		}
	}
}

func postUpgradeChecks(_ *UpgradeTestSuite) {
}
