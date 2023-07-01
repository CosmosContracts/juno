package upgrades

import (
	"fmt"
	"time"

	minttypes "github.com/CosmosContracts/juno/v16/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/CosmosContracts/juno/v16/app/keepers"
)

// Stops a vesting account and returns all tokens back to the Core-1 subdao.
func MoveVestingCoinFromVestingAccount(ctx sdk.Context, accAddr sdk.AccAddress, keepers *keepers.AppKeepers, core1SubDaoAddress string, bondDenom string) error {
	now := ctx.BlockHeader().Time

	core1AccAddr := sdk.MustAccAddressFromBech32(core1SubDaoAddress)

	stdAcc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	vacc, ok := stdAcc.(*authvestingtypes.PeriodicVestingAccount)
	if !ok {
		// For e2e testing
		fmt.Printf("account " + accAddr.String() + " is not a vesting account. This should not run on mainnet.\n")
		return nil
		// return fmt.Errorf("account " + accAddr.String() + " is not a vesting account")
	}

	// Shows locked funds
	showLockedCoins(vacc, now)

	// var expectedCoins sdk.Coin

	// // account balance, vesting amounts
	// accBal := keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	// vaccBal := vacc.GetVestingCoins(now)
	// expectedCoins = expectedCoins.Add(accBal).Add(vaccBal...)

	// Finish vesting period now.
	vacc.EndTime = 0
	vacc.BaseVestingAccount.EndTime = 0

	// Since these periods have not yet vested, we are un-vesting and then minting to Core1 for future use.
	ujunoAmt := 0
	for i := range vacc.VestingPeriods {
		vacc.VestingPeriods[i].Length = 0
		ujunoAmt += int(vacc.VestingPeriods[i].Amount.AmountOf("ujuno").Int64())
	}
	vacc.VestingPeriods = nil
	vacc.BaseVestingAccount.DelegatedVesting = sdk.Coins{}

	keepers.AccountKeeper.SetAccount(ctx, vacc)

	// Set it so any re-delegations are finished.
	if err := completeAllRedelegations(ctx, keepers, accAddr, now); err != nil {
		return err
	}

	// Instant unbond all delegations
	// TODO: What about rewards?
	if err := unbondAllAndFinish(ctx, now, keepers, accAddr); err != nil {
		return err
	}

	// Get balance
	accbal := keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	fmt.Printf("bal: %v\n", accbal)

	// Send all tokens from balance to the core-1 subdao address
	if e := keepers.BankKeeper.SendCoins(ctx, accAddr, core1AccAddr, sdk.NewCoins(accbal)); e != nil {
		return fmt.Errorf("error sending coins: %v", e)
	}

	// get bal of core1SubDaoAddress
	core1BalC := keepers.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(core1SubDaoAddress), bondDenom)
	fmt.Printf("core1Bal: %v\n", core1BalC)

	// get balance of accAddr
	accbal = keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	fmt.Printf("bal: %v\n", accbal)

	// mint ujunoAmt to core1SubDaoAddress (unvested amounts)
	tfrAmt := sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(int64(ujunoAmt))))
	if err := keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, tfrAmt); err != nil {
		return err
	}
	// transfer tfrAmt to core1 subdao
	if err := keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, core1AccAddr, tfrAmt); err != nil {
		return err
	}

	// update core1 bal
	core1BalC = keepers.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(core1SubDaoAddress), bondDenom)
	fmt.Printf("core1Bal: %v\n", core1BalC)

	// delete vacc
	// keepers.AccountKeeper.RemoveAccount(ctx, vacc)

	return fmt.Errorf("not implemented MoveVestingCoinFromVestAccount")

	// TODO: Delete said account? (no reason to have it or the base account anymore yea? any issues of doing this?)
	// if so, do we have to remove all the subAccounts first of the vacc/
	// keepers.AccountKeeper.RemoveAccount(ctx, vacc)

	// return sdk.Coin{}, fmt.Errorf("not implemented MoveVestingCoinFromVestAccount")
	return nil
}

func completeAllRedelegations(ctx sdk.Context, keepers *keepers.AppKeepers, accAddr sdk.AccAddress, now time.Time) error {
	for _, activeRedelegation := range keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535) {
		redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
		redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

		// set all entry completionTime to now so we can complete re-delegation
		for i := range activeRedelegation.Entries {
			activeRedelegation.Entries[i].CompletionTime = now
		}

		keepers.StakingKeeper.SetRedelegation(ctx, activeRedelegation)
		_, err := keepers.StakingKeeper.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
		if err != nil {
			return err
		}
	}

	return nil
}

func unbondAllAndFinish(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) error {
	// Unbond all delegations from the account
	for _, delegation := range keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr) {
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}

		time, err := keepers.StakingKeeper.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
		if err != nil {
			return err
		}
		fmt.Printf("time: %s and err:%v\n", time, err)
	}

	// Take all unbonding and complete them.
	for _, unbondingDelegation := range keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr) {
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

	return nil
}

func showLockedCoins(vacc *authvestingtypes.PeriodicVestingAccount, now time.Time) {
	lockedFromVesting := vacc.LockedCoinsFromVesting(vacc.GetVestingCoins(now))
	fmt.Printf("lockedVesting: %v\n", lockedFromVesting)
}
