package clock

import (
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/clock/keeper"
	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

// EndBlocker executes on contracts at the end of the block.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	message := []byte(types.EndBlockSudoMessage)

	p := k.GetParams(ctx)

	// Get all contracts
	contracts, err := k.GetAllContracts(ctx)
	if err != nil {
		log.Printf("[x/clock] Failed to get contracts: %v", err)
		return
	}

	// Track errors
	errorExecs := make([]string, len(contracts))
	errorExists := false

	// Function to handle contract execution errors. Returns true if error is present, false otherwise.
	handleError := func(err error, idx int, contractAddress string) bool {

		// Check if error is present
		if err != nil {

			// Flag error
			errorExists = true
			errorExecs[idx] = contractAddress

			// Attempt to jail contract, log error if present
			err := k.SetJailStatus(ctx, contractAddress, true)
			if err != nil {
				log.Printf("[x/clock] Failed to Error Contract %s: %v", contractAddress, err)
			}
		}

		return err != nil
	}

	// Execute all contracts that are not jailed
	for idx, contract := range contracts {

		// Skip jailed contracts
		if contract.IsJailed {
			continue
		}

		// Get sdk.AccAddress from contract address
		contractAddr := sdk.MustAccAddressFromBech32(contract.ContractAddress)
		if handleError(err, idx, contract.ContractAddress) {
			continue
		}

		// Create context with gas limit
		childCtx := ctx.WithGasMeter(sdk.NewGasMeter(p.ContractGasLimit))

		// Execute contract
		_, err = k.GetContractKeeper().Sudo(childCtx, contractAddr, message)
		if handleError(err, idx, contract.ContractAddress) {
			continue
		}
	}

	// Log errors if present
	if errorExists {
		log.Printf("[x/clock] Execute Errors: %v", errorExecs)
	}
}
