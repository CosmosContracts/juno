package v18

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v20/app/keepers"
	"github.com/CosmosContracts/juno/v20/app/upgrades"
	cwhookstypes "github.com/CosmosContracts/juno/v20/x/cw-hooks/types"
	feepaytypes "github.com/CosmosContracts/juno/v20/x/feepay/types"
)

func CreateV18UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// Run migrations
		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// x/cw-hooks
		gasLimit := uint64(250_000)
		if err := k.CWHooksKeeper.SetParams(ctx, cwhookstypes.NewParams(gasLimit)); err != nil {
			return nil, err
		}

		// x/feepay
		if err := k.FeePayKeeper.SetParams(ctx, feepaytypes.Params{
			EnableFeepay: true,
		}); err != nil {
			return nil, err
		}

		return versionMap, err
	}
}
