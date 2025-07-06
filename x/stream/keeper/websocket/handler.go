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
	bankKeeper bank.KeeperInterface,
	stakingKeeper staking.KeeperInterface,
	config *common.StreamConfig,
	logger common.Logger,
	connManager common.ConnectionManager,
	registry common.SubscriptionRegistry,
	circuitBreaker common.CircuitBreaker,
	appContext context.Context,
	allowAllOrigins bool,
) *Handler {
	deps := common.NewHandlerDependencies(config, logger, connManager, registry, circuitBreaker, appContext, allowAllOrigins)

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
