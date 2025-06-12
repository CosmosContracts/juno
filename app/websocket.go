package app

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/server/api"
)

// RegisterWebSocketRoutes registers WebSocket routes for the stream module
func (app *App) RegisterWebSocketRoutes(apiSvr *api.Server) error {
	// Get the websocket handler from stream keeper
	wsHandler := app.AppKeepers.StreamKeeper.WebSocketHandler()
	if wsHandler == nil {
		return fmt.Errorf("websocket handler not initialized")
	}

	// Register WebSocket endpoints
	// Bank endpoints
	apiSvr.Router.Handle("/ws/subscribe/bank/balance/{address}/{denom}", http.HandlerFunc(wsHandler.HandleBalanceSubscription))
	apiSvr.Router.Handle("/ws/subscribe/bank/balances/{address}", http.HandlerFunc(wsHandler.HandleAllBalancesSubscription))

	// Staking endpoints
	apiSvr.Router.Handle("/ws/subscribe/staking/delegations/{delegator}", http.HandlerFunc(wsHandler.HandleDelegationsSubscription))
	apiSvr.Router.Handle("/ws/subscribe/staking/delegation/{delegator}/{validator}", http.HandlerFunc(wsHandler.HandleDelegationSubscription))
	apiSvr.Router.Handle("/ws/subscribe/staking/unbonding-delegations/{delegator}", http.HandlerFunc(wsHandler.HandleUnbondingDelegationsSubscription))
	apiSvr.Router.Handle("/ws/subscribe/staking/unbonding-delegation/{delegator}/{validator}", http.HandlerFunc(wsHandler.HandleUnbondingDelegationSubscription))

	return nil
}