package common

import (
	"net/http"
)

// RouteHandler defines a function that handles a specific route
type RouteHandler func(http.ResponseWriter, *http.Request)

// ModuleHandler provides a generic handler for WebSocket modules
type ModuleHandler struct {
	QueryHandler
	routes map[string]RouteHandler
}

// NewModuleHandler creates a new generic module handler
func NewModuleHandler(deps *HandlerDependencies, provider QueryContextProvider) *ModuleHandler {
	return &ModuleHandler{
		QueryHandler: NewQueryHandler(deps, provider),
		routes:       make(map[string]RouteHandler),
	}
}

// RegisterRoute registers a route handler
func (h *ModuleHandler) RegisterRoute(pattern string, handler RouteHandler) {
	h.routes[pattern] = handler
}

// GetRoutes returns all registered routes
func (h *ModuleHandler) GetRoutes() map[string]RouteHandler {
	return h.routes
}

// ServeHTTP implements http.Handler interface for the module
func (h *ModuleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract the route pattern from the request
	// This is a simplified version - in practice you'd use a router
	pattern := r.URL.Path
	if handler, ok := h.routes[pattern]; ok {
		handler(w, r)
	} else {
		http.NotFound(w, r)
	}
}