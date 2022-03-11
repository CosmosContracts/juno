package lupercalia

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func getDelagtion(ctx sdk.Context, staking *stakingkeeper.Keeper) []*stakingtypes.Delegation {
	// address
	acctAddress, _ := sdk.AccAddressFromBech32("juno1aeh8gqu9wr4u8ev6edlgfq03rcy6v5twfn0ja8")

	// validators that whale delagates to
	acctValidators := staking.GetDelegatorValidators(ctx, acctAddress, 120)

	acctDelegations := []*stakingtypes.Delegation{}
	for _, v := range acctValidators {
		valAdress, _ := sdk.ValAddressFromBech32(v.OperatorAddress)

		del, _ := staking.GetDelegation(ctx, acctAddress, valAdress)

		acctDelegations = append(acctDelegations, &del)
	}
	return acctDelegations
}

func adjustDelegations(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	// get all whale delegations
	acctDelegations := getDelagtion(ctx, staking)

	acctAddress, _ := sdk.AccAddressFromBech32("juno1aeh8gqu9wr4u8ev6edlgfq03rcy6v5twfn0ja8")

	completionTime := ctx.BlockHeader().Time.Add(staking.UnbondingTime(ctx))

	for _, delegation := range acctDelegations {
		validator := delegation.GetValidatorAddr()
		// undelegate
		_, err := staking.Undelegate(ctx, acctAddress, delegation.GetValidatorAddr(), delegation.GetShares()) //nolint:errcheck // nolint because otherwise we'd have a time and nothing to do with it.
		if err != nil {
			panic(err)
		}

		ubd := stakingtypes.NewUnbondingDelegation(acctAddress, validator, ctx.BlockHeader().Height, completionTime, sdk.NewInt(1))
		staking.SetUnbondingDelegation(ctx, ubd)
	}
}

//CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	wasmKeeper *wasm.Keeper, staking *stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		adjustDelegations(ctx, staking)

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
