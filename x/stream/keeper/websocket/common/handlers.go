package common

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/middleware"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// BaseHandler contains common fields for all WebSocket handlers
type BaseHandler struct {
	Config         *StreamConfig
	Logger         Logger
	ConnManager    ConnectionManager
	Registry       SubscriptionRegistry
	CircuitBreaker CircuitBreaker
	AppContext     context.Context
	Upgrader       *websocket.Upgrader
}

// QueryContextGetter defines the interface for getting query context
type QueryContextGetter interface {
	GetQueryContext() (context.Context, error)
}

// GetQueryContextOrSendError gets the query context or sends an error message and returns false
func GetQueryContextOrSendError(getter QueryContextGetter, logger Logger, conn *websocket.Conn) (context.Context, bool) {
	queryCtx, err := getter.GetQueryContext()
	if err != nil {
		logger.Error("failed to get query context", "error", err)
		errorMsg := map[string]string{"error": "service temporarily unavailable"}
		conn.SetWriteDeadline(time.Now().Add(WriteWait))
		if sendErr := conn.WriteJSON(errorMsg); sendErr != nil {
			logger.Error("failed to send error message", "error", sendErr)
		}
		return nil, false
	}
	return queryCtx, true
}

// CheckCircuitBreaker checks if the circuit breaker allows the request
func CheckCircuitBreaker(config *StreamConfig, circuitBreaker CircuitBreaker, logger Logger, connManager ConnectionManager, conn *websocket.Conn, connectionID string) bool {
	if config.CircuitBreakerEnabled && circuitBreaker != nil {
		allowed, err := circuitBreaker.AllowRequest(connectionID)
		if !allowed {
			logger.Warn("circuit breaker blocked request", "connection_id", connectionID, "error", err)
			types.IncrementConnectionRejected("circuit_breaker")
			errorMsg := map[string]string{"error": "service temporarily unavailable"}
			conn.SetWriteDeadline(time.Now().Add(WriteWait))
			conn.WriteJSON(errorMsg)
			conn.Close()
			connManager.UnregisterConnection(connectionID)
			return false
		}
	}
	return true
}

// HandleWebSocketConnectionWithCallbacks is a wrapper around HandleWebSocketConnection with circuit breaker callbacks
func HandleWebSocketConnectionWithCallbacks(config *StreamConfig, circuitBreaker CircuitBreaker, logger Logger, conn *websocket.Conn, ctx context.Context, sendCh <-chan any, connectionID string, queryFunc func() any) {
	HandleWebSocketConnection(conn, ctx, sendCh, connectionID, queryFunc,
		func(conn *websocket.Conn, data any) error {
			return SendWebSocketMessageWithMetrics(conn, data)
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

// SendWebSocketMessageWithMetrics sends a message over WebSocket and increments metrics
func SendWebSocketMessageWithMetrics(conn *websocket.Conn, data any) error {
	return SendWebSocketMessage(conn, data, func() {
		types.IncrementMessagesSent()
	})
}

// ConnectionParams contains parameters for establishing a WebSocket connection
type ConnectionParams struct {
	Writer          http.ResponseWriter
	Request         *http.Request
	ConnectionID    string
	ValidationFunc  func() error
	InitialDataFunc func(context.Context) (any, error)
	SubscriptionKey SubscriptionKey
	QueryFunc       func() any
}

// HandleStandardConnection handles the standard WebSocket connection lifecycle
func (h *BaseHandler) HandleStandardConnection(params ConnectionParams) {
	h.HandleStandardConnectionWithProvider(params, h)
}

// HandleStandardConnectionWithProvider handles the standard WebSocket connection lifecycle with a custom QueryContextGetter
func (h *BaseHandler) HandleStandardConnectionWithProvider(params ConnectionParams, provider QueryContextGetter) {
	// Validate parameters
	if params.ValidationFunc != nil {
		if err := params.ValidationFunc(); err != nil {
			http.Error(params.Writer, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Check connection limits
	if !h.ConnManager.CheckConnectionLimits(params.Writer, params.Request) {
		return
	}

	conn, err := h.Upgrader.Upgrade(params.Writer, params.Request, nil)
	if err != nil {
		h.Logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := params.Request.RemoteAddr
	xForwardedFor := params.Request.Header.Get("X-Forwarded-For")
	connectionID := h.ConnManager.RegisterConnectionWithHeaders(remoteAddr, xForwardedFor)
	if connectionID == "" {
		conn.Close()
		return
	}

	// Check circuit breaker if enabled
	if !CheckCircuitBreaker(h.Config, h.CircuitBreaker, h.Logger, h.ConnManager, conn, connectionID) {
		return
	}
	defer func() {
		h.ConnManager.UnregisterConnection(connectionID)
		// Clean up circuit breaker state for this connection
		if h.Config.CircuitBreakerEnabled && h.CircuitBreaker != nil {
			if cb, ok := h.CircuitBreaker.(*middleware.CircuitBreaker); ok {
				cb.CleanupConnection(connectionID)
			}
		}
		if err := conn.Close(); err != nil {
			h.Logger.Error("failed to close websocket connection", "error", err, "connection_id", connectionID)
		}
	}()

	// Add subscription to this connection
	if !h.ConnManager.AddSubscription(connectionID) {
		return
	}
	defer h.ConnManager.RemoveSubscription(connectionID)

	ctx, cancel := context.WithCancel(h.AppContext)
	defer cancel()

	// Send initial data if provided
	if params.InitialDataFunc != nil {
		queryCtx, ok := GetQueryContextOrSendError(provider, h.Logger, conn)
		if !ok {
			return
		}

		initialData, err := params.InitialDataFunc(queryCtx)
		if err != nil {
			h.Logger.Error("failed to get initial data", "error", err)
			errorMsg := map[string]string{"error": "failed to retrieve initial data"}
			conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if sendErr := conn.WriteJSON(errorMsg); sendErr != nil {
				h.Logger.Error("failed to send error message", "error", sendErr)
			}
			return
		}

		if err := SendWebSocketMessageWithMetrics(conn, initialData); err != nil {
			return
		}
	}

	// Create subscription
	sendCh := make(chan any, h.Config.SubscriptionBufferSize)
	subscriber := h.Registry.Subscribe(params.SubscriptionKey, ctx, sendCh)
	defer h.Registry.Unsubscribe(subscriber)

	// Handle WebSocket connection
	HandleWebSocketConnectionWithCallbacks(h.Config, h.CircuitBreaker, h.Logger, conn, ctx, sendCh, connectionID, params.QueryFunc)
}

// GetQueryContext implements QueryContextGetter for BaseHandler
// This is a placeholder - each module's handler should embed BaseHandler and implement their own
func (h *BaseHandler) GetQueryContext() (context.Context, error) {
	panic("GetQueryContext must be implemented by the embedding struct")
}
