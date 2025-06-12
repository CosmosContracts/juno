package common

import (
	"net/http"
)

// ConnectionManagerAdapter adapts the keeper's ConnectionManager to common.ConnectionManager
type ConnectionManagerAdapter struct {
	manager ConnectionManagerInterface
}

// ConnectionManagerInterface defines the minimal interface we need from the keeper's ConnectionManager
type ConnectionManagerInterface interface {
	CheckConnectionLimits(w http.ResponseWriter, r *http.Request) bool
	RegisterConnectionWithHeaders(remoteAddr, xForwardedFor string) string
	UnregisterConnection(connectionID string)
	AddSubscription(connectionID string) bool
	RemoveSubscription(connectionID string)
}

// NewConnectionManagerAdapter creates a new connection manager adapter
func NewConnectionManagerAdapter(manager ConnectionManagerInterface) ConnectionManager {
	return &ConnectionManagerAdapter{manager: manager}
}

// CheckConnectionLimits implements ConnectionManager
func (c *ConnectionManagerAdapter) CheckConnectionLimits(w http.ResponseWriter, r *http.Request) bool {
	return c.manager.CheckConnectionLimits(w, r)
}

// RegisterConnectionWithHeaders implements ConnectionManager
func (c *ConnectionManagerAdapter) RegisterConnectionWithHeaders(remoteAddr, xForwardedFor string) string {
	return c.manager.RegisterConnectionWithHeaders(remoteAddr, xForwardedFor)
}

// UnregisterConnection implements ConnectionManager
func (c *ConnectionManagerAdapter) UnregisterConnection(connectionID string) {
	c.manager.UnregisterConnection(connectionID)
}

// AddSubscription implements ConnectionManager
func (c *ConnectionManagerAdapter) AddSubscription(connectionID string) bool {
	return c.manager.AddSubscription(connectionID)
}

// RemoveSubscription implements ConnectionManager
func (c *ConnectionManagerAdapter) RemoveSubscription(connectionID string) {
	c.manager.RemoveSubscription(connectionID)
}
