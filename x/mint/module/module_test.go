package mint_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/nft/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/CosmosContracts/juno/v27/x/mint/types"
)

// TestItCreatesModuleAccountOnInitBlock tests that the module account is created on InitBlock
func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper authkeeper.AccountKeeper

	app, err := simtestutil.SetupAtGenesis(testutil.AppConfig, &accountKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
