package veritas

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// UnityContractByteAddress is the bytes of the public key for the address of the Unity contract
// $ junod keys parse juno1nz96hjc926e6a74gyvkwvtt0qhu22wx049c6ph6f4q8kp3ffm9xq5938mr
// human: juno
// bytes: 988BABCB0556B3AEFAA8232CE62D6F05F8A538CFA971A0DF49A80F60C529D94C
const UnityContractByteAddress = "988BABCB0556B3AEFAA8232CE62D6F05F8A538CFA971A0DF49A80F60C529D94C"

const UnityContractPlaceHolderAddress = "5BEF9E5318ED6716A11179C70B06656E9FB91D241A1C594F344B325D9110D94C"

// CreateUpgradeHandler make upgrade handler
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		bondDenom := staking.BondDenom(ctx)

		accAddr, _ := sdk.AccAddressFromBech32(UnityContractPlaceHolderAddress)
		accCoin := bank.GetBalance(ctx, accAddr, bondDenom)

		// get Unity Contract Address and send coin to this address
		destAcc, _ := sdk.AccAddressFromHex(UnityContractByteAddress)
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
