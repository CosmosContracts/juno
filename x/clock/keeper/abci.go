package keeper

import (
	"context"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	helpers "github.com/CosmosContracts/juno/v27/app/helpers"
	"github.com/CosmosContracts/juno/v27/x/clock/types"
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
	errorExecs := make([]string, len(contracts))
	errorExists := false

	// Execute all contracts that are not jailed
	for idx, contract := range contracts {

		// Skip jailed contracts
		if contract.IsJailed {
			continue
		}

		// Get sdk.AccAddress from contract address
		contractAddr := sdk.MustAccAddressFromBech32(contract.ContractAddress)
		if handleError(ctx, k, logger, errorExecs, &errorExists, err, idx, contract.ContractAddress) {
			continue
		}

		// Create context with gas limit
		childCtx := sdkCtx.WithGasMeter(storetypes.NewGasMeter(p.ContractGasLimit))

		// Execute contract
		helpers.ExecuteContract(k.GetContractKeeper(), childCtx, contractAddr, endBlockSudoMessage, &err)
		if handleError(ctx, k, logger, errorExecs, &errorExists, err, idx, contract.ContractAddress) {
			continue
		}
	}

	// Log errors if present
	if errorExists {
		logger.Error("Failed to execute contracts", "contracts", errorExecs)
	}

	return nil
}

// Function to handle contract execution errors. Returns true if error is present, false otherwise.
func handleError(
	ctx context.Context,
	k Keeper,
	logger log.Logger,
	errorExecs []string,
	errorExists *bool,
	err error,
	idx int,
	contractAddress string,
) bool {
	// Check if error is present
	if err != nil {

		// Flag error
		*errorExists = true
		errorExecs[idx] = contractAddress

		// Attempt to jail contract, log error if present
		err := k.SetJailStatus(ctx, contractAddress, true)
		if err != nil {
			logger.Error("Failed to jail contract", "contract", contractAddress, "error", err)
		}
	}

	return err != nil
}
