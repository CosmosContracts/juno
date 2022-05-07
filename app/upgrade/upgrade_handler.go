package veritas

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// UnityContractAddress is the address of the Unity contract
const UnityContractAddress = "juno1nz96hjc926e6a74gyvkwvtt0qhu22wx049c6ph6f4q8kp3ffm9xq5938mr"

// UnityContractPlaceHolderAddress is the address where the funds were sent in prop 20
const UnityContractPlaceHolderAddress = "juno1t0heu5cca4n3dgg308rskpn9d60mj8fyrgw9jne5fve9mygsm9xqkcrpl2"

// CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		bondDenom := staking.BondDenom(ctx)

		accAddr, _ := sdk.AccAddressFromBech32(UnityContractPlaceHolderAddress)
		accCoin := bank.GetBalance(ctx, accAddr, bondDenom)

		// get Unity Contract Address and send coin to this address
		destAcc, _ := sdk.AccAddressFromBech32(UnityContractAddress)
		err := bank.SendCoins(ctx, accAddr, destAcc, sdk.NewCoins(accCoin))
		if err != nil {
			panic(err)
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
