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

func MoveDelegatorDelegationsToCommunityPool(ctx sdk.Context, delAcc sdk.AccAddress, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper, distr *distrkeeper.Keeper) {
	bondDenom := staking.BondDenom(ctx)
	delegatorDelegations := staking.GetAllDelegatorDelegations(ctx, delAcc)

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

		if validator.IsBonded() {
			amountToBeMovedFromBondedPool = amountToBeMovedFromBondedPool.Add(unbondedAmount)
		} else {
			amountToBeMovedFromNotBondedPool = amountToBeMovedFromNotBondedPool.Add(unbondedAmount)
		}
	}

	delegatorUnbondingDelegations := staking.GetAllUnbondingDelegations(ctx, delAcc)
	for _, unbondingDelegation := range delegatorUnbondingDelegations {
		for _, entry := range unbondingDelegation.Entries {
			amountToBeMovedFromNotBondedPool = amountToBeMovedFromNotBondedPool.Add(entry.Balance)
		}
		staking.RemoveUnbondingDelegation(ctx, unbondingDelegation)
	}

	coinsToBeMovedFromNotBondedPool := sdk.NewCoins(sdk.NewCoin(bondDenom, amountToBeMovedFromNotBondedPool))
	coinsToBeMovedFromBondedPool := sdk.NewCoins(sdk.NewCoin(bondDenom, amountToBeMovedFromBondedPool))

	if !coinsToBeMovedFromNotBondedPool.Empty() {
		notBondedPoolAcc := staking.GetNotBondedPool(ctx)
		distr.FundCommunityPool(ctx, coinsToBeMovedFromNotBondedPool, notBondedPoolAcc.GetAddress())
	}
	if !coinsToBeMovedFromBondedPool.Empty() {
		bondedPoolAcc := staking.GetBondedPool(ctx)
		distr.FundCommunityPool(ctx, coinsToBeMovedFromBondedPool, bondedPoolAcc.GetAddress())
	}
}

//CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper, distr *distrkeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		for _, addrString := range addressesToBeAdjusted {
			accAddr, _ := sdk.AccAddressFromBech32(addrString)
			// unbond the accAddr delegations, send all the unbonding and unbonded tokens to the community pool
			MoveDelegatorDelegationsToCommunityPool(ctx, accAddr, staking, bank, distr)
			// send 50k juno from the community pool to the accAddr if the master account has less than 50k juno
			accAddrAmount := bank.GetBalance(ctx, accAddr, staking.BondDenom(ctx)).Amount
			if sdk.NewIntFromUint64(50000000000).GT(accAddrAmount) {
				bank.SendCoinsFromModuleToAccount(ctx, distrtypes.ModuleName, accAddr, sdk.NewCoins(sdk.NewCoin(staking.BondDenom(ctx), sdk.NewIntFromUint64(50000000000).Sub(accAddrAmount))))
			}
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
