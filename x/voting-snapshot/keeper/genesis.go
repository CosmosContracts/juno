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

// ExportGenesis exports Params only — the snapshot index (VotingPower /
// TotalPower maps) is intentionally NOT exported.
//
// This is a documented in-place-upgrade-only assumption: Juno upgrades
// via x/upgrade in-place migrations, so the snapshot history never needs
// to round-trip through genesis JSON (it can be millions of rows). If
// the chain is ever restarted from an exported genesis, InitGenesis's
// BackfillFromStaking reseeds every delegator's power and the total at
// the restart height from live staking state — current voting power is
// correct immediately; only pre-restart *history* is lost, which
// matches what proposals opened before a hard restart can expect.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.GenesisState{Params: params}, nil
}
