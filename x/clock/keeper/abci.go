package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v30/app/utils"
	"github.com/CosmosContracts/juno/v30/x/clock/types"
)

var endBlockSudoMessage = []byte(types.EndBlockSudoMessage)

// EndBlocker executes on contracts at the end of the block.
func EndBlocker(ctx context.Context, k Keeper) error {
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyEndBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	logger := k.Logger(ctx)
	p := k.GetParams(ctx)

	// Get all contracts
	contracts, err := k.GetAllContracts(ctx)
	if err != nil {
		logger.Error("Failed to get contracts", "error", err)
		return err
	}

	// Track errors
	errorExecs := make([]string, 0, len(contracts))
	errorExists := false

	// Execute all contracts that are not jailed
	for _, contract := range contracts {
		// Skip jailed contracts
		if contract.IsJailed {
			continue
		}

		// Get sdk.AccAddress from contract address
		contractAddr := sdk.MustAccAddressFromBech32(contract.ContractAddress)

		// Create context with gas limit
		childCtx := sdkCtx.WithGasMeter(storetypes.NewGasMeter(p.ContractGasLimit))

		// Execute contract with a per-iteration error so a prior contract's
		// failure never cascades into jailing a healthy contract.
		var execErr error
		utils.ExecuteContract(k.GetContractKeeper(), childCtx, contractAddr, endBlockSudoMessage, &execErr)
		if execErr != nil {
			// Flag error
			errorExists = true
			errorExecs = append(errorExecs, contract.ContractAddress)

			// Attempt to jail contract, log error if present
			if jailErr := k.SetJailStatus(ctx, contract.ContractAddress, true); jailErr != nil {
				logger.Error("Failed to jail contract", "contract", contract.ContractAddress, "error", jailErr)
			}
		}
	}

	// Log errors if present
	if errorExists {
		logger.Error("Failed to execute contracts", "contracts", errorExecs)
	}

	return nil
}
