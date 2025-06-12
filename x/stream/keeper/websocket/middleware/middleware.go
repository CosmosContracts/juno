package middleware

import (
	"fmt"
	"sync"
	"time"
)

type CircuitBreakerState string

const (
	// CircuitBreakerClosed allows all requests through
	CircuitBreakerClosed CircuitBreakerState = "closed"
	// CircuitBreakerOpen blocks all requests
	CircuitBreakerOpen CircuitBreakerState = "open"
	// CircuitBreakerHalfOpen allows limited requests through for testing
	CircuitBreakerHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreaker implements a circuit breaker pattern for connection protection
type CircuitBreaker struct {
	failures     map[string]int                 // connectionID -> failure count
	lastFailTime map[string]time.Time           // connectionID -> last failure time
	state        map[string]CircuitBreakerState // connectionID -> state
	threshold    int                            // failure threshold to open circuit
	timeout      time.Duration                  // timeout before attempting to close circuit
	mu           sync.RWMutex                   // protects all maps
}

// NewCircuitBreaker creates a new circuit breaker instance
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failures:     make(map[string]int),
		lastFailTime: make(map[string]time.Time),
		state:        make(map[string]CircuitBreakerState),
		threshold:    threshold,
		timeout:      timeout,
	}
}

// RecordSuccess records a successful operation for a connection
func (cb *CircuitBreaker) RecordSuccess(connectionID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Reset failures on success
	delete(cb.failures, connectionID)
	delete(cb.lastFailTime, connectionID)

	// If in half-open state, close the circuit
	if cb.state[connectionID] == CircuitBreakerHalfOpen {
		cb.state[connectionID] = CircuitBreakerClosed
	}
}

// RecordFailure records a failed operation for a connection
func (cb *CircuitBreaker) RecordFailure(connectionID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures[connectionID]++
	cb.lastFailTime[connectionID] = time.Now()

	// If failures exceed threshold, open the circuit
	if cb.failures[connectionID] >= cb.threshold {
		cb.state[connectionID] = CircuitBreakerOpen
	}
}

// AllowRequest checks if a request should be allowed for a connection
func (cb *CircuitBreaker) AllowRequest(connectionID string) (bool, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, exists := cb.state[connectionID]
	if !exists {
		// Default to closed state for new connections
		cb.state[connectionID] = CircuitBreakerClosed
		state = CircuitBreakerClosed
	}

	switch state {
	case CircuitBreakerClosed:
		return true, nil

	case CircuitBreakerOpen:
		// Check if timeout has elapsed
		lastFail, exists := cb.lastFailTime[connectionID]
		if exists && time.Since(lastFail) > cb.timeout {
			// Move to half-open state
			cb.state[connectionID] = CircuitBreakerHalfOpen
			return true, nil
		}
		return false, fmt.Errorf("circuit breaker is open for connection %s", connectionID)

	case CircuitBreakerHalfOpen:
		// Allow request through for testing
		return true, nil

	default:
		return false, fmt.Errorf("unknown circuit breaker state: %s", state)
	}
}

// GetState returns the current state of the circuit breaker for a connection
func (cb *CircuitBreaker) GetState(connectionID string) CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state, exists := cb.state[connectionID]
	if !exists {
		return CircuitBreakerClosed
	}
	return state
}

// Reset resets the circuit breaker for a specific connection
func (cb *CircuitBreaker) Reset(connectionID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	delete(cb.failures, connectionID)
	delete(cb.lastFailTime, connectionID)
	delete(cb.state, connectionID)
}

// ResetAll resets all circuit breakers
func (cb *CircuitBreaker) ResetAll() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = make(map[string]int)
	cb.lastFailTime = make(map[string]time.Time)
	cb.state = make(map[string]CircuitBreakerState)
}

// GetMetrics returns current circuit breaker metrics for monitoring
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	openCount := 0
	halfOpenCount := 0
	closedCount := 0

	for _, state := range cb.state {
		switch state {
		case CircuitBreakerOpen:
			openCount++
		case CircuitBreakerHalfOpen:
			halfOpenCount++
		case CircuitBreakerClosed:
			closedCount++
		}
	}

	return map[string]any{
		"total_connections":  len(cb.state),
		"open_circuits":      openCount,
		"half_open_circuits": halfOpenCount,
		"closed_circuits":    closedCount,
		"threshold":          cb.threshold,
		"timeout_seconds":    cb.timeout.Seconds(),
	}
}

// CleanupStaleConnections removes entries for connections that no longer exist
func (cb *CircuitBreaker) CleanupStaleConnections(activeConnections map[string]bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Remove entries for connections that are no longer active
	for connectionID := range cb.failures {
		if !activeConnections[connectionID] {
			delete(cb.failures, connectionID)
			delete(cb.lastFailTime, connectionID)
			delete(cb.state, connectionID)
		}
	}
}

// UpdateThreshold updates the failure threshold
func (cb *CircuitBreaker) UpdateThreshold(threshold int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.threshold = threshold
}

// UpdateTimeout updates the circuit breaker timeout
func (cb *CircuitBreaker) UpdateTimeout(timeout time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.timeout = timeout
}

// SetLastFailTimeForTesting sets the last failure time for a connection (for testing only)
func (cb *CircuitBreaker) SetLastFailTimeForTesting(connectionID string, t time.Time) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.lastFailTime[connectionID] = t
}
