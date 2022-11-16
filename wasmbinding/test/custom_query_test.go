package wasmbinding_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CosmosContracts/juno/v12/app"
	appparams "github.com/CosmosContracts/juno/v12/app/params"
	"github.com/CosmosContracts/juno/v12/wasmbinding"
	"github.com/CosmosContracts/juno/v12/wasmbinding/bindings"
	"github.com/CosmosContracts/juno/v12/x/oracle/wasm"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

func mustLoad(path string) []byte {
	bz, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return bz
}

var (
	oracleCode = mustLoad("../testdata/oracle_querier.wasm")
)

func TestQueryExchangeRate(t *testing.T) {
	actor := app.RandomAccountAddress()
	junoApp := app.Setup(t, false, 1)
	ctx := junoApp.BaseApp.NewContext(false, tmtypes.Header{Height: 1, ChainID: "kujira-1", Time: time.Now().UTC()})

	wasmKeeper := junoApp.GetWasmKeeper()
	plugin := wasmbinding.NewQueryPlugin(junoApp.OracleKeeper)
	querier := wasmbinding.CustomQuerier(plugin)

	// Store Oracle querier
	storeOracleQuerierCode(t, ctx, junoApp, actor, oracleCode)

	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	// Init Oracle querier
	oracleQuerier := instantiateOracleQuerierContract(t, ctx, junoApp, actor)
	require.NotEmpty(t, oracleQuerier)

	actorAmount := sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(100000000000000)))
	err := simapp.FundAccount(
		junoApp.BankKeeper,
		ctx,
		oracleQuerier,
		actorAmount,
	)
	require.NoError(t, err)

	// Set exchange rate for coin "a"
	ExchangeRate := sdk.NewDecWithPrec(1792, 2)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "a", ExchangeRate)

	msg := json.RawMessage(`{"set_exchange_rate": {"denom":"a"}}`)

	// Call setExchangeRate
	err = app.ExecuteRawCustom(t, ctx, junoApp, oracleQuerier, actor, msg, sdk.Coin{})
	require.NoError(t, err)

	// Query Chain
	queryParams := wasm.ExchangeRateQueryParams{
		Denom: "a",
	}
	bz, err := json.Marshal(bindings.CosmosQuery{
		Oracle: &wasm.OracleQuery{
			ExchangeRate: &queryParams,
		},
	})
	require.NoError(t, err)

	res, err := querier(ctx, bz)
	require.NoError(t, err)

	var exchangeRatesResponse wasm.ExchangeRateQueryResponse
	err = json.Unmarshal(res, &exchangeRatesResponse)
	require.NoError(t, err)

	exchangeRate, err := sdk.NewDecFromStr(exchangeRatesResponse.Rate)
	require.NoError(t, err)

	// Query contract
	query := wasm.OracleQuery{
		ExchangeRate: &wasm.ExchangeRateQueryParams{
			Denom: "a",
		},
	}
	queryBz, err := json.Marshal(query)
	require.NoError(t, err)

	resBz, err := junoApp.GetWasmKeeper().QuerySmart(ctx, oracleQuerier, queryBz)
	require.NoError(t, err)
	var rate string

	err = json.Unmarshal(resBz, &rate)
	require.NoError(t, err)

	// convert to sdk.Dec to match precision
	oracleRate, err := sdk.NewDecFromStr(rate)
	require.NoError(t, err)

	require.Equal(t, oracleRate, exchangeRate)
}

func storeOracleQuerierCode(t *testing.T, ctx sdk.Context, junoApp *app.App, addr sdk.AccAddress, wasmCode []byte) {
	govKeeper := junoApp.GovKeeper

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
	contractKeeper := keeper.NewDefaultPermissionKeeper(junoApp.GetWasmKeeper())
	codeID := uint64(1)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}
