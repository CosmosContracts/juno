package bindings_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"math/rand"

	sdkmath "cosmossdk.io/math"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/CosmosContracts/juno/v27/app"
	"github.com/CosmosContracts/juno/v27/testutil/setup"
)

func CreateTestInput(t *testing.T) (*app.App, sdk.Context) {
	randomInt := rand.Intn(2048)
	homeDir := fmt.Sprintf("%d", randomInt)
	app := setup.Setup(false, homeDir, "juno-1", t)
	ctx := app.BaseApp.NewContext(false)
	return app, ctx
}

func FundAccount(t *testing.T, ctx context.Context, junoapp *app.App, acct sdk.AccAddress) {
	err := banktestutil.FundAccount(ctx, junoapp.AppKeepers.BankKeeper, acct, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdkmath.NewInt(10000000000)),
	))
	require.NoError(t, err)
}

// we need to make this deterministic (same every test run), as content might affect gas costs
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func RandomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func RandomBech32AccountAddress() string {
	return RandomAccountAddress().String()
}

func storeReflectCode(t *testing.T, ctx context.Context, junoapp *app.App, addr sdk.AccAddress) uint64 {
	wasmCode, err := os.ReadFile("./testdata/token_reflect.wasm")
	require.NoError(t, err)

	contractKeeper := keeper.NewDefaultPermissionKeeper(junoapp.AppKeepers.WasmKeeper)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	codeID, _, err := contractKeeper.Create(sdkCtx, addr, wasmCode, nil)
	require.NoError(t, err)

	return codeID
}

func instantiateReflectContract(t *testing.T, ctx context.Context, junoapp *app.App, funder sdk.AccAddress) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(junoapp.AppKeepers.WasmKeeper)
	codeID := uint64(1)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	addr, _, err := contractKeeper.Instantiate(sdkCtx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}

func fundAccount(t *testing.T, ctx context.Context, junoapp *app.App, addr sdk.AccAddress, coins sdk.Coins) {
	err := banktestutil.FundAccount(
		ctx,
		junoapp.AppKeepers.BankKeeper,
		addr,
		coins,
	)
	require.NoError(t, err)
}

func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.App, sdk.Context) {
	junoApp, ctx := CreateTestInput(t)
	wasmKeeper := junoApp.AppKeepers.WasmKeeper

	storeReflectCode(t, ctx, junoApp, addr)

	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return junoApp, ctx
}
