package mint_test

import (
	"testing"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v15/app"
	"github.com/CosmosContracts/juno/v15/x/mint/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	app := app.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: []byte("{}"),
		},
	)

	acc := app.AppKeepers.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
