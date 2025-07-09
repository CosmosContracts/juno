package common

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// QueryContextMiddleware wraps query functions with error handling
type QueryContextMiddleware struct {
	provider QueryContextProvider
	logger   Logger
}

// NewQueryContextMiddleware creates a new query context middleware
func NewQueryContextMiddleware(provider QueryContextProvider, logger Logger) *QueryContextMiddleware {
	return &QueryContextMiddleware{
		provider: provider,
		logger:   logger,
	}
}

// WithQueryContext wraps a function that needs query context with error handling
func (m *QueryContextMiddleware) WithQueryContext(fn func(context.Context) any) func() any {
	return func() any {
		ctx, err := m.provider.GetQueryContext()
		if err != nil {
			m.logger.Error("failed to get query context in wrapped query function", "error", err)
			return ErrorResponse("service temporarily unavailable")
		}
		m.logger.Debug("query context obtained successfully for websocket update")
		return fn(ctx)
	}
}

// WithQueryContextAndError wraps a function that returns data and error
func (m *QueryContextMiddleware) WithQueryContextAndError(fn func(context.Context) (any, error)) func() any {
	return func() any {
		ctx, err := m.provider.GetQueryContext()
		if err != nil {
			m.logger.Error("failed to get query context", "error", err)
			return ErrorResponse("service temporarily unavailable")
		}

		data, err := fn(ctx)
		if err != nil {
			m.logger.Error("query function error", "error", err)
			return ErrorResponse("failed to retrieve data")
		}
		return data
	}
}

// SendErrorResponse sends an error response over websocket
func SendErrorResponse(conn *websocket.Conn, message string) error {
	if err := conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
		return err
	}
	return conn.WriteJSON(ErrorResponse(message))
}

// QueryHandler is a generic handler that manages query context
type QueryHandler struct {
	BaseHandler
	provider QueryContextProvider
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(deps *HandlerDependencies, provider QueryContextProvider) QueryHandler {
	return QueryHandler{
		BaseHandler: deps.CreateBaseHandler(),
		provider:    provider,
	}
}

// GetQueryContext implements QueryContextGetter interface
func (h *QueryHandler) GetQueryContext() (context.Context, error) {
	return h.provider.GetQueryContext()
}

// WrapQueryFunc wraps a query function with context error handling
func (h *QueryHandler) WrapQueryFunc(fn func(context.Context) any) func() any {
	middleware := NewQueryContextMiddleware(h.provider, h.Logger)
	return middleware.WithQueryContext(fn)
}

// WrapInitialDataFunc wraps an initial data function with context
func (h *QueryHandler) WrapInitialDataFunc(fn func(context.Context) (any, error)) func(context.Context) (any, error) {
	return func(ctx context.Context) (any, error) {
		// Use the provided context first, fall back to getting a new one if needed
		if ctx == nil {
			newCtx, err := h.provider.GetQueryContext()
			if err != nil {
				return nil, err
			}
			ctx = newCtx
		}
		return fn(ctx)
	}
}

// HandleStandardConnection overrides BaseHandler to provide proper QueryContextGetter
func (h *QueryHandler) HandleStandardConnection(params ConnectionParams) {
	_ = h.HandleStandardConnectionWithProvider(params, h)
}
