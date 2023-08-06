package v13

import (
	"fmt"

	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
	// ICA
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v16/app/keepers"
	"github.com/CosmosContracts/juno/v16/app/upgrades"
	// types
	feesharetypes "github.com/CosmosContracts/juno/v16/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v16/x/tokenfactory/types"
)

func CreateV13UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// transfer module consensus version has been bumped to 2
		// the above is https://github.com/cosmos/ibc-go/blob/v5.1.0/docs/migrations/v3-to-v4.md
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// ICA - https://github.com/CosmosContracts/juno/blob/integrate_ica_changes/app/app.go#L846-L885
		// vm[icatypes.ModuleName] = mm.Modules[icatypes.ModuleName].ConsensusVersion()
		logger.Info("upgraded icatypes version")

		// Update ICS27 Host submodule params
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			// https://github.com/cosmos/ibc-go/blob/v4.2.0/docs/apps/interchain-accounts/parameters.md#allowmessages
			AllowMessages: []string{"*"},
		}

		// IBCFee
		// vm[ibcfeetypes.ModuleName] = mm.Modules[ibcfeetypes.ModuleName].ConsensusVersion()
		logger.Info(fmt.Sprintf("ibcfee module version %s set", fmt.Sprint(vm[ibcfeetypes.ModuleName])))

		// Run migrations
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		// New modules run AFTER the migrations, so to set the correct params after the default.

		// Set ICA Params
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)
		keepers.ICAControllerKeeper.SetParams(ctx, icacontrollertypes.Params{ControllerEnabled: true})
		logger.Info("upgraded ICAHostKeeper params")

		// TokenFactory
		newTokenFactoryParams := tokenfactorytypes.Params{
			DenomCreationFee: sdk.NewCoins(sdk.NewCoin(nativeDenom, sdk.NewInt(1000000))),
		}
		if err := keepers.TokenFactoryKeeper.SetParams(ctx, newTokenFactoryParams); err != nil {
			return nil, err
		}
		logger.Info("set tokenfactory params")

		// FeeShare
		newFeeShareParams := feesharetypes.Params{
			EnableFeeShare:  true,
			DeveloperShares: sdk.NewDecWithPrec(50, 2), // = 50%
			AllowedDenoms:   []string{nativeDenom},
		}
		if err := keepers.FeeShareKeeper.SetParams(ctx, newFeeShareParams); err != nil {
			return nil, err
		}
		logger.Info("set feeshare params")

		// Packet Forward middleware initial params
		keepers.PacketForwardKeeper.SetParams(ctx, packetforwardtypes.DefaultParams())

		return versionMap, err
	}
}
