package whale

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func getWhaleDelagtion(ctx sdk.Context, staking *stakingkeeper.Keeper) []*stakingtypes.Delegation {
	//whale address
	whaleAddress, _ := sdk.AccAddressFromBech32("juno1aeh8gqu9wr4u8ev6edlgfq03rcy6v5twfn0ja8")

	// validators that whale delagates to
	whaleValidators := staking.GetDelegatorValidators(ctx, whaleAddress, 120)

	whaleDelegations := []*stakingtypes.Delegation{}
	for _, v := range whaleValidators {
		valAdress, _ := sdk.ValAddressFromBech32(v.OperatorAddress)

		del, _ := staking.GetDelegation(ctx, whaleAddress, valAdress)

		whaleDelegations = append(whaleDelegations, &del)
	}
	return whaleDelegations
}

func whaleToBathroom(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	// get all whale delegations
	whaleDelegations := getWhaleDelagtion(ctx, staking)

	whaleAddress, _ := sdk.AccAddressFromBech32("juno1aeh8gqu9wr4u8ev6edlgfq03rcy6v5twfn0ja8")

	// the address of 1 validator that the whale delegate to
	whaleValidator := whaleDelegations[0].GetValidatorAddr()

	for _, delegation := range whaleDelegations {
		//undelegate
		staking.Undelegate(ctx, whaleAddress, delegation.GetValidatorAddr(), delegation.GetShares())
	}
	//set Unboding to verylow
	completionTime := ctx.BlockHeader().Time.Add(staking.UnbondingTime(ctx))

	ubd := stakingtypes.NewUnbondingDelegation(whaleAddress, whaleValidator, ctx.BlockHeader().Height, completionTime, sdk.NewInt(1))
	staking.SetUnbondingDelegation(ctx, ubd)
}

//CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	wasmKeeper *wasm.Keeper, staking *stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		whaleToBathroom(ctx, staking)
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
