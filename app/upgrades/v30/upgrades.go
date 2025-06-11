package v30

import (
	"context"
	"errors"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	log "cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v30/app/keepers"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

func CreateV30UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		logger := sdkCtx.Logger().With("upgrade", UpgradeName)

		// Run migrations
		logger.Info(fmt.Sprintf("v30: running migrations for: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("v30: post migration check: %v", versionMap))

		err = configureFeemarketParams(ctx, k, logger)

		return versionMap, nil
	}
}

func configureFeemarketParams(ctx context.Context, k *keepers.AppKeepers, logger log.Logger) error {
	mintParams, err := k.MintKeeper.GetParams(ctx)
	if err != nil {
		logger.Error("v30: failed to get x/mint params")
		return errors.New("v30: failed to get x/mint params")
	}

	consensusParams, err := k.ConsensusParamsKeeper.ParamsStore.Get(ctx)
	if err != nil {
		logger.Error("v30: failed to get x/consensus params")
		return errors.New("v30: failed to get x/consensus params")
	}

	newFeemarketParams := feemarkettypes.Params{
		Alpha:               feemarkettypes.DefaultAIMDAlpha,
		Beta:                feemarkettypes.DefaultAIMDBeta,
		Gamma:               feemarkettypes.DefaultAIMDGamma,
		Delta:               feemarkettypes.DefaultAIMDDelta,
		MinBaseGasPrice:     feemarkettypes.DefaultMinBaseGasPrice,
		MinLearningRate:     feemarkettypes.DefaultAIMDMinLearningRate,
		MaxLearningRate:     feemarkettypes.DefaultAIMDMaxLearningRate,
		MaxBlockUtilization: uint64(consensusParams.Block.MaxGas),
		Window:              16,
		FeeDenom:            mintParams.MintDenom,
		Enabled:             true,
		DistributeFees:      true,
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err = k.FeeMarketKeeper.SetParams(sdkCtx, newFeemarketParams)
	if err != nil {
		logger.Error("v30: failed to set x/feemarket params")
		return errors.New("v30: failed to set x/feemarket params")
	}

	logger.Info("v30: successfully set x/feemarket params")

	return nil
}
