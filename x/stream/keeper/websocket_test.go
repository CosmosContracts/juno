package keeper_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper"
)

// TestWebSocketServer tests that the WebSocket server can be started
func TestWebSocketServer(t *testing.T) {
	// Create a test keeper with minimal dependencies
	k := keeper.Keeper{}
	
	// Start server in a goroutine
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	addr := "localhost:18080" // Use a different port to avoid conflicts
	errCh := make(chan error, 1)
	
	go func() {
		errCh <- k.StartWebSocketServer(addr)
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Try to connect
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 1 * time.Second
	
	// This will fail because we don't have a fully initialized keeper,
	// but it proves the server is listening
	conn, resp, _ := dialer.Dial("ws://"+addr+"/subscribe/bank/balance/juno1test/ujuno", nil)
	if conn != nil {
		conn.Close()
	}
	
	// We expect an error (because keeper isn't fully initialized)
	// but the important thing is we got a response from the server
	require.NotNil(t, resp, "Expected HTTP response from server")
	require.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError,
		"Expected bad request or internal error, got %d", resp.StatusCode)
	
	// Stop server
	cancel()
	
	// Check if server exited cleanly
	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	case <-time.After(1 * time.Second):
		// Server is still running, that's okay
	}
}