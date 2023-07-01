package upgrades

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/CosmosContracts/juno/v16/app/keepers"
	minttypes "github.com/CosmosContracts/juno/v16/x/mint/types"
)

// Stops a vesting account and returns all tokens back to the Core-1 SubDAO.
func MoveVestingCoinFromVestingAccount(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, accAddr sdk.AccAddress, core1AccAddr sdk.AccAddress) error {
	now := ctx.BlockHeader().Time

	stdAcc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	vacc, ok := stdAcc.(*authvestingtypes.PeriodicVestingAccount)
	if !ok {
		// For e2e testing
		fmt.Printf("account " + accAddr.String() + " is not a vesting account. This should not run on mainnet.\n")
		return nil
	}

	// Shows locked funds
	showLockedCoins(vacc, now)

	// Gets all coins which have not been unlocked yet. (Will be minted to the Core-1 SubDAO later for usage.)
	unvestedCoins := getUnVestedCoins(vacc, bondDenom)

	// Clears the account so all unvested coins are unlocked & removes any future vesting periods.
	// (This way we can unbond and transfer all coins)
	clearVestingAccount(ctx, vacc, keepers)

	// Finish re-deleations.
	if err := completeAllRedelegations(ctx, keepers, accAddr, now); err != nil {
		return err
	}

	// Instant unbond all delegations
	if err := unbondAllAndFinish(ctx, now, keepers, accAddr); err != nil {
		return err
	}

	// Moves the accounts held balance to the SubDAO.
	_, err := migrateBalanceToCore1SubDao(ctx, vacc, keepers, core1AccAddr, bondDenom)
	if err != nil {
		return err
	}

	// Mints unvested tokens to the Core-1 SubDAO for future use.
	if err := transferUnvestedTokensToCore1SubDao(ctx, keepers, bondDenom, core1AccAddr, unvestedCoins); err != nil {
		return err
	}

	// TODO: Delete the account, not further actions needed. Any downside to this?
	keepers.AccountKeeper.RemoveAccount(ctx, vacc)

	// return fmt.Errorf("DEBUGGING: not implemented MoveVestingCoinFromVestAccount")
	return nil
}

func transferUnvestedTokensToCore1SubDao(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, core1AccAddr sdk.AccAddress, unvestedCoins sdk.Coins) error {
	// mint unvested coins to be transferred.
	if err := keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, unvestedCoins); err != nil {
		return err
	}

	// transfer unvested coins back to the to core1 subdao
	if err := keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, core1AccAddr, unvestedCoins); err != nil {
		return err
	}

	core1BalC := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)
	fmt.Printf("Updated Core1 SubDAO Balance: %v\n", core1BalC)

	return nil
}

func migrateBalanceToCore1SubDao(ctx sdk.Context, vacc *authvestingtypes.PeriodicVestingAccount, keepers *keepers.AppKeepers, core1AccAddr sdk.AccAddress, bondDenom string) (sdk.Coin, error) {
	accbal := keepers.BankKeeper.GetBalance(ctx, vacc.GetAddress(), bondDenom)

	if e := keepers.BankKeeper.SendCoins(ctx, vacc.GetAddress(), core1AccAddr, sdk.NewCoins(accbal)); e != nil {
		return sdk.Coin{}, fmt.Errorf("error sending coins: %v", e)
	}

	core1BalC := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)

	fmt.Printf("moved %v from %v to %v\n", accbal, vacc.GetAddress(), core1AccAddr)
	fmt.Printf("New Core1 Bal: %v\n", core1BalC)

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

func getUnVestedCoins(vacc *authvestingtypes.PeriodicVestingAccount, bondDenom string) sdk.Coins {
	mintAmt := sdk.ZeroInt()

	for i := range vacc.VestingPeriods {
		ujunoAmt := vacc.VestingPeriods[i].Amount.AmountOf(bondDenom)

		mintAmt = mintAmt.Add(ujunoAmt)
	}

	return sdk.NewCoins(sdk.NewCoin(bondDenom, mintAmt))
}

func clearVestingAccount(ctx sdk.Context, vacc *authvestingtypes.PeriodicVestingAccount, keepers *keepers.AppKeepers) {
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
