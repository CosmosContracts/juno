package keeper

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/google/uuid"
)

// ConnectionInfo tracks information about a WebSocket connection
type ConnectionInfo struct {
	connectionID  string
	remoteAddr    string
	xForwardedFor string
	subscriptions int32
	createdAt     time.Time
}

// ConnectionManager manages WebSocket connections and enforces limits
type ConnectionManager struct {
	mu                        sync.RWMutex
	connections               map[string]*ConnectionInfo // connection ID -> connection info
	connectionsByAddr         map[string][]string        // remote addr -> list of connection IDs
	totalConnections          int32
	maxConnections            int32
	maxSubscriptionsPerClient int32
	enableUUID                bool
	logger                    log.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(maxConnections, maxSubscriptionsPerClient int, logger log.Logger) *ConnectionManager {
	return &ConnectionManager{
		connections:               make(map[string]*ConnectionInfo),
		connectionsByAddr:         make(map[string][]string),
		maxConnections:            int32(maxConnections),
		maxSubscriptionsPerClient: int32(maxSubscriptionsPerClient),
		enableUUID:                true,
		logger:                    logger.With("component", "connection-manager"),
	}
}

// CanAcceptConnection checks if a new connection can be accepted
func (cm *ConnectionManager) CanAcceptConnection() bool {
	return atomic.LoadInt32(&cm.totalConnections) < cm.maxConnections
}

// RegisterConnection registers a new connection with UUID support
func (cm *ConnectionManager) RegisterConnection(remoteAddr string) string {
	return cm.RegisterConnectionWithHeaders(remoteAddr, "")
}

// RegisterConnectionWithHeaders registers a new connection with optional X-Forwarded-For header
func (cm *ConnectionManager) RegisterConnectionWithHeaders(remoteAddr, xForwardedFor string) string {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if we can accept new connections
	if cm.totalConnections >= cm.maxConnections {
		cm.logger.Warn("connection limit reached", "limit", cm.maxConnections)
		types.IncrementConnectionRejected("max_connections")
		return ""
	}

	// Generate unique connection ID
	connectionID := remoteAddr
	if cm.enableUUID {
		connectionID = uuid.New().String()
	}

	// Register the connection
	cm.connections[connectionID] = &ConnectionInfo{
		connectionID:  connectionID,
		remoteAddr:    remoteAddr,
		xForwardedFor: xForwardedFor,
		createdAt:     time.Now(),
	}

	// Track by address for cleanup
	cm.connectionsByAddr[remoteAddr] = append(cm.connectionsByAddr[remoteAddr], connectionID)

	atomic.AddInt32(&cm.totalConnections, 1)

	cm.logger.Debug("connection registered",
		"connection_id", connectionID,
		"remote_addr", remoteAddr,
		"x_forwarded_for", xForwardedFor,
		"total", cm.totalConnections)
	types.UpdateConnectionMetrics(cm.totalConnections)
	return connectionID
}

// UnregisterConnection removes a connection by connection ID
func (cm *ConnectionManager) UnregisterConnection(connectionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[connectionID]; exists {
		// Record connection duration
		types.RecordConnectionDuration(conn.createdAt)

		// Remove from connections map
		delete(cm.connections, connectionID)

		// Remove from address tracking
		if addrs, ok := cm.connectionsByAddr[conn.remoteAddr]; ok {
			newAddrs := make([]string, 0, len(addrs)-1)
			for _, id := range addrs {
				if id != connectionID {
					newAddrs = append(newAddrs, id)
				}
			}
			if len(newAddrs) == 0 {
				delete(cm.connectionsByAddr, conn.remoteAddr)
			} else {
				cm.connectionsByAddr[conn.remoteAddr] = newAddrs
			}
		}

		atomic.AddInt32(&cm.totalConnections, -1)
		cm.logger.Debug("connection unregistered",
			"connection_id", connectionID,
			"remote_addr", conn.remoteAddr,
			"total", cm.totalConnections)
		types.UpdateConnectionMetrics(cm.totalConnections)
	}
}

// CanAddSubscription checks if a connection can add another subscription
func (cm *ConnectionManager) CanAddSubscription(connectionID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connectionID]
	if !exists {
		return false
	}

	return atomic.LoadInt32(&conn.subscriptions) < cm.maxSubscriptionsPerClient
}

// AddSubscription increments the subscription count for a connection
func (cm *ConnectionManager) AddSubscription(connectionID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connectionID]
	if !exists {
		return false
	}

	if atomic.LoadInt32(&conn.subscriptions) >= cm.maxSubscriptionsPerClient {
		cm.logger.Warn("subscription limit reached for connection",
			"connection_id", connectionID,
			"remote_addr", conn.remoteAddr,
			"limit", cm.maxSubscriptionsPerClient)
		types.IncrementConnectionRejected("max_subscriptions")
		return false
	}

	atomic.AddInt32(&conn.subscriptions, 1)
	return true
}

// RemoveSubscription decrements the subscription count for a connection
func (cm *ConnectionManager) RemoveSubscription(connectionID string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if conn, exists := cm.connections[connectionID]; exists {
		atomic.AddInt32(&conn.subscriptions, -1)
	}
}

// GetStats returns connection statistics
func (cm *ConnectionManager) GetStats() (totalConnections int32, connectionDetails map[string]int32) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	totalConnections = cm.totalConnections
	connectionDetails = make(map[string]int32)

	for addr, conn := range cm.connections {
		connectionDetails[addr] = atomic.LoadInt32(&conn.subscriptions)
	}

	return
}

// CheckConnectionLimits is a helper that returns appropriate HTTP error if limits are exceeded
func (cm *ConnectionManager) CheckConnectionLimits(w http.ResponseWriter, r *http.Request) bool {
	// Check if we can accept new connections
	if !cm.CanAcceptConnection() {
		http.Error(w, "connection limit exceeded", http.StatusServiceUnavailable)
		return false
	}

	return true
}

// SetEnableUUID configures whether to use UUIDs for connection tracking
func (cm *ConnectionManager) SetEnableUUID(enable bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.enableUUID = enable
	cm.logger.Info("UUID connection tracking configured", "enabled", enable)
}

// GetActiveConnections returns a map of all active connection IDs
func (cm *ConnectionManager) GetActiveConnections() map[string]bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	activeConnections := make(map[string]bool, len(cm.connections))
	for connID := range cm.connections {
		activeConnections[connID] = true
	}
	return activeConnections
}
