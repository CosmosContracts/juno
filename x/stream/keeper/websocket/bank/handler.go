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
func NewHandler(ctx context.Context, keeper KeeperInterface, config *common.StreamConfig, logger common.Logger, connManager common.ConnectionManager, registry common.SubscriptionRegistry, circuitBreaker common.CircuitBreaker, upgrader *websocket.Upgrader) *Handler {
	deps := &common.HandlerDependencies{
		Config:         config,
		Logger:         logger,
		ConnManager:    connManager,
		Registry:       registry,
		CircuitBreaker: circuitBreaker,
		AppContext:     ctx,
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

// HandleSpendableBalancesSubscription handles spendable balances subscription WebSocket connections
func (h *Handler) HandleSpendableBalancesSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleSpendableBalances(w, r)
}

// HandleSpendableBalanceByDenomSubscription handles spendable balance by denom subscription WebSocket connections
func (h *Handler) HandleSpendableBalanceByDenomSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleSpendableBalanceByDenom(w, r)
}

// HandleTotalSupplySubscription handles total supply subscription WebSocket connections
func (h *Handler) HandleTotalSupplySubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleTotalSupply(w, r)
}

// HandleSupplyOfSubscription handles supply of subscription WebSocket connections
func (h *Handler) HandleSupplyOfSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleSupplyOf(w, r)
}

// HandleParamsSubscription handles params subscription WebSocket connections
func (h *Handler) HandleParamsSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleParams(w, r)
}

// HandleDenomsMetadataSubscription handles denoms metadata subscription WebSocket connections
func (h *Handler) HandleDenomsMetadataSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleDenomsMetadata(w, r)
}

// HandleDenomMetadataSubscription handles denom metadata subscription WebSocket connections
func (h *Handler) HandleDenomMetadataSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleDenomMetadata(w, r)
}

// HandleDenomOwnersSubscription handles denom owners subscription WebSocket connections
func (h *Handler) HandleDenomOwnersSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleDenomOwners(w, r)
}

// HandleSendEnabledSubscription handles send enabled subscription WebSocket connections
func (h *Handler) HandleSendEnabledSubscription(w http.ResponseWriter, r *http.Request) {
	h.module.handleSendEnabled(w, r)
}
