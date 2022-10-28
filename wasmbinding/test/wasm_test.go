package wasmbinding_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v11/x/oracle/wasm"

	"github.com/CosmosContracts/juno/v11/app"
	"github.com/CosmosContracts/juno/v11/wasmbinding"
	"github.com/CosmosContracts/juno/v11/wasmbinding/bindings"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestQueryExchangeRates(t *testing.T) {
	junoApp := app.Setup(t, false, 1)
	ctx := junoApp.BaseApp.NewContext(false, tmtypes.Header{Height: 1, ChainID: "kujira-1", Time: time.Now().UTC()})

	ExchangeRateB := sdk.NewDec(1700)
	ExchangeRateC := sdk.NewDecWithPrec(17, 1)
	ExchangeRateD := sdk.NewDecWithPrec(19, 1)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "a", sdk.NewDec(1))
	junoApp.OracleKeeper.SetExchangeRate(ctx, "b", ExchangeRateC)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "c", ExchangeRateB)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "d", ExchangeRateD)

	plugin := wasmbinding.NewQueryPlugin(junoApp.OracleKeeper)
	querier := wasmbinding.CustomQuerier(plugin)
	var err error

	// empty data will occur error
	_, err = querier(ctx, []byte{})
	require.Error(t, err)

	// not existing quote denom query
	queryParams := wasm.ExchangeRateQueryParams{
		Denom: "non-exist",
	}
	bz, err := json.Marshal(bindings.CosmosQuery{
		Oracle: &wasm.OracleQuery{
			ExchangeRate: &queryParams,
		},
	})
	require.NoError(t, err)

	res, err := querier(ctx, bz)
	require.Error(t, err)

	var exchangeRatesResponse wasm.ExchangeRateQueryResponse
	err = json.Unmarshal(res, &exchangeRatesResponse)
	require.Error(t, err)

	// not existing base denom query
	queryParams = wasm.ExchangeRateQueryParams{
		Denom: "c",
	}
	bz, err = json.Marshal(bindings.CosmosQuery{
		Oracle: &wasm.OracleQuery{
			ExchangeRate: &queryParams,
		},
	})
	require.NoError(t, err)

	res, err = querier(ctx, bz)
	require.NoError(t, err)

	queryParams = wasm.ExchangeRateQueryParams{
		Denom: "b",
	}
	bz, err = json.Marshal(bindings.CosmosQuery{
		Oracle: &wasm.OracleQuery{
			ExchangeRate: &queryParams,
		},
	})
	require.NoError(t, err)

	res, err = querier(ctx, bz)
	require.NoError(t, err)

	err = json.Unmarshal(res, &exchangeRatesResponse)
	require.NoError(t, err)
	require.Equal(t, exchangeRatesResponse, wasm.ExchangeRateQueryResponse{
		Rate: ExchangeRateB.String(),
	})
}

func TestSupply(t *testing.T) {
	junoApp := app.Setup(t, false, 1)
	ctx := junoApp.BaseApp.NewContext(false, tmtypes.Header{Height: 1, ChainID: "kujira-1", Time: time.Now().UTC()})

	plugin := wasmbinding.NewQueryPlugin(junoApp.OracleKeeper)
	querier := wasmbinding.CustomQuerier(plugin)

	var err error

	// empty data will occur error
	_, err = querier(ctx, []byte{})
	require.Error(t, err)

	queryParams := banktypes.QuerySupplyOfRequest{
		Denom: "a",
	}
	bz, err := json.Marshal(bindings.CosmosQuery{
		Bank: &bindings.BankQuery{
			Supply: &queryParams,
		},
	})
	require.NoError(t, err)
	var x banktypes.QuerySupplyOfResponse

	res, err := querier(ctx, bz)

	err = json.Unmarshal(res, &x)
	require.NoError(t, err)
}
