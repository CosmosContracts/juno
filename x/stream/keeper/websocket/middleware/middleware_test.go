package middleware_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/middleware"
)

func TestCircuitBreaker(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "new circuit breaker starts closed",
			test: testNewCircuitBreakerStartsClosed,
		},
		{
			name: "circuit opens after threshold failures",
			test: testCircuitOpensAfterThreshold,
		},
		{
			name: "circuit moves to half-open after timeout",
			test: testCircuitMovesToHalfOpen,
		},
		{
			name: "circuit closes on success in half-open state",
			test: testCircuitClosesOnSuccess,
		},
		{
			name: "circuit breaker reset",
			test: testCircuitBreakerReset,
		},
		{
			name: "circuit breaker metrics",
			test: testCircuitBreakerMetrics,
		},
		{
			name: "cleanup stale connections",
			test: testCleanupStaleConnections,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testNewCircuitBreakerStartsClosed(t *testing.T) {
	cb := middleware.NewCircuitBreaker(3, 30*time.Second)
	require.NotNil(t, cb)

	// New connection should be allowed
	allowed, err := cb.AllowRequest("conn1")
	require.True(t, allowed)
	require.NoError(t, err)

	// State should be closed
	state := cb.GetState("conn1")
	require.Equal(t, middleware.CircuitBreakerClosed, state)
}

func testCircuitOpensAfterThreshold(t *testing.T) {
	cb := middleware.NewCircuitBreaker(3, 30*time.Second)

	// Record failures up to threshold
	for i := 0; i < 3; i++ {
		cb.RecordFailure("conn1")
	}

	// Circuit should now be open
	allowed, err := cb.AllowRequest("conn1")
	require.False(t, allowed)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circuit breaker is open")

	state := cb.GetState("conn1")
	require.Equal(t, middleware.CircuitBreakerOpen, state)
}

func testCircuitMovesToHalfOpen(t *testing.T) {
	// Use a short timeout for testing
	cb := middleware.NewCircuitBreaker(2, 100*time.Millisecond)

	// Open the circuit
	cb.RecordFailure("conn1")
	cb.RecordFailure("conn1")

	// Verify circuit is open
	allowed, err := cb.AllowRequest("conn1")
	require.False(t, allowed)
	require.Error(t, err)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Circuit should now allow request (half-open)
	allowed, err = cb.AllowRequest("conn1")
	require.True(t, allowed)
	require.NoError(t, err)

	state := cb.GetState("conn1")
	require.Equal(t, middleware.CircuitBreakerHalfOpen, state)
}

func testCircuitClosesOnSuccess(t *testing.T) {
	cb := middleware.NewCircuitBreaker(2, 100*time.Millisecond)

	// Open the circuit
	cb.RecordFailure("conn1")
	cb.RecordFailure("conn1")

	// Wait for timeout to move to half-open
	time.Sleep(150 * time.Millisecond)

	// First request should be allowed (half-open)
	allowed, err := cb.AllowRequest("conn1")
	require.True(t, allowed)
	require.NoError(t, err)

	// Record success
	cb.RecordSuccess("conn1")

	// Circuit should now be closed
	state := cb.GetState("conn1")
	require.Equal(t, middleware.CircuitBreakerClosed, state)

	// Verify failures are reset
	cb.RecordFailure("conn1")
	allowed, err = cb.AllowRequest("conn1")
	require.True(t, allowed)
	require.NoError(t, err)
}

func testCircuitBreakerReset(t *testing.T) {
	cb := middleware.NewCircuitBreaker(2, 30*time.Second)

	// Open circuit for multiple connections
	cb.RecordFailure("conn1")
	cb.RecordFailure("conn1")
	cb.RecordFailure("conn2")
	cb.RecordFailure("conn2")

	// Verify both are open
	allowed1, _ := cb.AllowRequest("conn1")
	allowed2, _ := cb.AllowRequest("conn2")
	require.False(t, allowed1)
	require.False(t, allowed2)

	// Reset conn1
	cb.Reset("conn1")

	// conn1 should be allowed, conn2 still blocked
	allowed1, _ = cb.AllowRequest("conn1")
	allowed2, _ = cb.AllowRequest("conn2")
	require.True(t, allowed1)
	require.False(t, allowed2)

	// Reset all
	cb.ResetAll()

	// Both should be allowed
	allowed1, _ = cb.AllowRequest("conn1")
	allowed2, _ = cb.AllowRequest("conn2")
	require.True(t, allowed1)
	require.True(t, allowed2)
}

func testCircuitBreakerMetrics(t *testing.T) {
	cb := middleware.NewCircuitBreaker(2, 30*time.Second)

	// Create various states
	cb.RecordFailure("conn1")
	cb.RecordFailure("conn1") // Open

	cb.RecordFailure("conn2")       // Still closed (1 failure)
	_, _ = cb.AllowRequest("conn2") // This ensures conn2 is tracked in state

	// conn3 will be half-open
	cb.RecordFailure("conn3")
	cb.RecordFailure("conn3")
	cb.SetLastFailTimeForTesting("conn3", time.Now().Add(-31*time.Second)) // Force timeout
	_, _ = cb.AllowRequest("conn3")                                        // This will move it to half-open

	// Get metrics
	metrics := cb.GetMetrics()
	require.Equal(t, 3, metrics["total_connections"])
	require.Equal(t, 1, metrics["open_circuits"])
	require.Equal(t, 1, metrics["half_open_circuits"])
	require.Equal(t, 1, metrics["closed_circuits"])
	require.Equal(t, 2, metrics["threshold"])
	require.Equal(t, float64(30), metrics["timeout_seconds"])
}

func testCleanupStaleConnections(t *testing.T) {
	cb := middleware.NewCircuitBreaker(2, 30*time.Second)

	// Create some circuit states by calling AllowRequest first
	_, _ = cb.AllowRequest("conn1")
	_, _ = cb.AllowRequest("conn2")
	_, _ = cb.AllowRequest("conn3")

	// Now record failures
	cb.RecordFailure("conn1")
	cb.RecordFailure("conn2")
	cb.RecordFailure("conn3")

	// Mark only conn1 and conn3 as active
	activeConnections := map[string]bool{
		"conn1": true,
		"conn3": true,
	}

	// Cleanup stale connections
	cb.CleanupStaleConnections(activeConnections)

	// conn2 should be removed
	metrics := cb.GetMetrics()
	require.Equal(t, 2, metrics["total_connections"])

	// Verify conn2 starts fresh if it comes back
	allowed, err := cb.AllowRequest("conn2")
	require.True(t, allowed)
	require.NoError(t, err)
	require.Equal(t, middleware.CircuitBreakerClosed, cb.GetState("conn2"))
}
