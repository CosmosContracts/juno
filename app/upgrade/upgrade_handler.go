package whale

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// Fixes an error where validators can be created with a commission rate
// less than the network minimum rate.
var mockData []struct {
	valAddress   sdk.ValAddress
	sharesAmount sdk.Dec
}

var newInfo struct {
	whaleAddress sdk.AccAddress
	valAddress   sdk.ValAddress
}

func getWhaleData(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	//get whale address
	whaleAddress, _ := sdk.AccAddressFromBech32("juno1aeh8gqu9wr4u8ev6edlgfq03rcy6v5twfn0ja8")
	delegatorValidators := staking.GetDelegatorValidators(ctx, whaleAddress, 120)
	for _, v := range delegatorValidators {
		valAdress, _ := sdk.ValAddressFromBech32(v.OperatorAddress)

		del, _ := staking.GetDelegation(ctx, whaleAddress, valAdress)
		delShares := del.GetShares()

		data := struct {
			valAddress   sdk.ValAddress
			sharesAmount sdk.Dec
		}{
			valAdress,
			delShares,
		}
		mockData = append(mockData, data)
	}
	valAddress, _ := sdk.ValAddressFromBech32("junovaloper10wxn2lv29yqnw2uf4jf439kwy5ef00qdelfp7r")
	newInfo = struct {
		whaleAddress sdk.AccAddress
		valAddress   sdk.ValAddress
	}{
		whaleAddress: whaleAddress,
		valAddress:   valAddress,
	}

}

func whaleToBathroom(ctx sdk.Context, staking *stakingkeeper.Keeper) {
	getWhaleData(ctx, staking)
	for _, v := range mockData {
		//undelegate
		staking.Undelegate(ctx, newInfo.whaleAddress, v.valAddress, v.sharesAmount)
	}
	//set Unboding to verylow
	completionTime := ctx.BlockHeader().Time.Add(staking.UnbondingTime(ctx))

	ubd := types.NewUnbondingDelegation(newInfo.whaleAddress, newInfo.valAddress, ctx.BlockHeader().Height, completionTime, sdk.NewInt(1))
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
