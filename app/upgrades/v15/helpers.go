package v15

import (
	"github.com/CosmosContracts/juno/v15/app/keepers"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Undelegates all tokens, and send it back to the core1 subdao address for the vesting contract to instantiate
func MoveVestingCoinFromVestingAccount(ctx sdk.Context, accAddr sdk.AccAddress, keepers *keepers.AppKeepers, core1SubDaoAddress string) sdk.Coin {
	bondDenom := keepers.StakingKeeper.BondDenom(ctx)

	now := ctx.BlockHeader().Time

	// Complete any active re-delegations
	for _, activeRedelegation := range keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535) {
		redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
		redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

		// set all entry completionTime to now so we can complete redelegation
		for i := range activeRedelegation.Entries {
			activeRedelegation.Entries[i].CompletionTime = now
		}

		keepers.StakingKeeper.SetRedelegation(ctx, activeRedelegation)
		_, err := keepers.StakingKeeper.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
		if err != nil {
			panic(err)
		}
	}

	// Unbond all delegations from the account
	for _, delegation := range keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr) {
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}
		_, err := keepers.StakingKeeper.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
		if err != nil {
			panic(err)
		}
	}

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

	// Send coin to Core-1 subDao for holding on behalf of the core-1 members (interim step)
	destAcc, _ := sdk.AccAddressFromBech32(core1SubDaoAddress)
	err := keepers.BankKeeper.SendCoins(ctx, accAddr, destAcc, sdk.NewCoins(accCoin))
	if err != nil {
		panic(err)
	}

	ctx.Logger().Info("MoveVestingCoinFromVestingAccount", "account", accAddr.String(), "amount", accCoin)
	return accCoin
}
