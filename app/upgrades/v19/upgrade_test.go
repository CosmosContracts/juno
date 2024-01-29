package v19_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v19/app/apptesting"
	decorators "github.com/CosmosContracts/juno/v19/app/decorators"
	v19 "github.com/CosmosContracts/juno/v19/app/upgrades/v19"
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

	// Setup mainnet account
	_, err := v19.CreateMainnetVestingAccount(s.Ctx, s.App.AppKeepers)
	s.Require().NoError(err)

	// Create 3 generic validators
	val1 := s.SetupValidator(stakingtypes.Bonded)
	val2 := s.SetupValidator(stakingtypes.Bonded)
	val3 := s.SetupValidator(stakingtypes.Bonded)

	// Get last validator, set as mock jack validator
	val, found := s.App.AppKeepers.StakingKeeper.GetValidator(s.Ctx, val3)
	s.Require().True(found)
	v19.JackValidatorAddress = val.OperatorAddress

	validators := []sdk.ValAddress{val1, val2, val3}

	// Should equal 4, including default validator
	s.Require().Equal(4, len(s.App.AppKeepers.StakingKeeper.GetAllValidators(s.Ctx)))

	// Create delegations to each validator 2x, ensuring multiple delegations
	// are created for each validator and combined in state.
	for i := 0; i < 2; i++ {
		for _, delegator := range v19.Core1VestingAccounts {
			delegatorAddr := sdk.MustAccAddressFromBech32(delegator.Address)

			for _, validator := range validators {

				fmt.Println("Delegator: ", delegatorAddr.String())
				fmt.Println("Delegating to validator: ", validator.String())

				s.StakingHelper.Delegate(delegatorAddr, validator, sdk.NewInt(1_000_000))
			}
		}
	}

	preUpgradeChecks(s)

	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v19.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
	// Change Rate Decorator Test
	// Create a validator with a max change rate of 20%
	for i := 0; i < 100; i++ {

		// Create validator keys & desc
		valPub := secp256k1.GenPrivKey().PubKey()
		valAddr := sdk.ValAddress(valPub.Address())
		description := stakingtypes.NewDescription(fmt.Sprintf("test_moniker%d", i), "", "", "", "")
		validator, err := stakingtypes.NewValidator(
			valAddr,
			valPub,
			description,
		)
		s.Require().NoError(err)

		// Set validator commission
		changeRate := "1.00"
		if i < 100 {
			changeRate = fmt.Sprintf("0.%02d", i)
		}
		validator.Commission.MaxChangeRate.Set(sdk.MustNewDecFromStr(changeRate))

		// Set validator in kv store
		s.App.AppKeepers.StakingKeeper.SetValidator(s.Ctx, validator)
	}
}

func postUpgradeChecks(s *UpgradeTestSuite) {
	// Change Rate Decorator Test
	// Ensure all validators have a max change rate of 5%
	validators := s.App.AppKeepers.StakingKeeper.GetAllValidators(s.Ctx)
	for _, validator := range validators {
		s.Require().True(validator.Commission.MaxChangeRate.LTE(sdk.MustNewDecFromStr(decorators.MaxChangeRate)))
	}
}
