package bank

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

// Handler combines all bank-related WebSocket handlers
type Handler struct {
	module *Module
}

// NewHandler creates a new bank handler with legacy parameters (temporarily for compatibility)
func NewHandler(keeper KeeperInterface, config *common.StreamConfig, logger common.Logger, connManager common.ConnectionManager, registry common.SubscriptionRegistry, circuitBreaker common.CircuitBreaker, appContext context.Context, upgrader *websocket.Upgrader) *Handler {
	deps := &common.HandlerDependencies{
		Config:         config,
		Logger:         logger,
		ConnManager:    connManager,
		Registry:       registry,
		CircuitBreaker: circuitBreaker,
		AppContext:     appContext,
		Upgrader:       upgrader,
	}
	return NewHandlerWithDeps(keeper, deps)
}

// NewHandlerWithDeps creates a new bank handler with dependencies
func NewHandlerWithDeps(keeper KeeperInterface, deps *common.HandlerDependencies) *Handler {
	return &Handler{
		module: NewModule(keeper, deps),
	}
}


// HandleBalanceSubscription handles balance subscription WebSocket connections
func (h *Handler) HandleBalanceSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleBalance(w, r)
}

// HandleAllBalancesSubscription handles all balances subscription WebSocket connections
func (h *Handler) HandleAllBalancesSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleAllBalances(w, r)
}
