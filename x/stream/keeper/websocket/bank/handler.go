package bank

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

// Handler combines all bank-related WebSocket handlers
type Handler struct {
	balanceHandler    *BalanceHandler
	allBalancesHandler *AllBalancesHandler
}

// NewHandler creates a new bank handler
func NewHandler(keeper KeeperInterface, config *common.StreamConfig, logger common.Logger, connManager common.ConnectionManager, registry common.SubscriptionRegistry, circuitBreaker common.CircuitBreaker, appContext context.Context, upgrader *websocket.Upgrader) *Handler {
	return &Handler{
		balanceHandler:     NewBalanceHandler(keeper, config, logger, connManager, registry, circuitBreaker, appContext, upgrader),
		allBalancesHandler: NewAllBalancesHandler(keeper, config, logger, connManager, registry, circuitBreaker, appContext, upgrader),
	}
}

// HandleBalanceSubscription handles balance subscription WebSocket connections
func (h *Handler) HandleBalanceSubscription(w http.ResponseWriter, r *http.Request) {
	h.balanceHandler.Handle(w, r)
}

// HandleAllBalancesSubscription handles all balances subscription WebSocket connections
func (h *Handler) HandleAllBalancesSubscription(w http.ResponseWriter, r *http.Request) {
	h.allBalancesHandler.Handle(w, r)
}