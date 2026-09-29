package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"

	"cosmossdk.io/log"

	"github.com/CosmosContracts/juno/v31/app/endpoints/websocket/common"
	"github.com/CosmosContracts/juno/v31/x/stream/types"
	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

// ServerOptions contains the configuration required to build a WebSocket server.
type ServerOptions struct {
	ContextDone    <-chan struct{}
	MethodRegistry *encoding.DynamicRegistry
	Registry       *types.SubscriptionRegistry
	Invoker        *types.RouterInvoker
	Logger         log.Logger
	Upgrader       *websocket.Upgrader
}

// Server wraps the WebSocket handler together with its supporting components.
type Server struct {
	routes   []common.RouteDefinition
	registry *types.SubscriptionRegistry
	logger   log.Logger
}

func buildRoutes(
	handler *common.Handler,
	logger log.Logger,
	methodRegistry *encoding.DynamicRegistry,
	invoker *types.RouterInvoker,
) []common.RouteDefinition {
	return common.BuildWSRoutes(handler, methodRegistry, invoker, logger)
}

// NewServer constructs a WebSocket server for the stream module using the provided options.
func NewServer(opts ServerOptions) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = log.NewNopLogger()
	}
	logger = logger.With("component", "stream-websocket")

	if opts.MethodRegistry == nil {
		logger.Error("method registry is nil; skipping websocket server setup")
		return nil
	}
	if opts.Invoker == nil {
		logger.Error("query invoker is nil; skipping websocket server setup")
		return nil
	}
	if opts.Registry == nil {
		logger.Error("subscription registry is nil; skipping websocket server setup")
		return nil
	}

	upgrader := opts.Upgrader
	if upgrader == nil {
		upgrader = common.GetUpgrader(nil)
	}

	handler := &common.Handler{
		Logger:         logger,
		Registry:       opts.Registry,
		AppContextDone: opts.ContextDone,
		Upgrader:       upgrader,
	}

	wsRoutes := buildRoutes(handler, logger, opts.MethodRegistry, opts.Invoker)

	server := &Server{
		routes:   wsRoutes,
		registry: opts.Registry,
		logger:   logger,
	}

	return server
}

// RegisterRoutes registers all WebSocket routes with the provided function.
func (s *Server) RegisterRoutes(register func(pattern string, handler http.Handler)) {
	if s == nil || len(s.routes) == 0 {
		return
	}
	for _, route := range s.routes {
		register(route.Pattern, http.HandlerFunc(route.Handler))
	}
}

// HasRoutes reports whether the server produced any websocket routes.
func (s *Server) HasRoutes() bool {
	return s != nil && len(s.routes) > 0
}
