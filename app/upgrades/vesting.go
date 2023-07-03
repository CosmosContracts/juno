package upgrades

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

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
		fmt.Printf("account " + accAddr.String() + " is not a vesting account.\n")
		return nil
	}

	fmt.Printf("== Vesting Account Address: %s ==\n", vacc.GetAddress().String())

	// Gets non-vested coins (These get returned back to Core-1 SubDAO)
	// The SubDAO should increase exactly with this much.
	unvestedCoins := getStillVestingCoins(vacc, now)
	fmt.Printf("Locked / waiting to vest Coins: %v\n", unvestedCoins)

	// Get Core1 and before migration
	core1BeforeBal := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)

	// Clears the account so all all future vesting periods are removed.
	// Sets it as a standard base account.
	clearVestingAccount(ctx, vacc, keepers)

	// Complete any re-deleations to become standard delegations.
	if err := completeAllRedelegations(ctx, keepers, accAddr, now); err != nil {
		return err
	}

	// Instant unbond all delegations. Returns the amount of tokens (non rewards) which were returned.
	_, err := unbondAllAndFinish(ctx, now, keepers, accAddr)
	if err != nil {
		return err
	}

	// Moves unvested tokens to the Core-1 SubDAO for future use.
	if err := transferUnvestedTokensToCore1SubDao(ctx, keepers, bondDenom, core1AccAddr, unvestedCoins); err != nil {
		return err
	}

	// Ensure the post validation checks are met.
	if err := postValidation(ctx, keepers, bondDenom, accAddr, core1AccAddr, unvestedCoins, core1BeforeBal); err != nil {
		return err
	}

	// return fmt.Errorf("DEBUGGING: not implemented MoveVestingCoinFromVestAccount")
	return nil
}

func postValidation(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, accAddr sdk.AccAddress, core1AccAddr sdk.AccAddress, unvestedCoins sdk.Coins, core1BeforeBal sdk.Coin) error {
	// Core1 balance should only increase by exactly the core1Bal + unvestedCoins
	core1BalAfter := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)
	if !core1BeforeBal.Add(unvestedCoins[0]).IsEqual(core1BalAfter) {
		return fmt.Errorf("ERROR: core1BeforeBal (%v) + unvestedCoins (%v) != core1BalAfter (%v)", core1BeforeBal, unvestedCoins, core1BalAfter)
	}

	// vesting account should have no future vesting periods
	newVacc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	if _, ok := newVacc.(*authvestingtypes.PeriodicVestingAccount); ok {
		return fmt.Errorf("ERROR: account %s still is a vesting account", accAddr.String())
	}

	// ensure the account has 0 delegations, redelegations, or unbonding delegations
	delegations := keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr)
	if len(delegations) != 0 {
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

func transferUnvestedTokensToCore1SubDao(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, core1AccAddr sdk.AccAddress, unvestedCoins sdk.Coins) error {
	// mint unvested coins to be transferred.
	fmt.Printf("Minting Unvested Coins back to Core-1: %v\n", unvestedCoins)
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

// Returns the amount of tokens which were unbonded (not rewards)
func unbondAllAndFinish(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) (math.Int, error) {
	unbondedAmt := math.ZeroInt()

	// Unbond all delegations from the account
	for _, delegation := range keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr) {
		// fmt.Printf("delegation: %v\n", delegation)
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}

		_, err := keepers.StakingKeeper.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	// Take all unbonding and complete them.
	for _, unbondingDelegation := range keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr) {
		// fmt.Printf("unbondingDelegation: %v\n", unbondingDelegation)
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

func getStillVestingCoins(vacc *authvestingtypes.PeriodicVestingAccount, now time.Time) sdk.Coins {
	return vacc.GetVestingCoins(now)
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
	vacc.BaseVestingAccount.DelegatedFree = sdk.Coins{}

	keepers.AccountKeeper.SetAccount(ctx, vacc)
	keepers.AccountKeeper.SetAccount(ctx, vacc.BaseAccount)
	fmt.Println("Vesting Account set to BaseAccount, not more vesting periods.")
}
