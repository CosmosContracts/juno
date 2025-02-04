package keeper

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
