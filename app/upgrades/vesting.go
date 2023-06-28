package upgrades

import (
	"fmt"
	"time"

	"github.com/CosmosContracts/juno/v16/app/keepers"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// TODO: Redelegations check as well.
func undelegate(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress, vacc *authvestingtypes.ContinuousVestingAccount) error {
	// Unbond all delegations from the account
	delegations := keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr)

	for _, delegation := range delegations {
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}

		// fmt.Printf("delegation: %s\n", delegation)
		// fmt.Printf("validatorValAddr: %s\n", validatorValAddr)

		// vacc.TrackUndelegation(vacc.DelegatedVesting) ??
		time, err := (*keepers.StakingKeeper).Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
		if err != nil {
			return err
		}
		fmt.Printf("time: %s and err:%v\n", time, err)
	}

	unbonding := keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr)
	// fmt.Printf("unbonding: %s\n", unbonding)

	for _, unbondingDelegation := range unbonding {
		validatorStringAddr := unbondingDelegation.ValidatorAddress
		validatorValAddr, _ := sdk.ValAddressFromBech32(validatorStringAddr)

		// Complete unbonding delegation
		for i := range unbondingDelegation.Entries {
			unbondingDelegation.Entries[i].CompletionTime = now
		}

		keepers.StakingKeeper.SetUnbondingDelegation(ctx, unbondingDelegation)
		_, err := keepers.StakingKeeper.CompleteUnbonding(ctx, accAddr, validatorValAddr)
		if err != nil {
			return err
		}
	}

	// check there are no more delegations.
	delegations = keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr)
	if len(delegations) > 0 {
		panic("delegations not empty")
	}

	return nil
}

func checkLockedCoins(vacc *authvestingtypes.ContinuousVestingAccount, now time.Time) sdk.Coins {
	locked := vacc.LockedCoins(now)
	lockedFromVesting := vacc.LockedCoinsFromVesting(vacc.GetVestingCoins(now))
	fmt.Printf("locked: %v\n", locked)
	fmt.Printf("lockedVesting: %v\n", lockedFromVesting)
	return locked
}

// Stops a vesting account and returns all tokens back to the Core-1 subdao.
func MoveVestingCoinFromVestingAccount(ctx sdk.Context, accAddr sdk.AccAddress, keepers *keepers.AppKeepers, core1SubDaoAddress string, bondDenom string) (sdk.Coin, error) {
	// var err error

	now := ctx.BlockHeader().Time

	core1AccAddr := sdk.MustAccAddressFromBech32(core1SubDaoAddress)

	stdAcc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	vacc, ok := stdAcc.(*authvestingtypes.ContinuousVestingAccount)
	if !ok {
		panic("not a ContinuousVestingAccount vesting account")
	}

	// should show numbers
	l := checkLockedCoins(vacc, now)
	fmt.Printf("l: %v\n", l)

	// Finish vesting period now.
	vacc.EndTime = 1
	vacc.BaseVestingAccount.EndTime = 1
	keepers.AccountKeeper.SetAccount(ctx, vacc)

	// Instant unbond all tokens, goes into balance.
	if err := undelegate(ctx, now, keepers, accAddr, vacc); err != nil {
		return sdk.Coin{}, err
	}

	// Get balance
	accbal := keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	fmt.Printf("bal: %v\n", accbal)

	// Send all tokens from balance to the core-1 subdao address
	if e := keepers.BankKeeper.SendCoins(ctx, accAddr, core1AccAddr, sdk.NewCoins(accbal)); e != nil {
		return sdk.Coin{}, fmt.Errorf("error sending coins: %v", e)
	}

	// get bal of core1SubDaoAddress
	core1BalC := keepers.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(core1SubDaoAddress), bondDenom)
	fmt.Printf("core1Bal: %v\n", core1BalC)

	// get balance of accAddr
	accbal = keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	fmt.Printf("bal: %v\n", accbal)

	// TODO: Delete said account? (no reason to have it or the abse account anymore yea? any issues of doing this?)
	// if so, do we have to remove all the subAccounts first of the vacc/
	keepers.AccountKeeper.RemoveAccount(ctx, vacc)

	// return sdk.Coin{}, fmt.Errorf("not implemented MoveVestingCoinFromVestAccount")
	return sdk.Coin{}, nil
}
