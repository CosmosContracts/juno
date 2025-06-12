package staking

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

// Handler combines all staking-related WebSocket handlers
type Handler struct {
	delegationsHandler *DelegationsHandler
	unbondingHandler   *UnbondingHandler
}

// NewHandler creates a new staking handler with legacy parameters (temporarily for compatibility)
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

// NewHandlerWithDeps creates a new staking handler with dependencies
func NewHandlerWithDeps(keeper KeeperInterface, deps *common.HandlerDependencies) *Handler {
	return &Handler{
		delegationsHandler: NewDelegationsHandler(keeper, deps),
		unbondingHandler:   NewUnbondingHandler(keeper, deps),
	}
}


// HandleDelegationsSubscription handles delegations subscription WebSocket connections
func (h *Handler) HandleDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	h.delegationsHandler.Handle(w, r)
}

// HandleDelegationSubscription handles delegation subscription WebSocket connections
func (h *Handler) HandleDelegationSubscription(w http.ResponseWriter, r *http.Request) {
	h.delegationsHandler.HandleDelegation(w, r)
}

// HandleUnbondingDelegationsSubscription handles unbonding delegations subscription WebSocket connections
func (h *Handler) HandleUnbondingDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	h.unbondingHandler.HandleUnbondingDelegations(w, r)
}

// HandleUnbondingDelegationSubscription handles unbonding delegation subscription WebSocket connections
func (h *Handler) HandleUnbondingDelegationSubscription(w http.ResponseWriter, r *http.Request) {
	h.unbondingHandler.HandleUnbondingDelegation(w, r)
}
