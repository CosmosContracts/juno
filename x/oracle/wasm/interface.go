package wasm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"github.com/CosmosContracts/juno/v12/x/oracle/keeper"
)

// ExchangeRateQueryParams query request params for exchange rates
type ExchangeRateQueryParams struct {
	Denom string `json:"denom"`
}

// OracleQuery custom query interface for oracle querier
type OracleQuery struct {
	ExchangeRate *ExchangeRateQueryParams `json:"exchange_rate,omitempty"`
}

// ExchangeRateQueryResponse - exchange rates query response item
type ExchangeRateQueryResponse struct {
	Rate string `json:"rate"`
}

// QueryCustom implements custom query interface
func Handle(keeper keeper.Keeper, ctx sdk.Context, q *OracleQuery) (any, error) {
	if q.ExchangeRate != nil {
		rate, err := keeper.GetExchangeRate(ctx, q.ExchangeRate.Denom)
		if err != nil {
			return nil, err
		}

		return ExchangeRateQueryResponse{
			Rate: rate.String(),
		}, nil
	}

	return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Oracle variant"}
}
