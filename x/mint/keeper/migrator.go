package keeper

import (
	"github.com/CosmosContracts/juno/v16/x/mint/exported"
	v2 "github.com/CosmosContracts/juno/v16/x/mint/migrations/v2"
	v3 "github.com/CosmosContracts/juno/v16/x/mint/migrations/v3"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper         Keeper
	legacySubspace exported.Subspace
}

func NewMigrator(k Keeper, ss exported.Subspace) Migrator {
	return Migrator{
		keeper:         k,
		legacySubspace: ss,
	}
}

// Migrate migrates the x/mint module state from the consensus version
// 1 to version 2
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(ctx, ctx.KVStore(m.keeper.storeKey), m.keeper.cdc)
}

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.Migrate(ctx, ctx.KVStore(m.keeper.storeKey), m.legacySubspace, m.keeper.cdc)
}
