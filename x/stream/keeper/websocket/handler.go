package websocket

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/bank"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/staking"
)

// Handler combines all WebSocket handlers
type Handler struct {
	bankHandler    *bank.Handler
	stakingHandler *staking.Handler
	registry       *common.ModuleRegistry
	config         *common.StreamConfig
	logger         common.Logger
	upgrader       *websocket.Upgrader
}

// NewHandler creates a new WebSocket handler
func NewHandler(
	ctx context.Context,
	bankKeeper bank.KeeperInterface,
	stakingKeeper staking.KeeperInterface,
	config *common.StreamConfig,
	logger common.Logger,
	connManager common.ConnectionManager,
	registry common.SubscriptionRegistry,
	circuitBreaker common.CircuitBreaker,
	allowAllOrigins bool,
) *Handler {
	deps := common.NewHandlerDependencies(ctx, config, logger, connManager, registry, circuitBreaker, allowAllOrigins)

	// Create module registry
	moduleRegistry := common.NewModuleRegistry()

	// Register modules
	bankModule := bank.NewModule(bankKeeper, deps)
	stakingModule := staking.NewModule(stakingKeeper, deps)

	if err := moduleRegistry.RegisterModule(bankModule); err != nil {
		logger.Error("failed to register bank module", "error", err)
	}
	if err := moduleRegistry.RegisterModule(stakingModule); err != nil {
		logger.Error("failed to register staking module", "error", err)
	}

	return &Handler{
		bankHandler:    bank.NewHandlerWithDeps(bankKeeper, deps),
		stakingHandler: staking.NewHandlerWithDeps(stakingKeeper, deps),
		registry:       moduleRegistry,
		config:         config,
		logger:         logger,
		upgrader:       deps.Upgrader,
	}
}

// Bank handlers

func (h *Handler) HandleBalanceSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleBalanceSubscription(w, r)
}

func (h *Handler) HandleAllBalancesSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleAllBalancesSubscription(w, r)
}

func (h *Handler) HandleSpendableBalancesSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleSpendableBalancesSubscription(w, r)
}

func (h *Handler) HandleSpendableBalanceByDenomSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleSpendableBalanceByDenomSubscription(w, r)
}

func (h *Handler) HandleTotalSupplySubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleTotalSupplySubscription(w, r)
}

func (h *Handler) HandleSupplyOfSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleSupplyOfSubscription(w, r)
}

func (h *Handler) HandleParamsSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleParamsSubscription(w, r)
}

func (h *Handler) HandleDenomsMetadataSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleDenomsMetadataSubscription(w, r)
}

func (h *Handler) HandleDenomMetadataSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleDenomMetadataSubscription(w, r)
}

func (h *Handler) HandleDenomOwnersSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleDenomOwnersSubscription(w, r)
}

func (h *Handler) HandleSendEnabledSubscription(w http.ResponseWriter, r *http.Request) {
	h.bankHandler.HandleSendEnabledSubscription(w, r)
}

// Staking handlers

func (h *Handler) HandleDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	h.stakingHandler.HandleDelegationsSubscription(w, r)
}

func (h *Handler) HandleDelegationSubscription(w http.ResponseWriter, r *http.Request) {
	h.stakingHandler.HandleDelegationSubscription(w, r)
}

func (h *Handler) HandleUnbondingDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	h.stakingHandler.HandleUnbondingDelegationsSubscription(w, r)
}

func (h *Handler) HandleUnbondingDelegationSubscription(w http.ResponseWriter, r *http.Request) {
	h.stakingHandler.HandleUnbondingDelegationSubscription(w, r)
}

// GetModuleRegistry returns the module registry for dynamic module management
func (h *Handler) GetModuleRegistry() *common.ModuleRegistry {
	return h.registry
}

// ListModules returns all registered module names
func (h *Handler) ListModules() []string {
	if h.registry != nil {
		return h.registry.ListModules()
	}
	return []string{}
}

// ListRoutes returns all registered routes with their module information
func (h *Handler) ListRoutes() map[string]string {
	if h.registry != nil {
		return h.registry.ListRoutes()
	}
	return map[string]string{}
}

// RegisterRoutes registers all WebSocket routes with the given router
// The router should have a Handle method that accepts a pattern and handler
func (h *Handler) RegisterRoutes(registerFunc func(pattern string, handler http.Handler)) {
	if h.registry == nil {
		return
	}

	// Get all route definitions from the registry
	routeDefinitions := h.registry.GetAllRouteDefinitions()

	// Register each route with the router
	for _, routeDef := range routeDefinitions {
		registerFunc(routeDef.Pattern, http.HandlerFunc(routeDef.Handler))
	}
}
