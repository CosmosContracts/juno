package wasmbinding_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v12/x/oracle/wasm"

	"github.com/CosmosContracts/juno/v12/app"
	"github.com/CosmosContracts/juno/v12/wasmbinding"
	"github.com/CosmosContracts/juno/v12/wasmbinding/bindings"

	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestQueryExchangeRates(t *testing.T) {
	junoApp := app.Setup(t, false, 1)
	ctx := junoApp.BaseApp.NewContext(false, tmtypes.Header{Height: 1, ChainID: "kujira-1", Time: time.Now().UTC()})

	ExchangeRateC := sdk.NewDec(1700)
	ExchangeRateB := sdk.NewDecWithPrec(17, 1)
	ExchangeRateD := sdk.NewDecWithPrec(19, 1)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "a", sdk.NewDec(1))
	junoApp.OracleKeeper.SetExchangeRate(ctx, "b", ExchangeRateB)
	junoApp.OracleKeeper.SetExchangeRate(ctx, "c", ExchangeRateC)
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

	_, err = querier(ctx, bz)
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
