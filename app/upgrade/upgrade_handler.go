package unity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// Address of the account which will have all juno sent to the unity proposal
var addressesToBeAdjusted = []string{
	"juno1ep4unl6pdet64ph43sqhxd9hvftdmealqmzrze",
}

// UnityContractByteAddress is the bytes of the public key for the address of the Unity contract
// $ junod keys parse juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8
// human: juno
// bytes: ADE4A5F5803A439835C636395A8D648DEE57B2FC90D98DC17FA887159B69638B
const UnityContractByteAddress = "ADE4A5F5803A439835C636395A8D648DEE57B2FC90D98DC17FA887159B69638B"

// ClawbackCoinFromAccount undelegate all amount in adjusted address (bypass 28 day unbonding), and send it to dead address
func ClawbackCoinFromAccount(ctx sdk.Context, accAddr sdk.AccAddress, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) {
	bondDenom := staking.BondDenom(ctx)

	now := ctx.BlockHeader().Time

	// this loop will complete all delegator's active redelegations
	for _, activeRedelegation:= range staking.GetRedelegations(ctx, accAddr, 65535) {
		// src/dest validator addresses of this redelegation
		redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
		redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

		// set all entry completionTime to now so we can complete redelegation
		for i := range activeRedelegation.Entries {
			activeRedelegation.Entries[i].CompletionTime = now
		}

		staking.SetRedelegation(ctx, activeRedelegation)
		_, err := staking.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
		if err != nil {
			panic(err)
		}
	}

	// this loop will turn all delegator's delegations into unbonding delegations
	for _, delegation := range staking.GetAllDelegatorDelegations(ctx, accAddr) {
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := staking.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}
		_, err := staking.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares()) //nolint:errcheck // nolint because otherwise we'd have a time and nothing to do with it.
		if err != nil {
			panic(err)
		}
	}


	// this loop will complete all delegator's unbonding delegations
	for _, unbondingDelegation := range staking.GetAllUnbondingDelegations(ctx, accAddr) {
		// validator address of this unbonding delegation
		validatorStringAddr := unbondingDelegation.ValidatorAddress
		validatorValAddr, _ := sdk.ValAddressFromBech32(validatorStringAddr)

		// set all entry completionTime to now so we can complete unbonding delegation
		for i := range unbondingDelegation.Entries {
			unbondingDelegation.Entries[i].CompletionTime = now
		}
		staking.SetUnbondingDelegation(ctx, unbondingDelegation)
		_, err := staking.CompleteUnbonding(ctx, accAddr, validatorValAddr)
		if err != nil {
			panic(err)
		}
	}

	// account balance after finishing unbonding
	accCoin := bank.GetBalance(ctx, accAddr, bondDenom)

	//get Unity Contract Address and send coin to this address
	destAcc, _ := sdk.AccAddressFromHex(UnityContractByteAddress)
	err := bank.SendCoins(ctx, accAddr, destAcc, sdk.NewCoins(accCoin))
	if err != nil {
		panic(err)
	}
}

//CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		for _, addrString := range addressesToBeAdjusted {
			accAddr, _ := sdk.AccAddressFromBech32(addrString)
			ClawbackCoinFromAccount(ctx, accAddr, staking, bank)
		}
		// force an update of validator min commission
		// we already did this for moneta
		// but validators could have snuck in changes in the
		// interim
		// and via state sync to post-moneta
		validators := staking.GetAllValidators(ctx)
		// hard code this because we don't want
		// a) a fork or
		// b) immediate reaction with additional gov props
		minCommissionRate := sdk.NewDecWithPrec(5, 2)
		for _, v := range validators {
			if v.Commission.Rate.LT(minCommissionRate) {
				if v.Commission.MaxRate.LT(minCommissionRate) {
					v.Commission.MaxRate = minCommissionRate
				}

				v.Commission.Rate = minCommissionRate
				v.Commission.UpdateTime = ctx.BlockHeader().Time

				// call the before-modification hook since we're about to update the commission
				staking.BeforeValidatorModified(ctx, v.GetOperator())

				staking.SetValidator(ctx, v)
			}
		}

		// Set wasm old version to 1 if we want to call wasm's InitGenesis ourselves
		// in this upgrade logic ourselves
		// vm[wasm.ModuleName] = wasm.ConsensusVersion

		// otherwise we run this, which will run wasm.InitGenesis(wasm.DefaultGenesis())
		// and then override it after
		newVM, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVM, err
		}

		// override here
		return newVM, err
	}

}
