package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

func (k Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) error {
	if err := k.Params.Set(ctx, gs.Params); err != nil {
		return err
	}
	// Seed snapshots from staking state after genesis bonding has
	// settled. We rely on `votingsnapshot` running AFTER staking in
	// `orderInitBlockers` (verified in app/modules.go); at this point
	// `stakingKeeper.TotalBondedTokens` reflects the chain's finalized
	// genesis stake. Without this, the AfterDelegationModified hooks
	// fired during staking's own InitGenesis observe TotalBondedTokens
	// = 0 (the bond pool isn't aggregated yet), so the chain-wide
	// total snapshot at genesis would be wrong even though per-
	// delegator snapshots come out right.
	return k.BackfillFromStaking(ctx)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.GenesisState{Params: params}, nil
}
