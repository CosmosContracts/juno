package v23

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/CosmosContracts/juno/v26/app/keepers"
)

// Stops a vesting account and returns all tokens back to the community pool
func MoveVestingCoinFromVestingAccount(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, owner string, accAddr sdk.AccAddress) error {
	now := ctx.BlockHeader().Time

	stdAcc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	vacc, ok := stdAcc.(*authvestingtypes.PeriodicVestingAccount)
	if !ok {
		// For e2e testing
		fmt.Printf("account " + accAddr.String() + " is not a vesting account.\n")
		return nil
	}

	fmt.Printf("\n\n== Vesting Account Address: %s (%s) ==\n", vacc.GetAddress().String(), owner)

	// Gets vesting coins (These get returned back to community pool)
	// we should filter vesting coins to only include the bondDenom
	vestingCoins := vacc.GetVestingCoins(now)
	fmt.Printf("All Vesting Coins: %v\n", vestingCoins)
	vestingJuno := vestingCoins.AmountOf(bondDenom)
	fmt.Printf("Vesting Junos: %v\n", vestingJuno)

	// Display locked & spendable funds
	lockedCoins := keepers.BankKeeper.LockedCoins(ctx, accAddr)
	fmt.Printf("Locked Coins: %v\n", lockedCoins)
	spendableCoins := keepers.BankKeeper.SpendableCoins(ctx, accAddr)
	fmt.Printf("Spendable Coins: %v\n", spendableCoins)

	// Instantly complete any re-deleations.
	amt, err := completeAllRedelegations(ctx, now, keepers, accAddr)
	if err != nil {
		return err
	}
	fmt.Println("Redelegated Amount: ", amt)

	// Instantly unbond all delegations.
	amt, err = unbondAllAndFinish(ctx, now, keepers, accAddr)
	if err != nil {
		return err
	}
	fmt.Println("Unbonded Amount: ", amt)

	// Community pool balance before transfer
	cpoolBeforeBal := keepers.DistrKeeper.GetFeePool(ctx).CommunityPool

	// Set the vesting account to a base account
	keepers.AccountKeeper.SetAccount(ctx, vacc.BaseAccount)

	// Moves vesting tokens to the council.
	if err := transferUnvestedTokensToCommunityPool(ctx, keepers, accAddr, sdk.NewCoin(bondDenom, vestingJuno)); err != nil {
		return err
	}

	// Log new council balance
	cpoolAfterBal := keepers.DistrKeeper.GetFeePool(ctx).CommunityPool

	fmt.Printf("Community Pool Balance Before: %v\n", cpoolBeforeBal)
	fmt.Printf("Community Pool Balance After: %v\n", cpoolAfterBal)

	// Ensure the post validation checks are met.
	err = postValidation(ctx, keepers, bondDenom, accAddr, vestingCoins, cpoolBeforeBal)
	return err
}

func postValidation(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, accAddr sdk.AccAddress, vestingCoins sdk.Coins, cpoolBeforeBal sdk.DecCoins) error {
	// Community pool juno balance should only increase by exactly the vestedCoins
	cpoolAfterBal := keepers.DistrKeeper.GetFeePool(ctx).CommunityPool

	// only count vesting junos
	vestingJuno := vestingCoins.AmountOf(bondDenom)

	if !cpoolBeforeBal.AmountOf(bondDenom).Add(vestingJuno.ToLegacyDec()).Equal(cpoolAfterBal.AmountOf(bondDenom)) {
		return fmt.Errorf("ERROR: community pool balance before (%v) + unvested juno (%v) from unvestedCoins (%v) != core1BalAfter (%v)", cpoolBeforeBal, vestingJuno, vestingCoins, cpoolAfterBal)
	}

	// vesting account should have no future vesting periods
	newVacc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	if _, ok := newVacc.(*authvestingtypes.PeriodicVestingAccount); ok {
		return fmt.Errorf("ERROR: account %s still is a vesting account", accAddr.String())
	}

	// ensure the account has 0 delegations, redelegations, or unbonding delegations,
	delegations := keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr)
	if !(len(delegations) == 0) {
		return fmt.Errorf("ERROR: account %s still has delegations", accAddr.String())
	}

	redelegations := keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535)
	if len(redelegations) != 0 {
		return fmt.Errorf("ERROR: account %s still has redelegations", accAddr.String())
	}

	unbondingDelegations := keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr)
	if len(unbondingDelegations) != 0 {
		return fmt.Errorf("ERROR: account %s still has unbonding delegations", accAddr.String())
	}

	return nil
}

// Transfer funds from the vesting account to the Council SubDAO.
func transferUnvestedTokensToCommunityPool(ctx sdk.Context, keepers *keepers.AppKeepers, accAddr sdk.AccAddress, vestingJuno sdk.Coin) error {
	fmt.Printf("Sending Vesting Juno to Community pool: %v\n", vestingJuno)
	err := keepers.DistrKeeper.FundCommunityPool(ctx, sdk.NewCoins(vestingJuno), accAddr)
	return err
}

// Completes all re-delegations and returns the amount of tokens which were re-delegated.
func completeAllRedelegations(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) (math.Int, error) {
	redelegatedAmt := math.ZeroInt()

	for _, activeRedelegation := range keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535) {
		redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
		redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

		// set all entry completionTime to now so we can complete re-delegation
		for i := range activeRedelegation.Entries {
			activeRedelegation.Entries[i].CompletionTime = now
			redelegatedAmt = redelegatedAmt.Add(math.Int(activeRedelegation.Entries[i].SharesDst))
		}

		keepers.StakingKeeper.SetRedelegation(ctx, activeRedelegation)
		_, err := keepers.StakingKeeper.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
		if err != nil {
			return redelegatedAmt, err
		}
	}

	return redelegatedAmt, nil
}

// Returns the amount of tokens which were unbonded (not rewards)
func unbondAllAndFinish(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) (math.Int, error) {
	unbondedAmt := math.ZeroInt()

	// Unbond all delegations from the account
	for _, delegation := range keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr) {
		validatorValAddr := delegation.GetValidatorAddr()
		if _, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr); !found {
			continue
		}

		_, err := keepers.StakingKeeper.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	// Take all unbonding and complete them.
	for _, unbondingDelegation := range keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr) {
		validatorStringAddr := unbondingDelegation.ValidatorAddress
		validatorValAddr, _ := sdk.ValAddressFromBech32(validatorStringAddr)

		// Complete unbonding delegation
		for i := range unbondingDelegation.Entries {
			unbondingDelegation.Entries[i].CompletionTime = now
			unbondedAmt = unbondedAmt.Add(unbondingDelegation.Entries[i].Balance)
		}

		keepers.StakingKeeper.SetUnbondingDelegation(ctx, unbondingDelegation)
		_, err := keepers.StakingKeeper.CompleteUnbonding(ctx, accAddr, validatorValAddr)
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	return unbondedAmt, nil
}
