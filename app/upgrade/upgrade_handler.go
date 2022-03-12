package lupercalia

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var addressesToBeAdjusted = []string{
	"juno1aeh8gqu9wr4u8ev6edlgfq03rcy6v5twfn0ja8",
}

func MoveDelegatorDelegationsToCommunityPool(ctx sdk.Context, delAcc sdk.AccAddress, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) {
	bondDenom := staking.BondDenom(ctx)

	fmt.Printf("denom = %s \n", bondDenom)

	delegatorDelegations := staking.GetAllDelegatorDelegations(ctx, delAcc)

	fmt.Printf("delegatorDelegations = %v \n", delegatorDelegations)

	amountToBeMovedFromNotBondedPool := sdk.ZeroInt()
	amountToBeMovedFromBondedPool := sdk.ZeroInt()

	for _, delegation := range delegatorDelegations {

		validatorValAddr := delegation.GetValidatorAddr()
		validator, found := staking.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}

		unbondedAmount, err := staking.Unbond(ctx, delAcc, validatorValAddr, delegation.GetShares()) //nolint:errcheck // nolint because otherwise we'd have a time and nothing to do with it.
		if err != nil {
			panic(err)
		}

		fmt.Printf("unbondedAmount = %d \n", unbondedAmount.Uint64())

		if validator.IsBonded() {
			amountToBeMovedFromBondedPool = amountToBeMovedFromBondedPool.Add(unbondedAmount)
		} else {
			amountToBeMovedFromNotBondedPool = amountToBeMovedFromNotBondedPool.Add(unbondedAmount)
		}
	}

	delegatorUnbondingDelegations := staking.GetAllUnbondingDelegations(ctx, delAcc)
	for _, unbondingDelegation := range delegatorUnbondingDelegations {
		for _, entry := range unbondingDelegation.Entries {
			fmt.Printf("entry.Balance = %d \n", entry.Balance.Uint64())
			amountToBeMovedFromNotBondedPool = amountToBeMovedFromNotBondedPool.Add(entry.Balance)
		}
		staking.RemoveUnbondingDelegation(ctx, unbondingDelegation)
	}

	coinsToBeMovedFromNotBondedPool := sdk.NewCoins(sdk.NewCoin(bondDenom, amountToBeMovedFromNotBondedPool))
	fmt.Printf("coinsToBeMovedFromNotBondedPool = %d \n", coinsToBeMovedFromNotBondedPool.AmountOf(bondDenom).Uint64())

	coinsToBeMovedFromBondedPool := sdk.NewCoins(sdk.NewCoin(bondDenom, amountToBeMovedFromBondedPool))
	fmt.Printf("coinsToBeMovedFromBondedPool = %d \n", coinsToBeMovedFromBondedPool.AmountOf(bondDenom).Uint64())

	bank.SendCoinsFromModuleToModule(ctx, stakingtypes.NotBondedPoolName, distrtypes.ModuleName, coinsToBeMovedFromNotBondedPool)
	bank.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, distrtypes.ModuleName, coinsToBeMovedFromBondedPool)
}

//CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	wasmKeeper *wasm.Keeper, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		for _, addrString := range addressesToBeAdjusted {
			accAddr, _ := sdk.AccAddressFromBech32(addrString)
			// unbond the accAddr delegations, send all the unbonding and unbonded tokens to the community pool
			MoveDelegatorDelegationsToCommunityPool(ctx, accAddr, staking, bank)
			// send 50k juno from the community pool to the accAddr
			bank.SendCoinsFromModuleToAccount(ctx, distrtypes.ModuleName, accAddr, sdk.NewCoins(sdk.NewCoin(staking.BondDenom(ctx), sdk.NewIntFromUint64(50000000000))))
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
