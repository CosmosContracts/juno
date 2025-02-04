package v27

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v27/app/keepers"
)

func CreateV27UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrade", UpgradeName)

		// Run migrations
		logger.Info(fmt.Sprintf("v27: running migrations for: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("v27: post migration check: %v", versionMap))

		// Send MEV profits from module account to community pool
		err = sendMEVtoCommunityPool(ctx, k, logger)
		if err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}

func sendMEVtoCommunityPool(ctx context.Context, k *keepers.AppKeepers, logger log.Logger) error {
	mevModuleAddress := sdk.MustAccAddressFromBech32(mevModuleAccount)
	mevModuleTokenAmount, ok := sdkmath.NewIntFromString(mevModuleAmount)
	if !ok {
		logger.Error(fmt.Sprintf("v27: failed to parse MEV module token amount"))
		return fmt.Errorf("v27: failed to parse MEV module token amount")
	}
	params, err := k.MintKeeper.GetParams(ctx)
	coins := sdk.NewCoins(
		sdk.NewCoin(
			params.MintDenom,
			mevModuleTokenAmount,
		),
	)
	err = k.DistrKeeper.FundCommunityPool(ctx, coins, mevModuleAddress)
	if err != nil {
		logger.Error(fmt.Sprintf("v27: failed to fund community pool with MEV profits: %v", coins))
		return err
	}

	return nil
}
