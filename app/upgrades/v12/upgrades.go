package v12

import (
	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	wasm "github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmosContracts/juno/v12/app/keepers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateV12UpgradeHandler makes an upgrade handler for v12 of Juno
func CreateV12UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// transfer module consensus version has been bumped to 2
		// the above is https://github.com/cosmos/ibc-go/blob/v5.1.0/docs/migrations/v3-to-v4.md

		// Set the creation fee for the token factory to cost 1 JUNO token
		newTokenFactoryParams := tokenfactorytypes.Params{
			DenomCreationFee: sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(1000000))),
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, newTokenFactoryParams)

		// https://github.com/CosmWasm/wasmd/blob/016e3bc06b0e2a7c680dc1c9f78104ec931dde41/x/wasm/module_integration_test.go
		vm[wasm.ModuleName] = 1

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
