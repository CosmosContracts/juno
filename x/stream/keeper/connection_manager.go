package keeper

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
)

// ConnectionInfo tracks information about a WebSocket connection
type ConnectionInfo struct {
	remoteAddr    string
	subscriptions int32
	createdAt     time.Time
}

// ConnectionManager manages WebSocket connections and enforces limits
type ConnectionManager struct {
	mu                        sync.RWMutex
	connections               map[string]*ConnectionInfo // remote addr -> connection info
	totalConnections          int32
	maxConnections            int32
	maxSubscriptionsPerClient int32
	logger                    log.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(maxConnections, maxSubscriptionsPerClient int, logger log.Logger) *ConnectionManager {
	return &ConnectionManager{
		connections:               make(map[string]*ConnectionInfo),
		maxConnections:            int32(maxConnections),
		maxSubscriptionsPerClient: int32(maxSubscriptionsPerClient),
		logger:                    logger.With("component", "connection-manager"),
	}
}

// CanAcceptConnection checks if a new connection can be accepted
func (cm *ConnectionManager) CanAcceptConnection() bool {
	return atomic.LoadInt32(&cm.totalConnections) < cm.maxConnections
}

// RegisterConnection registers a new connection
func (cm *ConnectionManager) RegisterConnection(remoteAddr string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if we can accept new connections
	if cm.totalConnections >= cm.maxConnections {
		cm.logger.Warn("connection limit reached", "limit", cm.maxConnections)
		IncrementConnectionRejected("max_connections")
		return false
	}

	// Register the connection
	cm.connections[remoteAddr] = &ConnectionInfo{
		remoteAddr: remoteAddr,
		createdAt:  time.Now(),
	}
	atomic.AddInt32(&cm.totalConnections, 1)

	cm.logger.Debug("connection registered", "remote_addr", remoteAddr, "total", cm.totalConnections)
	UpdateConnectionMetrics(cm.totalConnections)
	return true
}

// UnregisterConnection removes a connection
func (cm *ConnectionManager) UnregisterConnection(remoteAddr string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[remoteAddr]; exists {
		// Record connection duration
		RecordConnectionDuration(conn.createdAt)

		delete(cm.connections, remoteAddr)
		atomic.AddInt32(&cm.totalConnections, -1)
		cm.logger.Debug("connection unregistered", "remote_addr", remoteAddr, "total", cm.totalConnections)
		UpdateConnectionMetrics(cm.totalConnections)
	}
}

// CanAddSubscription checks if a connection can add another subscription
func (cm *ConnectionManager) CanAddSubscription(remoteAddr string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[remoteAddr]
	if !exists {
		return false
	}

	return atomic.LoadInt32(&conn.subscriptions) < cm.maxSubscriptionsPerClient
}

// AddSubscription increments the subscription count for a connection
func (cm *ConnectionManager) AddSubscription(remoteAddr string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[remoteAddr]
	if !exists {
		return false
	}

	if atomic.LoadInt32(&conn.subscriptions) >= cm.maxSubscriptionsPerClient {
		cm.logger.Warn("subscription limit reached for connection",
			"remote_addr", remoteAddr,
			"limit", cm.maxSubscriptionsPerClient)
		IncrementConnectionRejected("max_subscriptions")
		return false
	}

	atomic.AddInt32(&conn.subscriptions, 1)
	return true
}

// RemoveSubscription decrements the subscription count for a connection
func (cm *ConnectionManager) RemoveSubscription(remoteAddr string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if conn, exists := cm.connections[remoteAddr]; exists {
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
