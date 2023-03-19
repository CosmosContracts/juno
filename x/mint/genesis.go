package mint

import (
	mintkeeper "github.com/CosmosContracts/juno/v13/x/mint/keeper"
	minttypes "github.com/CosmosContracts/juno/v13/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper mintkeeper.Keeper, ak minttypes.AccountKeeper, data *minttypes.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)
	ak.GetModuleAccount(ctx, minttypes.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper mintkeeper.Keeper) *minttypes.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	return minttypes.NewGenesisState(minter, params)
}
