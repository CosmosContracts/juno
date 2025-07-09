package common

import (
	"context"
	"errors"
	"fmt"
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
		// Send error as close message
		_ = CloseWithMessage(conn, websocket.CloseTryAgainLater, "Service temporarily unavailable")
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
			if err := conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
				logger.Error("failed to set write deadline", "error", err)
			}
			if err := conn.WriteJSON(errorMsg); err != nil {
				logger.Error("failed to write error message", "error", err)
			}
			if err := conn.Close(); err != nil {
				logger.Error("failed to close connection", "error", err)
			}
			connManager.UnregisterConnection(connectionID)
			return false
		}
	}
	return true
}

// HandleWebSocketConnectionWithCallbacks is a wrapper around HandleWebSocketConnection with circuit breaker callbacks
func HandleWebSocketConnectionWithCallbacks(ctx context.Context, config *StreamConfig, circuitBreaker CircuitBreaker, logger Logger, conn *websocket.Conn, sendCh <-chan any, connectionID string, queryFunc func() any) {
	HandleWebSocketConnection(ctx, conn, sendCh, connectionID, queryFunc,
		SendWebSocketMessageWithMetrics,
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
	_ = h.HandleStandardConnectionWithProvider(params, h)
}

// HandleStandardConnectionWithProvider handles the standard WebSocket connection lifecycle with a custom QueryContextGetter
func (h *BaseHandler) HandleStandardConnectionWithProvider(params ConnectionParams, provider QueryContextGetter) (retErr error) {
	// Validate parameters
	if params.ValidationFunc != nil {
		if err := params.ValidationFunc(); err != nil {
			http.Error(params.Writer, err.Error(), http.StatusBadRequest)
			return err
		}
	}

	// Check connection limits
	if !h.ConnManager.CheckConnectionLimits(params.Writer, params.Request) {
		return errors.New("connection limit exceeded")
	}

	conn, err := h.Upgrader.Upgrade(params.Writer, params.Request, nil)
	if err != nil {
		h.Logger.Error("websocket upgrade failed", "error", err)
		return err
	}

	// Track if we need to close with an error
	var closeError *struct {
		code int
		msg  string
	}

	// Add panic recovery that sends close message immediately
	defer func() {
		if r := recover(); r != nil {
			h.Logger.Error("panic in WebSocket handler", "panic", r)
			// Send close message immediately
			closeMsg := fmt.Sprintf("Internal error: %v", r)
			if len(closeMsg) > 123 {
				closeMsg = closeMsg[:123]
			}
			h.Logger.Info("Sending close message for panic", "msg", closeMsg)
			// Try to send close message right away
			_ = conn.SetWriteDeadline(time.Now().Add(WriteWait))
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, closeMsg))
			// Set closeError for final cleanup
			closeError = &struct {
				code int
				msg  string
			}{
				code: websocket.CloseInternalServerErr,
				msg:  closeMsg,
			}
		}
	}()

	// Register connection
	remoteAddr := params.Request.RemoteAddr
	xForwardedFor := params.Request.Header.Get("X-Forwarded-For")
	connectionID := h.ConnManager.RegisterConnectionWithHeaders(remoteAddr, xForwardedFor)
	if connectionID == "" {
		if err := conn.Close(); err != nil {
			h.Logger.Error("failed to close connection", "error", err)
		}
		return nil
	}

	// Check circuit breaker if enabled
	if !CheckCircuitBreaker(h.Config, h.CircuitBreaker, h.Logger, h.ConnManager, conn, connectionID) {
		return nil
	}
	defer func() {
		h.ConnManager.UnregisterConnection(connectionID)
		// Clean up circuit breaker state for this connection
		if h.Config.CircuitBreakerEnabled && h.CircuitBreaker != nil {
			if cb, ok := h.CircuitBreaker.(*middleware.CircuitBreaker); ok {
				cb.CleanupConnection(connectionID)
			}
		}
		// Send close message if we have an error
		if closeError != nil {
			h.Logger.Info("Sending close message", "code", closeError.code, "msg", closeError.msg)
			if err := CloseWithMessage(conn, closeError.code, closeError.msg); err != nil {
				h.Logger.Error("failed to send close message", "error", err)
			}
		}
		if err := conn.Close(); err != nil {
			h.Logger.Error("failed to close websocket connection", "error", err, "connection_id", connectionID)
		}
	}()

	// Add subscription to this connection
	if !h.ConnManager.AddSubscription(connectionID) {
		return nil
	}
	defer h.ConnManager.RemoveSubscription(connectionID)

	ctx, cancel := context.WithCancel(h.AppContext)
	defer cancel()

	// Send initial data if provided
	if params.InitialDataFunc != nil {
		queryCtx, ok := GetQueryContextOrSendError(provider, h.Logger, conn)
		if !ok {
			return nil
		}

		// Wrap InitialDataFunc call to catch panics
		func() {
			defer func() {
				if r := recover(); r != nil {
					h.Logger.Error("panic in InitialDataFunc", "panic", r)
					// Send close message immediately for panic
					closeMsg := fmt.Sprintf("Initial data error: %v", r)
					if len(closeMsg) > 123 {
						closeMsg = closeMsg[:123]
					}
					_ = CloseWithMessage(conn, websocket.CloseInternalServerErr, closeMsg)
					// Set error for outer handler
					closeError = &struct {
						code int
						msg  string
					}{
						code: websocket.CloseInternalServerErr,
						msg:  closeMsg,
					}
				}
			}()

			initialData, err := params.InitialDataFunc(queryCtx)
			if err != nil {
				h.Logger.Error("failed to get initial data", "error", err)
				// Set error for close message
				closeError = &struct {
					code int
					msg  string
				}{
					code: websocket.CloseUnsupportedData,
					msg:  fmt.Sprintf("Failed to retrieve initial data: %v", err),
				}
				return
			}

			if err := SendWebSocketMessageWithMetrics(conn, initialData); err != nil {
				return
			}
		}()

		// If we set a closeError, return
		if closeError != nil {
			return nil
		}
	}

	// Create subscription
	sendCh := make(chan any, h.Config.SubscriptionBufferSize)
	subscriber := h.Registry.Subscribe(ctx, params.SubscriptionKey, sendCh)
	defer h.Registry.Unsubscribe(subscriber)

	// Handle WebSocket connection
	HandleWebSocketConnectionWithCallbacks(ctx, h.Config, h.CircuitBreaker, h.Logger, conn, sendCh, connectionID, params.QueryFunc)
	return nil
}

// GetQueryContext implements QueryContextGetter for BaseHandler
// This is a placeholder - each module's handler should embed BaseHandler and implement their own
func (*BaseHandler) GetQueryContext() (context.Context, error) {
	panic("GetQueryContext must be implemented by the embedding struct")
}
