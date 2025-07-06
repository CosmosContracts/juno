package common

import (
	"fmt"
	"net/http"
)

// Module defines the interface for a WebSocket module
type Module interface {
	// Name returns the module name (e.g., "bank", "staking")
	Name() string
	
	// Routes returns all routes this module handles
	// Keys should be route patterns, values are the handlers
	Routes() map[string]RouteHandler
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
	
	// Register all routes from the module
	for pattern, handler := range module.Routes() {
		if _, exists := r.routes[pattern]; exists {
			return fmt.Errorf("route %s already registered", pattern)
		}
		r.routes[pattern] = RouteInfo{
			Module:  name,
			Handler: handler,
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