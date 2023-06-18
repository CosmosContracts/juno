// this file used from osmosis, ref: https://github.com/CosmosContracts/juno/v16/blob/2ce971f4c6aa85d3ef7ba33d60e0ae74b923ab83/app/keepers/querier.go
// Original Author: https://github.com/nicolaslara

package keepers

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QuerierWrapper is a local wrapper around BaseApp that exports only the Queryable interface.
// This is used to pass the baseApp to Async ICQ without exposing all methods
type QuerierWrapper struct {
	querier sdk.Queryable
}

var _ sdk.Queryable = QuerierWrapper{}

func NewQuerierWrapper(querier sdk.Queryable) QuerierWrapper {
	return QuerierWrapper{querier: querier}
}

func (q QuerierWrapper) Query(req abci.RequestQuery) abci.ResponseQuery {
	return q.querier.Query(req)
}
