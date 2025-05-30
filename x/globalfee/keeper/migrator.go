package keeper

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper Keeper
}

func NewMigrator(k Keeper) Migrator {
	return Migrator{
		keeper: k,
	}
}

// Migrate1to2 migrates the x/mint module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/mint
// module state.
// func (m Migrator) Migrate1to2(ctx sdk.Context) error {
// 	return v2.Migrate(ctx, ctx.KVStore(m.keeper.storeKey), m.keeper.cdc, m.bondDenom)
// }
