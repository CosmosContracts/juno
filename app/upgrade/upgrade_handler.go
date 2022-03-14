package lupercalia

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var addressesToBeAdjusted = []string{
	"juno1aeh8gqu9wr4u8ev6edlgfq03rcy6v5twfn0ja8",
}

func MoveAccountCoinToCommunityPool(ctx sdk.Context, accAddr sdk.AccAddress, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper, distr *distrkeeper.Keeper) {
	bondDenom := staking.BondDenom(ctx)

	// this loop will turn all delegator's delegations into unbonding delegations
	for _, delegation := range staking.GetAllDelegatorDelegations(ctx, accAddr) {
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := staking.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}
		staking.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares()) //nolint:errcheck // nolint because otherwise we'd have a time and nothing to do with it.
	}

	now := ctx.BlockHeader().Time

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
		staking.CompleteUnbonding(ctx, accAddr, validatorValAddr)
	}

	// account balance after finishing unbonding
	accCoin := bank.GetBalance(ctx, accAddr, bondDenom)
	// move all coin from that acc to community pool
	distr.FundCommunityPool(ctx, sdk.NewCoins(accCoin), accAddr)
}

//CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper, distr *distrkeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		for _, addrString := range addressesToBeAdjusted {
			accAddr, _ := sdk.AccAddressFromBech32(addrString)
			// move all juno from acc to community pool (uncluding bonded juno)
			MoveAccountCoinToCommunityPool(ctx, accAddr, staking, bank, distr)
			// send 50k juno from the community pool to the accAddr
			bank.SendCoinsFromModuleToAccount(ctx, distrtypes.ModuleName, accAddr, sdk.NewCoins(sdk.NewCoin(staking.BondDenom(ctx), sdk.NewIntFromUint64(50000000000))))
			feePool := distr.GetFeePool(ctx)
			coin := sdk.NewCoin(staking.BondDenom(ctx), sdk.NewIntFromUint64(50000000000))
			feePool.CommunityPool = feePool.CommunityPool.Sub(sdk.NewDecCoinsFromCoins(coin))
			distr.SetFeePool(ctx, feePool)
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
