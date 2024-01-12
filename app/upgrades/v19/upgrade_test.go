package v19_test

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
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

func (s *UpgradeTestSuite) NextBlock(amt int) {
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + int64(amt))
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})
}

// Ensures the test does not error out.
func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()
	preUpgradeChecks(s)

	// Core-1 Multisig mock
	c1m, unvested := v19.CreateMainnetVestingAccount(s.Ctx, s.App.AppKeepers)
	c1mAddr := c1m.GetAddress()
	fmt.Printf("c1mAddr unvested: %+v\n", unvested)

	core1Prebal := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, c1mAddr)
	fmt.Printf("Core1 bal: %s\n", core1Prebal)

	// create many validators to confirm the unbonding code works
	newVal1 := s.SetupValidator(stakingtypes.Bonded)
	newVal2 := s.SetupValidator(stakingtypes.Bonded)
	newVal3 := s.SetupValidator(stakingtypes.Bonded)

	// Delegate tokens of the core1 multisig account
	s.StakingHelper.Delegate(c1mAddr, newVal1, sdk.NewInt(1))
	s.StakingHelper.Delegate(c1mAddr, newVal2, sdk.NewInt(2))
	s.StakingHelper.Delegate(c1mAddr, newVal3, sdk.NewInt(3))
	s.Require().Equal(3, len(s.App.AppKeepers.StakingKeeper.GetAllDelegatorDelegations(s.Ctx, c1mAddr)))

	// == UPGRADE ==
	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v19.UpgradeName, upgradeHeight)
	postUpgradeChecks(s)

	// == POST VERIFICATION ==
	updatedAcc := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, c1mAddr)
	_, ok := updatedAcc.(*vestingtypes.PeriodicVestingAccount)
	s.Require().False(ok)

	s.Require().Equal(0, len(s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, c1mAddr)))
	s.Require().Equal(0, len(s.App.AppKeepers.StakingKeeper.GetAllDelegatorDelegations(s.Ctx, c1mAddr)))

	charterBal := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(v19.CharterCouncil))
	fmt.Printf("Council Post Upgrade Balance: %s\n", charterBal) // this should be >= unvested + balance before the upgrade
	s.Require().True(charterBal.AmountOf("ujuno").GTE(unvested.Add(core1Prebal.AmountOf("ujuno"))))
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
