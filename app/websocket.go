package app

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/server/api"
)

// RegisterWebSocketRoutes registers WebSocket routes for the stream module
func (app *App) RegisterWebSocketRoutes(apiSvr *api.Server) error {
	// Get the stream keeper
	streamKeeper := app.AppKeepers.StreamKeeper

	// Register WebSocket endpoints
	// Bank endpoints
	apiSvr.Router.Handle("/ws/subscribe/bank/balance/{address}/{denom}", http.HandlerFunc(streamKeeper.HandleBalanceSubscription))
	apiSvr.Router.Handle("/ws/subscribe/bank/balances/{address}", http.HandlerFunc(streamKeeper.HandleAllBalancesSubscription))

	// Staking endpoints
	apiSvr.Router.Handle("/ws/subscribe/staking/delegations/{delegator}", http.HandlerFunc(streamKeeper.HandleDelegationsSubscription))
	apiSvr.Router.Handle("/ws/subscribe/staking/delegation/{delegator}/{validator}", http.HandlerFunc(streamKeeper.HandleDelegationSubscription))
	apiSvr.Router.Handle("/ws/subscribe/staking/unbonding-delegations/{delegator}", http.HandlerFunc(streamKeeper.HandleUnbondingDelegationsSubscription))
	apiSvr.Router.Handle("/ws/subscribe/staking/unbonding-delegation/{delegator}/{validator}", http.HandlerFunc(streamKeeper.HandleUnbondingDelegationSubscription))

	return nil
}