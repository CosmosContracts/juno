package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/cw-hooks/keeper"
	v2 "github.com/CosmosContracts/juno/v31/x/cw-hooks/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *keeper.Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(cwHooksKeeper *keeper.Keeper) Migrator {
	return Migrator{
		keeper: cwHooksKeeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStoreToCollections(ctx, m.keeper.GetStoreService(), m.keeper.GetCdc(), m.keeper)
}
