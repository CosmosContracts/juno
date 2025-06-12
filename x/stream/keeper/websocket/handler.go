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

	return &Handler{
		bankHandler:    bank.NewHandlerWithDeps(bankKeeper, deps),
		stakingHandler: staking.NewHandlerWithDeps(stakingKeeper, deps),
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
