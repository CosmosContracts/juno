package staking

import (
	"context"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

// UnbondingHandler handles unbonding delegations subscription WebSocket connections
type UnbondingHandler struct {
	keeper         KeeperInterface
	config         *common.StreamConfig
	logger         common.Logger
	connManager    common.ConnectionManager
	registry       common.SubscriptionRegistry
	circuitBreaker common.CircuitBreaker
	appContext     context.Context
	upgrader       *websocket.Upgrader
}

// NewUnbondingHandler creates a new unbonding handler
func NewUnbondingHandler(keeper KeeperInterface, config *common.StreamConfig, logger common.Logger, connManager common.ConnectionManager, registry common.SubscriptionRegistry, circuitBreaker common.CircuitBreaker, appContext context.Context, upgrader *websocket.Upgrader) *UnbondingHandler {
	return &UnbondingHandler{
		keeper:         keeper,
		config:         config,
		logger:         logger,
		connManager:    connManager,
		registry:       registry,
		circuitBreaker: circuitBreaker,
		appContext:     appContext,
		upgrader:       upgrader,
	}
}

// HandleUnbondingDelegations handles unbonding delegations subscription WebSocket connections
func (h *UnbondingHandler) HandleUnbondingDelegations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !h.connManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	connectionID := h.connManager.RegisterConnectionWithHeaders(remoteAddr, xForwardedFor)
	if connectionID == "" {
		conn.Close()
		return
	}

	// Check circuit breaker if enabled
	if !checkCircuitBreaker(h.config, h.circuitBreaker, h.logger, h.connManager, conn, connectionID) {
		return
	}
	defer func() {
		h.connManager.UnregisterConnection(connectionID)
		if err := conn.Close(); err != nil {
			h.logger.Error("failed to close websocket connection", "error", err, "connection_id", connectionID)
		}
	}()

	// Add subscription to this connection
	if !h.connManager.AddSubscription(connectionID) {
		return
	}
	defer h.connManager.RemoveSubscription(connectionID)

	ctx, cancel := context.WithCancel(h.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := getQueryContextOrSendError(h.keeper, h.logger, conn)
	if !ok {
		return
	}

	// Send initial unbonding delegations
	unbondingDelegations, err := h.keeper.GetStakingKeeper().GetAllUnbondingDelegations(queryCtx, delAddr)
	if err != nil {
		h.logger.Error("failed to get unbonding delegations", "error", err)
		unbondingDelegations = []stakingtypes.UnbondingDelegation{}
	}
	if err := sendWebSocketMessage(conn, unbondingDelegations); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, delegatorAddress, "", "")
	sendCh := make(chan any, h.config.SubscriptionBufferSize)
	subscriber := h.registry.Subscribe(common.NewSubscriptionKeyAdapter(subKey), ctx, sendCh)
	defer h.registry.Unsubscribe(subscriber)

	handleWebSocketConnection(h.config, h.circuitBreaker, h.logger, conn, ctx, sendCh, connectionID, func() any {
		queryCtx, err := h.keeper.GetQueryContext()
		if err != nil {
			return []stakingtypes.UnbondingDelegation{}
		}
		unbondingDelegations, err := h.keeper.GetStakingKeeper().GetAllUnbondingDelegations(queryCtx, delAddr)
		if err != nil {
			return []stakingtypes.UnbondingDelegation{}
		}
		return unbondingDelegations
	})
}

// HandleUnbondingDelegation handles unbonding delegation subscription WebSocket connections
func (h *UnbondingHandler) HandleUnbondingDelegation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]
	validatorAddress := vars["validator"]

	// Validate addresses
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}
	valAddr, err := sdk.ValAddressFromBech32(validatorAddress)
	if err != nil {
		http.Error(w, "invalid validator address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !h.connManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	connectionID := h.connManager.RegisterConnectionWithHeaders(remoteAddr, xForwardedFor)
	if connectionID == "" {
		conn.Close()
		return
	}

	// Check circuit breaker if enabled
	if !checkCircuitBreaker(h.config, h.circuitBreaker, h.logger, h.connManager, conn, connectionID) {
		return
	}
	defer func() {
		h.connManager.UnregisterConnection(connectionID)
		if err := conn.Close(); err != nil {
			h.logger.Error("failed to close websocket connection", "error", err, "connection_id", connectionID)
		}
	}()

	// Add subscription to this connection
	if !h.connManager.AddSubscription(connectionID) {
		return
	}
	defer h.connManager.RemoveSubscription(connectionID)

	ctx, cancel := context.WithCancel(h.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := getQueryContextOrSendError(h.keeper, h.logger, conn)
	if !ok {
		return
	}

	// Send initial unbonding delegation
	unbondingDelegation, err := h.keeper.GetStakingKeeper().GetUnbondingDelegation(queryCtx, delAddr, valAddr)
	data := map[string]any{"found": err == nil}
	if err == nil {
		data["unbonding_delegation"] = unbondingDelegation
	}
	if err := sendWebSocketMessage(conn, data); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, delegatorAddress, validatorAddress, "")
	sendCh := make(chan any, h.config.SubscriptionBufferSize)
	subscriber := h.registry.Subscribe(common.NewSubscriptionKeyAdapter(subKey), ctx, sendCh)
	defer h.registry.Unsubscribe(subscriber)

	handleWebSocketConnection(h.config, h.circuitBreaker, h.logger, conn, ctx, sendCh, connectionID, func() any {
		queryCtx, err := h.keeper.GetQueryContext()
		if err != nil {
			return map[string]any{"found": false}
		}
		unbondingDelegation, err := h.keeper.GetStakingKeeper().GetUnbondingDelegation(queryCtx, delAddr, valAddr)
		data := map[string]any{"found": err == nil}
		if err == nil {
			data["unbonding_delegation"] = unbondingDelegation
		}
		return data
	})
}