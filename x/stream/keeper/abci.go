package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

// PreBlocker is called before the beginning of each block
// It updates the query context that will be used by streaming queries
func (k *Keeper) PreBlocker(ctx context.Context) error {
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyPreBlocker)

	// Store the current context for use by streaming queries
	k.SetQueryContext(ctx)
	k.logger.Debug("PreBlocker: stored query context for streaming RPCs")
	return nil
}
