package v12

import (
	"strings"

	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v12/app/keepers"
	feesharetypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
)

// Returns "ujuno" if the chainID starts with "juno-" (ex: juno-1 or juno-t1 for local)
// else its the uni testnet
func GetChainsDenomToken(chainID string) string {
	if strings.HasPrefix(chainID, "juno-") {
		return "ujuno"
	}
	return "ujunox"
}

// CreateV12UpgradeHandler makes an upgrade handler for v12 of Juno
func CreateV12UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		nativeDenom := GetChainsDenomToken(ctx.ChainID())
		// transfer module consensus version has been bumped to 2
		// the above is https://github.com/cosmos/ibc-go/blob/v5.1.0/docs/migrations/v3-to-v4.md

		// TokenFactory
		newTokenFactoryParams := tokenfactorytypes.Params{
			DenomCreationFee: sdk.NewCoins(sdk.NewCoin(nativeDenom, sdk.NewInt(1000000))),
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, newTokenFactoryParams)

		// FeeShare
		newFeeShareParams := feesharetypes.Params{
			EnableFeeShare:  true,
			DeveloperShares: sdk.NewDecWithPrec(50, 2), // = 50%
			AllowedDenoms:   []string{nativeDenom},
		}
		keepers.FeeShareKeeper.SetParams(ctx, newFeeShareParams)

		// IBCFee
		vm[ibcfeetypes.ModuleName] = mm.Modules[ibcfeetypes.ModuleName].ConsensusVersion()

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
