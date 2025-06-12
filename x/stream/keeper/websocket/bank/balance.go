package bank

import (
	"context"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

// BalanceHandler handles balance subscription WebSocket connections
type BalanceHandler struct {
	keeper          KeeperInterface
	config          *common.StreamConfig
	logger          common.Logger
	connManager     common.ConnectionManager
	registry        common.SubscriptionRegistry
	circuitBreaker  common.CircuitBreaker
	appContext      context.Context
	upgrader        *websocket.Upgrader
}

// NewBalanceHandler creates a new balance handler
func NewBalanceHandler(keeper KeeperInterface, config *common.StreamConfig, logger common.Logger, connManager common.ConnectionManager, registry common.SubscriptionRegistry, circuitBreaker common.CircuitBreaker, appContext context.Context, upgrader *websocket.Upgrader) *BalanceHandler {
	return &BalanceHandler{
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

// Handle handles balance subscription WebSocket connections
func (h *BalanceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	denom := vars["denom"]

	// Validate address
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	// Validate denom
	if err := h.keeper.ValidateDenom(r.Context(), denom); err != nil {
		http.Error(w, "invalid denom: "+err.Error(), http.StatusBadRequest)
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
	if !h.checkCircuitBreaker(conn, connectionID) {
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
	queryCtx, ok := h.getQueryContextOrSendError(conn)
	if !ok {
		return
	}

	// Send initial balance
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		h.logger.Error("failed to parse address", "error", err, "address", address)
		errorMsg := map[string]string{"error": "invalid address format"}
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if sendErr := conn.WriteJSON(errorMsg); sendErr != nil {
			h.logger.Error("failed to send error message", "error", sendErr)
		}
		return
	}
	balance := h.keeper.GetBankKeeper().GetBalance(queryCtx, addr, denom)
	if err := h.sendWebSocketMessage(conn, balance); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, address, "", denom)
	sendCh := make(chan any, h.config.SubscriptionBufferSize)
	subscriber := h.registry.Subscribe(common.NewSubscriptionKeyAdapter(subKey), ctx, sendCh)
	defer h.registry.Unsubscribe(subscriber)

	h.handleWebSocketConnection(conn, ctx, sendCh, connectionID, func() any {
		queryCtx, err := h.keeper.GetQueryContext()
		if err != nil {
			return map[string]string{"error": "service temporarily unavailable"}
		}
		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return map[string]string{"error": "invalid address format"}
		}
		return h.keeper.GetBankKeeper().GetBalance(queryCtx, addr, denom)
	})
}

// getQueryContextOrSendError gets the query context or sends an error message and returns false
func (h *BalanceHandler) getQueryContextOrSendError(conn *websocket.Conn) (context.Context, bool) {
	queryCtx, err := h.keeper.GetQueryContext()
	if err != nil {
		h.logger.Error("failed to get query context", "error", err)
		errorMsg := map[string]string{"error": "service temporarily unavailable"}
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if sendErr := conn.WriteJSON(errorMsg); sendErr != nil {
			h.logger.Error("failed to send error message", "error", sendErr)
		}
		return nil, false
	}
	return queryCtx, true
}

// checkCircuitBreaker checks if the circuit breaker allows the request
func (h *BalanceHandler) checkCircuitBreaker(conn *websocket.Conn, connectionID string) bool {
	if h.config.CircuitBreakerEnabled && h.circuitBreaker != nil {
		allowed, err := h.circuitBreaker.AllowRequest(connectionID)
		if !allowed {
			h.logger.Warn("circuit breaker blocked request", "connection_id", connectionID, "error", err)
			types.IncrementConnectionRejected("circuit_breaker")
			errorMsg := map[string]string{"error": "service temporarily unavailable"}
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			conn.WriteJSON(errorMsg)
			conn.Close()
			h.connManager.UnregisterConnection(connectionID)
			return false
		}
	}
	return true
}

// handleWebSocketConnection handles the WebSocket connection lifecycle
func (h *BalanceHandler) handleWebSocketConnection(conn *websocket.Conn, ctx context.Context, sendCh <-chan any, connectionID string, queryFunc func() any) {
	common.HandleWebSocketConnection(conn, ctx, sendCh, connectionID, queryFunc, h.sendWebSocketMessage,
		func(connectionID string) {
			h.logger.Error("failed to send websocket message", "connection_id", connectionID)
			if h.config.CircuitBreakerEnabled && h.circuitBreaker != nil {
				h.circuitBreaker.RecordFailure(connectionID)
			}
		},
		func(connectionID string) {
			if h.config.CircuitBreakerEnabled && h.circuitBreaker != nil {
				h.circuitBreaker.RecordSuccess(connectionID)
			}
		})
}

// sendWebSocketMessage sends a message over WebSocket
func (h *BalanceHandler) sendWebSocketMessage(conn *websocket.Conn, data any) error {
	return common.SendWebSocketMessage(conn, data, func() {
		types.IncrementMessagesSent()
	})
}