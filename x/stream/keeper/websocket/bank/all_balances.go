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

// AllBalancesHandler handles all balances subscription WebSocket connections
type AllBalancesHandler struct {
	keeper         KeeperInterface
	config         *common.StreamConfig
	logger         common.Logger
	connManager    common.ConnectionManager
	registry       common.SubscriptionRegistry
	circuitBreaker common.CircuitBreaker
	appContext     context.Context
	upgrader       *websocket.Upgrader
}

// NewAllBalancesHandler creates a new all balances handler
func NewAllBalancesHandler(keeper KeeperInterface, config *common.StreamConfig, logger common.Logger, connManager common.ConnectionManager, registry common.SubscriptionRegistry, circuitBreaker common.CircuitBreaker, appContext context.Context, upgrader *websocket.Upgrader) *AllBalancesHandler {
	return &AllBalancesHandler{
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

// Handle handles all balances subscription WebSocket connections
func (h *AllBalancesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
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

	// Send initial balances
	balances := h.keeper.GetBankKeeper().GetAllBalances(queryCtx, addr)
	if err := sendWebSocketMessage(conn, map[string]any{"balances": balances}); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, address, "", "")
	sendCh := make(chan any, h.config.SubscriptionBufferSize)
	subscriber := h.registry.Subscribe(common.NewSubscriptionKeyAdapter(subKey), ctx, sendCh)
	defer h.registry.Unsubscribe(subscriber)

	handleWebSocketConnection(h.config, h.circuitBreaker, h.logger, conn, ctx, sendCh, connectionID, func() any {
		queryCtx, err := h.keeper.GetQueryContext()
		if err != nil {
			return map[string]string{"error": "service temporarily unavailable"}
		}
		return map[string]any{"balances": h.keeper.GetBankKeeper().GetAllBalances(queryCtx, addr)}
	})
}

// Common helper functions that can be shared
func getQueryContextOrSendError(keeper KeeperInterface, logger common.Logger, conn *websocket.Conn) (context.Context, bool) {
	queryCtx, err := keeper.GetQueryContext()
	if err != nil {
		logger.Error("failed to get query context", "error", err)
		errorMsg := map[string]string{"error": "service temporarily unavailable"}
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if sendErr := conn.WriteJSON(errorMsg); sendErr != nil {
			logger.Error("failed to send error message", "error", sendErr)
		}
		return nil, false
	}
	return queryCtx, true
}

func checkCircuitBreaker(config *common.StreamConfig, circuitBreaker common.CircuitBreaker, logger common.Logger, connManager common.ConnectionManager, conn *websocket.Conn, connectionID string) bool {
	if config.CircuitBreakerEnabled && circuitBreaker != nil {
		allowed, err := circuitBreaker.AllowRequest(connectionID)
		if !allowed {
			logger.Warn("circuit breaker blocked request", "connection_id", connectionID, "error", err)
			types.IncrementConnectionRejected("circuit_breaker")
			errorMsg := map[string]string{"error": "service temporarily unavailable"}
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			conn.WriteJSON(errorMsg)
			conn.Close()
			connManager.UnregisterConnection(connectionID)
			return false
		}
	}
	return true
}

func handleWebSocketConnection(config *common.StreamConfig, circuitBreaker common.CircuitBreaker, logger common.Logger, conn *websocket.Conn, ctx context.Context, sendCh <-chan any, connectionID string, queryFunc func() any) {
	common.HandleWebSocketConnection(conn, ctx, sendCh, connectionID, queryFunc, 
		func(conn *websocket.Conn, data any) error {
			return sendWebSocketMessage(conn, data)
		},
		func(connectionID string) {
			logger.Error("failed to send websocket message", "connection_id", connectionID)
			if config.CircuitBreakerEnabled && circuitBreaker != nil {
				circuitBreaker.RecordFailure(connectionID)
			}
		},
		func(connectionID string) {
			if config.CircuitBreakerEnabled && circuitBreaker != nil {
				circuitBreaker.RecordSuccess(connectionID)
			}
		})
}

func sendWebSocketMessage(conn *websocket.Conn, data any) error {
	return common.SendWebSocketMessage(conn, data, func() {
		types.IncrementMessagesSent()
	})
}