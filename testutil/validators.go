package testutil

import (
	"time"

	sdkmath "cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SetupValidator sets up a validator and returns the ValAddress.
func (s *KeeperTestHelper) SetupValidator(bondStatus stakingtypes.BondStatus) sdk.ValAddress {
	pks := simtestutil.CreateTestPubKeys(1)
	valPub := pks[0]
	valAddr := sdk.ValAddress(valPub.Bytes())
	stakingParams, err := s.App.AppKeepers.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	bondDenom := stakingParams.BondDenom
	bondAmt := sdk.DefaultPowerReduction
	selfBond := sdk.NewCoins(sdk.Coin{Amount: bondAmt, Denom: bondDenom})

	s.FundAcc(sdk.AccAddress(valAddr), selfBond)

	stakingCoin := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: selfBond[0].Amount}
	zeroDec := sdkmath.LegacyZeroDec()
	zeroCommission := stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
	valCreateMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddr.String(),
		valPub,
		stakingCoin,
		stakingtypes.Description{},
		zeroCommission,
		sdkmath.OneInt(),
	)
	s.Require().NoError(err)
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(s.App.AppKeepers.StakingKeeper)
	res, err := stakingMsgSvr.CreateValidator(s.Ctx, valCreateMsg)
	s.Require().NoError(err)
	s.Require().NotNil(res)

	val, err := s.App.AppKeepers.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	val = val.UpdateStatus(bondStatus)
	err = s.App.AppKeepers.StakingKeeper.SetValidator(s.Ctx, val)
	s.Require().NoError(err)

	consAddr, err := val.GetConsAddr()
	s.Suite.Require().NoError(err)

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		s.Ctx.BlockHeight(),
		0,
		time.Unix(0, 0),
		false,
		0,
	)
	err = s.App.AppKeepers.SlashingKeeper.SetValidatorSigningInfo(s.Ctx, consAddr, signingInfo)
	s.Require().NoError(err)

	return valAddr
}

// SetupMultipleValidators setups "numValidator" validators and returns their address in string
func (s *KeeperTestHelper) SetupMultipleValidators(numValidator int) []string {
	valAddrs := []string{}
	for i := 0; i < numValidator; i++ {
		valAddr := s.SetupValidator(stakingtypes.Bonded)
		valAddrs = append(valAddrs, valAddr.String())
	}
	return valAddrs
}
