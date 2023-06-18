package v14

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v16/app/keepers"
	"github.com/CosmosContracts/juno/v16/app/upgrades"
	globalfeetypes "github.com/CosmosContracts/juno/v16/x/globalfee/types"
)

func CreateV14UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))

		// Run migrations
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// GlobalFee
		minGasPrices := sdk.DecCoins{
			// 0.0025ujuno
			sdk.NewDecCoinFromDec(nativeDenom, sdk.NewDecWithPrec(25, 4)),
			// 0.001 ATOM CHANNEL-1 -> `junod q ibc-transfer denom-trace ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9`
			sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.NewDecWithPrec(1, 3)),
		}
		s, ok := keepers.ParamsKeeper.GetSubspace(globalfeetypes.ModuleName)
		if !ok {
			panic("global fee params subspace not found")
		}
		s.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)
		logger.Info(fmt.Sprintf("upgraded global fee params to %s", minGasPrices))

		return versionMap, err
	}
}
