package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/CosmosContracts/juno/v16/x/globalfee/migrations/v2"
)

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper    Keeper
	bondDenom string
}

func NewMigrator(k Keeper, bondDenom string) Migrator {
	return Migrator{
		keeper:    k,
		bondDenom: bondDenom,
	}
}

// Migrate1to2 migrates the x/mint module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/mint
// module state.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(ctx, ctx.KVStore(m.keeper.storeKey), m.keeper.cdc, m.bondDenom)
}
