package upgrades

import (
	"fmt"
	"time"

	"github.com/CosmosContracts/juno/v16/app/keepers"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

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
	locakedVesting := vacc.LockedCoinsFromVesting(vacc.GetVestingCoins(now))
	fmt.Printf("locked: %v\n", locked)
	fmt.Printf("locakedVesting: %v\n", locakedVesting)
	return locked
}

// Undelegates all tokens, and send it back to the core1 subdao address for the vesting contract to instantiate
func MoveVestingCoinFromVestingAccount(ctx sdk.Context, accAddr sdk.AccAddress, keepers *keepers.AppKeepers, core1SubDaoAddress string, bondDenom string) (sdk.Coin, error) {
	var err error

	now := ctx.BlockHeader().Time

	stdAcc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	vacc, ok := stdAcc.(*authvestingtypes.ContinuousVestingAccount)
	if !ok {
		panic("not a ContinuousVestingAccount vesting account")
	}

	// should show numbers
	l := checkLockedCoins(vacc, now)
	fmt.Printf("l: %v\n", l)

	// set end time to now to unlock.
	vacc.EndTime = now.Unix() - 1_000_000
	vacc.BaseVestingAccount.EndTime = now.Unix() - 1_000_000
	// vacc.DelegatedFree = sdk.NewCoins()
	// vacc.DelegatedVesting = sdk.NewCoins()
	// vacc.OriginalVesting = sdk.NewCoins()
	// vacc.BaseVestingAccount.DelegatedFree = sdk.NewCoins()
	// vacc.BaseVestingAccount.DelegatedVesting = sdk.NewCoins()
	// vacc.BaseVestingAccount.OriginalVesting = sdk.NewCoins()

	// Nothing shown here.
	checkLockedCoins(vacc, now)

	// undelegate all.
	if err := undelegate(ctx, now, keepers, accAddr, vacc); err != nil {
		return sdk.Coin{}, err
	}

	bal := keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	fmt.Printf("Balance: %s\n", bal)
	if err := keepers.BankKeeper.SendCoinsFromAccountToModule(ctx, accAddr, distrtypes.ModuleName, sdk.NewCoins(bal)); err != nil {
		return sdk.Coin{}, err
	}

	// get distrtypes.ModuleName balance
	distrtypesModuleBalance := keepers.BankKeeper.GetBalance(ctx, sdk.AccAddress(distrtypes.ModuleName), bondDenom)
	fmt.Printf("distrtypesModuleBalance: %v\n", distrtypesModuleBalance)

	// update vacc to be entirely spendable
	// vacc.DelegatedFree = sdk.NewCoins(bal)
	// vacc.BaseVestingAccount.DelegatedFree = sdk.NewCoins(bal)

	fmt.Printf("MoveVestingCoinFromVestingAccount: %s\n", vacc)

	// send funds since they are unlocked.
	// spendable balance here is nil.
	fmt.Print("sending\n", keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom))
	if err := keepers.BankKeeper.SendCoins(ctx, accAddr, sdk.AccAddress(core1SubDaoAddress), sdk.NewCoins(bal)); err != nil {
		return sdk.Coin{}, err
	}

	fmt.Printf("MoveVestingCoinFromVestingAccount: %s\n", vacc)

	core1Bal := keepers.BankKeeper.GetBalance(ctx, sdk.AccAddress(core1SubDaoAddress), bondDenom)
	fmt.Printf("core1Bal: %v\n", core1Bal)

	return sdk.Coin{}, fmt.Errorf("not implemented")

	// undelegates
	undelegateAmount := vacc.DelegatedVesting.AmountOf(bondDenom)
	fmt.Printf("DelegatedVesting before: %v\n", undelegateAmount)
	vacc.TrackUndelegation(vacc.DelegatedVesting)
	fmt.Printf("DelegatedVesting after: %v\n", vacc.DelegatedVesting)

	vacc.BaseVestingAccount.EndTime = now.Unix() - 1

	stdbal := keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	fmt.Printf("stdbal before: %v\n", stdbal)

	if err := keepers.BankKeeper.SendCoinsFromAccountToModule(ctx, accAddr, distrtypes.ModuleName, sdk.NewCoins(stdbal)); err != nil {
		panic(err)
	}

	// then transfer that to core 1 fgrom ModuleToAccount
	if err := keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, distrtypes.ModuleName, sdk.AccAddress(core1SubDaoAddress), sdk.NewCoins(stdbal)); err != nil {
		panic(err)
	}

	// TODO: vesting

	// get current held balance from stdAcc
	stdbal = keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)
	fmt.Printf("stdbal after: %v\n", stdbal)

	// get balance of core1 (needs to be the original vesting ammount)
	core1Bal = keepers.BankKeeper.GetBalance(ctx, sdk.AccAddress(core1SubDaoAddress), bondDenom)
	fmt.Printf("core1Bal: %v\n", core1Bal)

	// print out vacc
	fmt.Printf("MoveVestingCoinFromVestingAccount: %s\n", vacc)

	return core1Bal, nil
	// return sdk.Coin{}, fmt.Errorf("not implemented")

	// // get balance of accAddr
	// bal := (&keepers.BankKeeper).GetBalance(ctx, accAddr, bondDenom)
	// fmt.Printf("MoveVestingCoinFromVestingAccount: %s\n", bal)
	// // account info
	// accInfo := (&keepers.AccountKeeper).GetAccount(ctx, accAddr)
	// fmt.Printf("MoveVestingCoinFromVestingAccount: %s\n", accInfo)
	// for _, delegation := range keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr) {
	// 	validatorValAddr := delegation.GetValidatorAddr()
	// 	_, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr)
	// 	if !found {
	// 		continue
	// 	}
	// 	fmt.Printf("MoveVestingCoinFromVestingAccount: %s\n", delegation)
	// 	// _, err := keepers.StakingKeeper.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
	// 	// if err != nil {
	// 	// 	panic(err)
	// 	// }
	// }
	// return sdk.Coin{}, fmt.Errorf("not implemented")

	// TODO: Is this needed?
	// Complete any active re-delegations
	// for _, activeRedelegation := range keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535) {
	// 	redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
	// 	redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

	// 	// set all entry completionTime to now so we can complete redelegation
	// 	for i := range activeRedelegation.Entries {
	// 		activeRedelegation.Entries[i].CompletionTime = now
	// 	}

	// 	keepers.StakingKeeper.SetRedelegation(ctx, activeRedelegation)
	// 	_, err := keepers.StakingKeeper.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	// Complete all delegator's unbonding delegations
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
			panic(err)
		}
	}

	// account balance after finishing unbonding
	accCoin := keepers.BankKeeper.GetBalance(ctx, accAddr, bondDenom)

	vacc.OriginalVesting = vacc.OriginalVesting.Add(accCoin)

	// make entire balance spendable
	// keepers.AccountKeeper.SetAccount(ctx, )

	// Send coin to Core-1 subDao for holding on behalf of the core-1 members
	destAcc, _ := sdk.AccAddressFromBech32(core1SubDaoAddress)
	err = keepers.BankKeeper.SendCoins(ctx, accAddr, destAcc, sdk.NewCoins(accCoin))
	if err != nil {
		panic(err)
	}

	ctx.Logger().Info("MoveVestingCoinFromVestingAccount", "account", accAddr.String(), "amount", accCoin)

	// return accCoin, nil
	return accCoin, fmt.Errorf("not implemented")
}
