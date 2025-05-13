package v29

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v29/app/keepers"
)

func CreateV29UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrade", UpgradeName)

		// Run migrations
		logger.Info(fmt.Sprintf("v29: running migrations for: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("v29: post migration check: %v", versionMap))

		// Configure x/gov params, fixes expedited proposal deposit being way too low
		err = configureGovV1Params(ctx, k, logger)
		if err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}

func configureGovV1Params(ctx context.Context, k *keepers.AppKeepers, logger log.Logger) error {
	govParams, err := k.GovKeeper.Keeper.Params.Get(ctx)
	if err != nil {
		logger.Error("v29: failed to get x/gov params")
		return fmt.Errorf("v29: failed to get x/gov params")
	}

	mintParams, err := k.MintKeeper.GetParams(ctx)
	if err != nil {
		logger.Error("v29: failed to get x/mint params")
		return fmt.Errorf("v29: failed to get x/mint params")
	}

	expeditedMinDepositInt, ok := sdkmath.NewIntFromString(expeditedMinDeposit)
	if ok != true {
		logger.Error("v29: failed to parse expedited min deposit")
		return fmt.Errorf("v29: failed to parse expedited min deposit")
	}

	govParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(mintParams.MintDenom, expeditedMinDepositInt))

	err = k.GovKeeper.Keeper.Params.Set(ctx, govParams)
	if err != nil {
		logger.Error("v29: failed to set updated x/gov params")
		return fmt.Errorf("v29: failed to set updated x/gov params")
	}

	logger.Info("v29: successfully set updated x/gov params")

	return nil
}
