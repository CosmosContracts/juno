package common

import (
	"fmt"
	"net/http"
)

// RouteDefinition defines a WebSocket route with its pattern and handler
type RouteDefinition struct {
	Pattern string       // The full route pattern, e.g., "/ws/subscribe/bank/balance/{address}/{denom}"
	Handler RouteHandler // The handler function for this route
}

// Module defines the interface for a WebSocket module
type Module interface {
	// Name returns the module name (e.g., "bank", "staking")
	Name() string

	// Routes returns all routes this module handles
	// Keys should be route patterns, values are the handlers
	Routes() map[string]RouteHandler

	// RouteDefinitions returns all route definitions with full patterns
	RouteDefinitions() []RouteDefinition
}

// ModuleRegistry manages WebSocket modules
type ModuleRegistry struct {
	modules map[string]Module
	routes  map[string]RouteInfo
}

// RouteInfo contains information about a registered route
type RouteInfo struct {
	Module  string
	Handler RouteHandler
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[string]Module),
		routes:  make(map[string]RouteInfo),
	}
}

// RegisterModule registers a new module
func (r *ModuleRegistry) RegisterModule(module Module) error {
	name := module.Name()
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %s already registered", name)
	}

	r.modules[name] = module

	// Register all routes from the module using RouteDefinitions if available
	if routeDefiner, ok := module.(interface{ RouteDefinitions() []RouteDefinition }); ok {
		for _, routeDef := range routeDefiner.RouteDefinitions() {
			if _, exists := r.routes[routeDef.Pattern]; exists {
				return fmt.Errorf("route %s already registered", routeDef.Pattern)
			}
			r.routes[routeDef.Pattern] = RouteInfo{
				Module:  name,
				Handler: routeDef.Handler,
			}
		}
	} else {
		// Fallback to old method for backward compatibility
		for pattern, handler := range module.Routes() {
			if _, exists := r.routes[pattern]; exists {
				return fmt.Errorf("route %s already registered", pattern)
			}
			r.routes[pattern] = RouteInfo{
				Module:  name,
				Handler: handler,
			}
		}
	}

	return nil
}

// GetHandler returns the handler for a given route pattern
func (r *ModuleRegistry) GetHandler(pattern string) (RouteHandler, bool) {
	info, exists := r.routes[pattern]
	if !exists {
		return nil, false
	}
	return info.Handler, true
}

// GetModule returns a module by name
func (r *ModuleRegistry) GetModule(name string) (Module, bool) {
	module, exists := r.modules[name]
	return module, exists
}

// ListModules returns all registered module names
func (r *ModuleRegistry) ListModules() []string {
	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}

// ListRoutes returns all registered routes with their module information
func (r *ModuleRegistry) ListRoutes() map[string]string {
	routes := make(map[string]string)
	for pattern, info := range r.routes {
		routes[pattern] = info.Module
	}
	return routes
}

// GetAllRouteDefinitions returns all registered route definitions
func (r *ModuleRegistry) GetAllRouteDefinitions() []RouteDefinition {
	defs := make([]RouteDefinition, 0, len(r.routes))
	for pattern, info := range r.routes {
		defs = append(defs, RouteDefinition{
			Pattern: pattern,
			Handler: info.Handler,
		})
	}
	return defs
}

// ServeHTTP implements http.Handler for the registry
func (r *ModuleRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pattern := req.URL.Path
	handler, exists := r.GetHandler(pattern)
	if !exists {
		http.NotFound(w, req)
		return
	}
	handler(w, req)
}
