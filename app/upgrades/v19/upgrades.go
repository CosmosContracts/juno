package v19

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	decorators "github.com/CosmosContracts/juno/v19/app/decorators"
	"github.com/CosmosContracts/juno/v19/app/keepers"
	"github.com/CosmosContracts/juno/v19/app/upgrades"
	wasmlctypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
)

func CreateV19UpgradeHandler(
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

		// Change Rate Decorator Migration
		// Ensure all Validators have a max change rate of 5%
		maxChangeRate := sdk.MustNewDecFromStr(decorators.MaxChangeRate)
		validators := k.StakingKeeper.GetAllValidators(ctx)

		for _, validator := range validators {
			if validator.Commission.MaxChangeRate.GT(maxChangeRate) {
				validator.Commission.MaxChangeRate.Set(maxChangeRate)
				k.StakingKeeper.SetValidator(ctx, validator)
			}
		}

		// https://github.com/cosmos/ibc-go/blob/main/docs/docs/03-light-clients/04-wasm/03-integration.md
		params := k.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, wasmlctypes.Wasm)
		k.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		return versionMap, err
	}
}
