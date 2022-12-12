package v12

import (
	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v12/app/keepers"
	feesharetypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	oracletypes "github.com/CosmosContracts/juno/v12/x/oracle/types"
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

		// Oracle
		newOracleParams := oracletypes.DefaultParams()
		keepers.OracleKeeper.SetParams(ctx, newOracleParams)

		// TokenFactory
		newTokenFactoryParams := tokenfactorytypes.Params{
			DenomCreationFee: sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(1000000))),
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, newTokenFactoryParams)

		newFeeShareParams := feesharetypes.Params{
			EnableFeeShare:  true,
			DeveloperShares: sdk.NewDecWithPrec(50, 2), // = 50%
			AllowedDenoms:   []string{"ujuno"},
		}
		keepers.FeeShareKeeper.SetParams(ctx, newFeeShareParams)

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
