package v30

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	log "cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v30/app/keepers"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	votingsnapshottypes "github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
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
		if err != nil {
			return nil, err
		}

		err = configureCWHooksParams(ctx, k, logger)
		if err != nil {
			return nil, err
		}

		// x/voting-snapshot seeding happens in the module's InitGenesis, which
		// RunMigrations above already executed for the newly-added store —
		// params are set and BackfillFromStaking has walked the delegation
		// set exactly once. Do NOT backfill again here: the walk is
		// O(n_delegations) and doubling it doubles the upgrade-block cost.
		// Cheap insurance in case module-init ordering ever changes: make
		// sure params exist so reads/hooks never hit a missing-params error.
		hasParams, err := k.VotingSnapshotKeeper.Params.Has(ctx)
		if err != nil {
			return nil, errorsmod.Wrap(err, "v30: failed to check x/voting-snapshot params")
		}
		if !hasParams {
			logger.Error("v30: x/voting-snapshot params missing after RunMigrations (InitGenesis did not run?); setting defaults")
			if err := k.VotingSnapshotKeeper.Params.Set(ctx, votingsnapshottypes.DefaultParams()); err != nil {
				return nil, errorsmod.Wrap(err, "v30: failed to set default x/voting-snapshot params")
			}
		}

		return versionMap, nil
	}
}

func configureFeemarketParams(ctx context.Context, k *keepers.AppKeepers, logger log.Logger) error {
	stakingParams, err := k.StakingKeeper.GetParams(ctx)
	if err != nil {
		logger.Error("v30: failed to get x/staking params")
		return errorsmod.Wrap(err, "v30: failed to get x/staking params")
	}

	consensusParams, err := k.ConsensusParamsKeeper.ParamsStore.Get(ctx)
	if err != nil {
		logger.Error("v30: failed to get x/consensus params")
		return errorsmod.Wrap(err, "v30: failed to get x/consensus params")
	}

	// MaxGas == -1 ("unbounded") would cast to 2^64-1 below and seed
	// feemarket with a nonsense block-utilization ceiling, sending AIMD
	// base-fee adjustment haywire. juno-1 sets a positive max_gas, but
	// guard against operators running this binary on a chain that left
	// max_gas unbounded.
	if consensusParams.Block == nil || consensusParams.Block.MaxGas <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "v30: consensus block.max_gas must be a positive value before feemarket init")
	}

	newFeemarketParams := feemarkettypes.Params{
		Alpha:               sdkmath.LegacyMustNewDecFromStr("0.004"),
		Beta:                sdkmath.LegacyMustNewDecFromStr("0.983"),
		Gamma:               sdkmath.LegacyMustNewDecFromStr("0.2"),
		Delta:               sdkmath.LegacyMustNewDecFromStr("0.00000000000125"),
		MinBaseGasPrice:     sdkmath.LegacyMustNewDecFromStr("0.075"),
		MinLearningRate:     sdkmath.LegacyMustNewDecFromStr("0.0015"),
		MaxLearningRate:     sdkmath.LegacyMustNewDecFromStr("0.05"),
		MaxBlockUtilization: uint64(consensusParams.Block.MaxGas),
		Window:              60,
		FeeDenom:            stakingParams.BondDenom,
		Enabled:             true,
		DistributeFees:      true,
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err = k.FeeMarketKeeper.SetParams(sdkCtx, newFeemarketParams)
	if err != nil {
		return errorsmod.Wrap(err, "v30: failed to set x/feemarket params")
	}

	newState := feemarkettypes.NewState(
		newFeemarketParams.Window,
		newFeemarketParams.MinBaseGasPrice,
		newFeemarketParams.MinLearningRate,
	)
	if err := k.FeeMarketKeeper.SetState(sdkCtx, newState); err != nil {
		return errorsmod.Wrap(err, "v30: failed to rebuild x/feemarket state")
	}

	logger.Info("v30: successfully configured x/feemarket")

	return nil
}

func configureCWHooksParams(ctx context.Context, k *keepers.AppKeepers, logger log.Logger) error {
	params, err := k.CWHooksKeeper.Params.Get(ctx)
	if err != nil {
		logger.Error("v30: failed to get x/cw-hooks params")
		return errorsmod.Wrap(err, "v30: failed to get x/cw-hooks params")
	}

	params.ContractFailureRemovalThreshold = 3

	if err := k.CWHooksKeeper.Params.Set(ctx, params); err != nil {
		logger.Error("v30: failed to set x/cw-hooks params")
		return errorsmod.Wrap(err, "v30: failed to set x/cw-hooks params")
	}

	logger.Info("v30: successfully set x/cw-hooks params")

	return nil
}
