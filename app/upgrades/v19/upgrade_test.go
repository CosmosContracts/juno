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
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
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

func (s *UpgradeTestSuite) setupMockCore1MultisigAccount() *vestingtypes.PeriodicVestingAccount {
	core1Multisig := v19.CreateMainnetVestingAccount()
	s.App.AppKeepers.AccountKeeper.SetAccount(s.Ctx, core1Multisig)
	return core1Multisig
}

// Ensures the test does not error out.
func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()
	preUpgradeChecks(s)

	// Core-1 account mock up
	va := s.setupMockCore1MultisigAccount()
	fmt.Println(va)

	// delegate to a validator
	valAddr1 := s.SetupValidator(stakingtypes.Bonded)
	valAddr2 := s.SetupValidator(stakingtypes.Bonded)
	valAddr3 := s.SetupValidator(stakingtypes.Bonded)

	// delegate to a validator
	s.DelegateToValidator(valAddr1, 1000000000)

	// upgrade
	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v19.UpgradeName, upgradeHeight)
	postUpgradeChecks(s)

	// TODO: check it was modified
	updatedAcc := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, va.GetAddress())
	fmt.Println(updatedAcc)
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
