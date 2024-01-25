package v19_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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

	// Create Strange Love Validator
	pub, err := codectypes.NewAnyWithValue(secp256k1.GenPrivKey().PubKey())
	s.Require().NoError(err)
	slVal := stakingtypes.Validator{
		OperatorAddress: "junovaloper130mdu9a0etmeuw52qfxk73pn0ga6gawk2tz77l",
		ConsensusPubkey: pub,
		Status:          stakingtypes.Bonded,
		Tokens:          sdk.NewInt(237205439625),
		DelegatorShares: sdk.NewDec(237205439625),
		Description: stakingtypes.Description{
			Moniker:         "strangelove",
			Identity:        "158DA6C7FCFB7BD23988D9C0D0D8B80F1C5C70B5",
			Website:         "",
			SecurityContact: "",
			Details:         "",
		},
		Commission: stakingtypes.Commission{
			CommissionRates: stakingtypes.CommissionRates{
				Rate:          sdk.MustNewDecFromStr("0.100000000000000000"),
				MaxRate:       sdk.MustNewDecFromStr("0.200000000000000000"),
				MaxChangeRate: sdk.MustNewDecFromStr("0.010000000000000000"),
			},
			UpdateTime: time.Date(2021, 10, 1, 15, 0, 0, 0, time.UTC),
		},
		MinSelfDelegation:       sdk.NewInt(1),
		Jailed:                  false,
		UnbondingHeight:         0,
		UnbondingTime:           time.Time{},
		UnbondingOnHoldRefCount: 0,
		UnbondingIds:            nil,
	}

	// err = json.Unmarshal([]byte(slValidator), &slVal)
	// s.Require().NoError(err)
	s.App.AppKeepers.StakingKeeper.SetValidator(s.Ctx, slVal)

	// Create additional validators
	val1 := s.SetupValidator(stakingtypes.Bonded)
	val2 := s.SetupValidator(stakingtypes.Bonded)
	val3 := s.SetupValidator(stakingtypes.Bonded)

	validators := []sdk.ValAddress{val1, val2, val3}

	// Should equal 5, including default validator
	s.Require().Equal(5, len(s.App.AppKeepers.StakingKeeper.GetAllValidators(s.Ctx)))

	// Create delegations to each validator
	for _, delegator := range v19.Core1VestingAccounts {
		delegatorAddr := sdk.MustAccAddressFromBech32(delegator.Address)

		for _, validator := range validators {

			fmt.Println("Delegator: ", delegatorAddr.String())
			fmt.Println("Delegating to validator: ", validator.String())

			s.StakingHelper.Delegate(delegatorAddr, validator, sdk.NewInt(1_000_000))
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
