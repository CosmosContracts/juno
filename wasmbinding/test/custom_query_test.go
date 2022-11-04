package wasmbinding_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CosmosContracts/juno/v11/app"
	appparams "github.com/CosmosContracts/juno/v11/app/params"
	"github.com/CosmosContracts/juno/v11/x/oracle/wasm"
	oracle "github.com/CosmosContracts/juno/v11/x/oracle/wasm"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.App, sdk.Context) {
	junoApp, ctx := CreateTestInput(t)
	wasmKeeper := junoApp.WasmKeeper()
	storeOracleQuerierCode(t, ctx, junoApp, addr)

	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return junoApp, ctx
}

func TestQueryExchangeRate(t *testing.T) {
	actor := RandomAccountAddress()
	junoApp, ctx := SetupCustomApp(t, actor)

	oracleQuerier := instantiateOracleQuerierContract(t, ctx, junoApp, actor)
	require.NotEmpty(t, oracleQuerier)

	actorAmount := sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(100000000000000)))
	fundAccount(t, ctx, junoApp, actor, actorAmount)

	ExchangeRateC := sdk.NewDec(1700)
	ExchangeRateB := sdk.NewDecWithPrec(17, 1)
	ExchangeRateD := sdk.NewDecWithPrec(19, 1)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "a", sdk.NewDec(1))
	junoApp.OracleKeeper.SetExchangeRate(ctx, "b", ExchangeRateB)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "c", ExchangeRateC)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "d", ExchangeRateD)

	msg := json.RawMessage(`{"get_exchange_rate": {"denom":"b"}}`)
	err := executeRawCustom(t, ctx, junoApp, oracleQuerier, actor, msg, sdk.Coin{})
	require.NoError(t, err)

	query := oracle.OracleQuery{
		ExchangeRate: &wasm.ExchangeRateQueryParams{
			Denom: "b",
		},
	}
	resp := oracle.ExchangeRateQueryResponse{}
	queryRate(t, ctx, junoApp, oracleQuerier, query, &resp)

}

func queryRate(t *testing.T, ctx sdk.Context, junoApp *app.App, contract sdk.AccAddress, request oracle.OracleQuery, response interface{}) {
	queryBz, err := json.Marshal(request)
	require.NoError(t, err)

	resBz, err := junoApp.WasmKeeper().QuerySmart(ctx, contract, queryBz)
	var rate string
	json.Unmarshal(resBz, &rate)
	require.NoError(t, err)
	require.Equal(t, rate, "1.7")
}

func storeOracleQuerierCode(t *testing.T, ctx sdk.Context, junoApp *app.App, addr sdk.AccAddress) {
	govKeeper := junoApp.GovKeeper
	wasmCode, err := os.ReadFile("../testdata/oracle_querier.wasm")
	require.NoError(t, err)

	src := wasmtypes.StoreCodeProposalFixture(func(p *wasmtypes.StoreCodeProposal) {
		p.RunAs = addr.String()
		p.WASMByteCode = wasmCode
	})

	// when stored
	storedProposal, err := govKeeper.SubmitProposal(ctx, src)
	require.NoError(t, err)

	// and proposal execute
	handler := govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
	err = handler(ctx, storedProposal.GetContent())
	require.NoError(t, err)
}

func instantiateOracleQuerierContract(t *testing.T, ctx sdk.Context, junoApp *app.App, funder sdk.AccAddress) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(junoApp.WasmKeeper())
	codeID := uint64(1)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}

func fundAccount(t *testing.T, ctx sdk.Context, juno *app.App, addr sdk.AccAddress, coins sdk.Coins) {
	err := simapp.FundAccount(
		juno.BankKeeper,
		ctx,
		addr,
		coins,
	)
	require.NoError(t, err)
}

func executeRawCustom(t *testing.T, ctx sdk.Context, juno *app.App, contract sdk.AccAddress, sender sdk.AccAddress, msg json.RawMessage, funds sdk.Coin) error {
	oracleBz, err := json.Marshal(msg)
	require.NoError(t, err)
	// no funds sent if amount is 0
	var coins sdk.Coins
	if !funds.Amount.IsNil() {
		coins = sdk.Coins{funds}
	}

	contractKeeper := keeper.NewDefaultPermissionKeeper(juno.WasmKeeper())
	_, err = contractKeeper.Execute(ctx, contract, sender, oracleBz, coins)
	return err
}
