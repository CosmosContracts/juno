package upgrades

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	minttypes "github.com/CosmosContracts/juno/v16/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/CosmosContracts/juno/v16/app/keepers"
)

func getUnVestedCoins(ctx sdk.Context, vacc *authvestingtypes.PeriodicVestingAccount, keepers *keepers.AppKeepers, bondDenom string) sdk.Coins {
	var mintAmt math.Int = sdk.ZeroInt()

	for i := range vacc.VestingPeriods {
		ujunoAmt := vacc.VestingPeriods[i].Amount.AmountOf(bondDenom)

		mintAmt = mintAmt.Add(ujunoAmt)
	}

	return sdk.NewCoins(sdk.NewCoin(bondDenom, mintAmt))
}

func clearVestingAccount(ctx sdk.Context, vacc *authvestingtypes.PeriodicVestingAccount, keepers *keepers.AppKeepers, bondDenom string) {
	// Finish vesting period now.
	vacc.EndTime = 0
	vacc.BaseVestingAccount.EndTime = 0

	for i := range vacc.VestingPeriods {
		vacc.VestingPeriods[i].Length = 0
	}

	vacc.VestingPeriods = nil
	vacc.BaseVestingAccount.DelegatedVesting = sdk.Coins{}

	keepers.AccountKeeper.SetAccount(ctx, vacc)
}

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
	}

	// Shows locked funds
	showLockedCoins(vacc, now)

	// Gets all coins which have not been unlocked yet.
	unvestedCoins := getUnVestedCoins(ctx, vacc, keepers, bondDenom)

	// Clears the account so all unvested coins are unlocked.
	clearVestingAccount(ctx, vacc, keepers, bondDenom)

	// Finish redeleations.
	if err := completeAllRedelegations(ctx, keepers, accAddr, now); err != nil {
		return err
	}

	// Instant unbond all delegations
	if err := unbondAllAndFinish(ctx, now, keepers, accAddr); err != nil {
		return err
	}

	// Moves the accounts held balance to the subDAO (ex: all unbonded tokens).
	movedBal, err := migrateBalanceToCore1SubDao(ctx, vacc, keepers, core1AccAddr, bondDenom)
	if err != nil {
		return err
	}
	fmt.Printf("movedBal: %v\n", movedBal)

	// mint unvested coins to be transfered.
	if err := keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, unvestedCoins); err != nil {
		return err
	}

	// transfer unvested coins back to the to core1 subdao
	if err := keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, core1AccAddr, unvestedCoins); err != nil {
		return err
	}

	// update core1 bal
	core1BalC := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)
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

func migrateBalanceToCore1SubDao(ctx sdk.Context, vacc *authvestingtypes.PeriodicVestingAccount, keepers *keepers.AppKeepers, core1AccAddr sdk.AccAddress, bondDenom string) (sdk.Coin, error) {
	// Get vesting account balance
	accbal := keepers.BankKeeper.GetBalance(ctx, vacc.GetAddress(), bondDenom)
	fmt.Printf("accbal: %v\n", accbal)

	// Send all tokens from balance to the core-1 subdao address
	if e := keepers.BankKeeper.SendCoins(ctx, vacc.GetAddress(), core1AccAddr, sdk.NewCoins(accbal)); e != nil {
		return sdk.Coin{}, fmt.Errorf("error sending coins: %v", e)
	}

	core1BalC := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)
	fmt.Printf("core1Bal: %v\n", core1BalC)

	return accbal, nil
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
