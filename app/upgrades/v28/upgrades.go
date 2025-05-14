package v28

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v29/app/keepers"
)

func CreateV28UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrade", UpgradeName)

		// Run migrations
		logger.Info(fmt.Sprintf("v28: running migrations for: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("v28: post migration check: %v", versionMap))

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

	// get mev module account balance
	mintParams, err := k.MintKeeper.GetParams(ctx)
	if err != nil {
		logger.Error("v28: failed to get x/mint params")
		return errors.New("v28: failed to get x/mint params")
	}

	mevModuleTokenAmount := k.BankKeeper.GetBalance(ctx, mevModuleAddress, mintParams.MintDenom)

	// skip if balance is 0 (testnet etc)
	if !mevModuleTokenAmount.IsZero() {
		coins := sdk.NewCoins(mevModuleTokenAmount)
		err = k.DistrKeeper.FundCommunityPool(ctx, coins, mevModuleAddress)
		if err != nil {
			logger.Error(fmt.Sprintf("v28: failed to fund community pool with MEV profits: %v", coins))
			return err
		}
	}
	return nil
}
