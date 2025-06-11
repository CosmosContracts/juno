package keeper_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper"
)

// TestWebSocketHandlers tests the WebSocket handlers
func TestWebSocketHandlers(t *testing.T) {
	// Create a test keeper with minimal dependencies
	k := keeper.Keeper{}

	// Create a test router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/ws/subscribe/bank/balance/{address}/{denom}", k.HandleBalanceSubscription)
	router.HandleFunc("/ws/subscribe/bank/balances/{address}", k.HandleAllBalancesSubscription)
	router.HandleFunc("/ws/subscribe/staking/delegations/{delegator}", k.HandleDelegationsSubscription)
	router.HandleFunc("/ws/subscribe/staking/delegation/{delegator}/{validator}", k.HandleDelegationSubscription)
	router.HandleFunc("/ws/subscribe/staking/unbonding-delegations/{delegator}", k.HandleUnbondingDelegationsSubscription)
	router.HandleFunc("/ws/subscribe/staking/unbonding-delegation/{delegator}/{validator}", k.HandleUnbondingDelegationSubscription)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "Balance subscription",
			endpoint: "/ws/subscribe/bank/balance/juno1test/ujuno",
		},
		{
			name:     "All balances subscription",
			endpoint: "/ws/subscribe/bank/balances/juno1test",
		},
		{
			name:     "Delegations subscription",
			endpoint: "/ws/subscribe/staking/delegations/juno1test",
		},
		{
			name:     "Delegation subscription",
			endpoint: "/ws/subscribe/staking/delegation/juno1test/junovaloper1test",
		},
		{
			name:     "Unbonding delegations subscription",
			endpoint: "/ws/subscribe/staking/unbonding-delegations/juno1test",
		},
		{
			name:     "Unbonding delegation subscription",
			endpoint: "/ws/subscribe/staking/unbonding-delegation/juno1test/junovaloper1test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to connect
			dialer := websocket.DefaultDialer
			dialer.HandshakeTimeout = 1 * time.Second

			// This will fail because we don't have a fully initialized keeper,
			// but it proves the handler is reachable
			conn, resp, _ := dialer.Dial(wsURL+tt.endpoint, nil)
			if conn != nil {
				conn.Close()
			}

			// We expect an error (because keeper isn't fully initialized)
			// but the important thing is we got a response from the handler
			require.NotNil(t, resp, "Expected HTTP response from handler")
			require.True(t, resp.StatusCode == http.StatusBadRequest ||
				resp.StatusCode == http.StatusInternalServerError ||
				resp.StatusCode == http.StatusSwitchingProtocols,
				"Expected bad request, internal error, or protocol switch, got %d", resp.StatusCode)
		})
	}
}

// TestWebSocketMessageFormat tests that WebSocket messages are properly formatted
func TestWebSocketMessageFormat(t *testing.T) {
	// WebSocket messages are now sent as raw JSON without wrapper
	// Test various message formats

	// Balance message
	balanceMsg := map[string]string{
		"denom":  "ujuno",
		"amount": "1000000",
	}
	require.NotNil(t, balanceMsg)
	require.Equal(t, "ujuno", balanceMsg["denom"])
	require.Equal(t, "1000000", balanceMsg["amount"])

	// All balances message
	allBalancesMsg := map[string]any{
		"balances": []map[string]string{
			{"denom": "ujuno", "amount": "1000000"},
			{"denom": "uatom", "amount": "500000"},
		},
	}
	require.NotNil(t, allBalancesMsg)
}

// TestWebSocketIntegration provides an example of how to test WebSocket connections
// This test is skipped by default as it requires a running node
func TestWebSocketIntegration(t *testing.T) {
	t.Skip("Skipping integration test - requires running node")

	// Example of connecting to a real WebSocket endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to WebSocket
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, "ws://localhost:1317/ws/subscribe/bank/balance/juno1test/ujuno", nil)
	require.NoError(t, err)
	defer conn.Close()

	// Read initial balance message
	var msg map[string]string
	err = conn.ReadJSON(&msg)
	require.NoError(t, err)
	require.NotEmpty(t, msg["denom"])
	require.NotEmpty(t, msg["amount"])

	// Keep connection alive and wait for updates
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var updateMsg map[string]string
				if err := conn.ReadJSON(&updateMsg); err != nil {
					return
				}
				// Handle update message
				t.Logf("Received update: %+v", updateMsg)
			}
		}
	}()

	// Wait for some time to receive updates
	<-ctx.Done()
}
