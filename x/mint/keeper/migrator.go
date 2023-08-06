package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/CosmosContracts/juno/v16/x/mint/migrations/v2"
	v3 "github.com/CosmosContracts/juno/v16/x/mint/migrations/v3"
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

// Deprecated: Migrate1to2 was an old upgrade.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(ctx, ctx.KVStore(m.keeper.storeKey), m.keeper.cdc)
}

// Migrate1to2 migrates the x/mint module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/mint
// module state.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.Migrate(ctx, ctx.KVStore(m.keeper.storeKey), m.keeper.cdc, m.bondDenom)
}
