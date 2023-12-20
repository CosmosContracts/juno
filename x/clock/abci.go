package clock

import (
	"time"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v19/x/clock/keeper"
	"github.com/CosmosContracts/juno/v19/x/clock/types"

	helpers "github.com/CosmosContracts/juno/v19/app/helpers"
)

var endBlockSudoMessage = []byte(types.EndBlockSudoMessage)

// EndBlocker executes on contracts at the end of the block.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := k.Logger(ctx)
	p := k.GetParams(ctx)

	// Get all contracts
	contracts, err := k.GetAllContracts(ctx)
	if err != nil {
		logger.Error("Failed to get contracts", "error", err)
		return
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
		childCtx := ctx.WithGasMeter(sdk.NewGasMeter(p.ContractGasLimit))

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
}

// Function to handle contract execution errors. Returns true if error is present, false otherwise.
func handleError(
	ctx sdk.Context,
	k keeper.Keeper,
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
