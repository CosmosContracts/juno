package v23

import (
	"fmt"

	"github.com/CosmosContracts/juno/v23/app/keepers"
	"github.com/CosmosContracts/juno/v23/app/upgrades"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
)

func CreateV23UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// migrate ICQ params
		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			if subspace.Name() == icqtypes.ModuleName {
				keyTable = icqtypes.ParamKeyTable()
			} else {
				continue
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Migrate Tendermint consensus parameters from x/params module to a deprecated x/consensus module.
		// The old params module is required to still be imported in your app.go in order to handle this migration.
		baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)

		// Run migrations
		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		
		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// convert pob builder account to an actual module account
		// during upgrade from v15 to v16 it wasn't correctly created, and since it received tokens on mainnet is not a base account
		// it's like this on both mainnet and uni
		if ctx.ChainID() == "juno-1" || ctx.ChainID() == "uni-6" {
			logger.Info("converting x/pob builder module account")

			address := sdk.MustAccAddressFromBech32("juno1ma4sw9m2nvtucny6lsjhh4qywvh86zdh5dlkd4")

			acc := keepers.AccountKeeper.NewAccount(
				ctx,
				authtypes.NewModuleAccount(
					authtypes.NewBaseAccountWithAddress(address),
					"builder",
				),
			)
			keepers.AccountKeeper.SetAccount(ctx, acc)

			logger.Info("x/pob builder module address is now a module account")
		}

		return versionMap, err
	}
}
