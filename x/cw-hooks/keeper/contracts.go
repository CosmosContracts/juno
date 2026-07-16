package keeper

import (
	"context"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/app/utils"
	"github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
)

func (k Keeper) SetContract(ctx context.Context, key collections.Prefix, info types.ContractInfo) error {
	contractAddr, err := sdk.AccAddressFromBech32(info.ContractAddress)
	if err != nil {
		return err
	}

	return k.Contracts.Set(ctx, types.BuildContractPrimaryKey(key, contractAddr), info)
}

func (k Keeper) IsContractRegistered(ctx context.Context, key collections.Prefix, contractAddr sdk.AccAddress) (bool, error) {
	return k.Contracts.Has(ctx, types.BuildContractPrimaryKey(key, contractAddr))
}

func (k Keeper) GetAllContracts(ctx context.Context, key collections.Prefix) (list []types.ContractInfo, err error) {
	iter, err := k.Contracts.Iterate(ctx, collections.NewPrefixedPairRange[[]byte, sdk.AccAddress](key.Bytes()))
	if err != nil {
		return nil, err
	}
	values, err := iter.Values()
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (k Keeper) DeleteContract(ctx context.Context, key collections.Prefix, contractAddr sdk.AccAddress) error {
	return k.Contracts.Remove(ctx, types.BuildContractPrimaryKey(key, contractAddr))
}

func (k Keeper) ExecuteMessageOnContracts(ctx context.Context, key collections.Prefix, msgBz []byte) error {
	p := k.GetParams(ctx)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	contracts, err := k.GetAllContracts(ctx, key)
	if err != nil {
		return err
	}

	for _, c := range contracts {
		gasLimitCtx := sdkCtx.WithGasMeter(storetypes.NewGasMeter(p.ContractGasLimit))
		addr, err := sdk.AccAddressFromBech32(c.ContractAddress)
		if err != nil {
			return err
		}

		var execErr error
		utils.ExecuteContract(k.GetContractKeeper(), gasLimitCtx, addr, msgBz, &execErr)
		if execErr != nil {
			k.Logger(ctx).Debug("ExecuteMessageOnContracts err", "error", execErr, "contract", c.ContractAddress)
			if err := k.handleContractFailure(ctx, key, addr, c, execErr, p.ContractFailureRemovalThreshold); err != nil {
				return err
			}
			continue
		}

		if err := k.resetFailureCounter(ctx, key, c); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) dispatchHookMessage(ctx context.Context, keyPrefix collections.Prefix, msgBz []byte, hookName string) error {
	if err := k.ExecuteMessageOnContracts(ctx, keyPrefix, msgBz); err != nil {
		k.Logger(ctx).Error("cw-hook contract execution failed", "hook", hookName, "error", err)
	}

	return nil
}

func (k Keeper) handleContractFailure(
	ctx context.Context,
	key collections.Prefix,
	addr sdk.AccAddress,
	info types.ContractInfo,
	execErr error,
	threshold uint64,
) error {
	info.FailureCounter++
	info.LatestError = execErr.Error()

	if threshold > 0 && uint64(info.FailureCounter) >= threshold {
		k.Logger(ctx).Info("removing contract due to repeated failures", "contract", info.ContractAddress, "module", key)
		return k.DeleteContract(ctx, key, addr)
	}

	return k.SetContract(ctx, key, info)
}

func (k Keeper) resetFailureCounter(ctx context.Context, key collections.Prefix, info types.ContractInfo) error {
	if info.FailureCounter == 0 && info.LatestError == "" {
		return nil
	}

	info.FailureCounter = 0
	info.LatestError = ""
	return k.SetContract(ctx, key, info)
}
