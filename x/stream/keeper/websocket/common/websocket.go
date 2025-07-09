package common

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WebSocket configuration
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// GetUpgrader returns a websocket upgrader with configurable CORS
func GetUpgrader(allowAllOrigins bool) *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if allowAllOrigins {
				return true
			}
			// Default to same-origin only
			origin := r.Header.Get("Origin")
			return origin == "" || origin == "http://"+r.Host || origin == "https://"+r.Host
		},
	}
}

// HandleWebSocketConnection handles the WebSocket connection lifecycle
func HandleWebSocketConnection(ctx context.Context, conn *websocket.Conn, sendCh <-chan any, connectionID string, queryFunc func() any, sendFunc func(*websocket.Conn, any) error, onFailure func(string), onSuccess func(string)) {
	conn.SetReadLimit(maxMessageSize)
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	// Start ping ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	// Ensure we send a close message on exit
	defer func() {
		deadline := time.Now().Add(writeWait)
		if err := conn.SetWriteDeadline(deadline); err != nil {
			return
		}
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"))
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sendCh:
			// Re-query data and send update
			data := queryFunc()
			if err := sendFunc(conn, data); err != nil {
				if onFailure != nil {
					onFailure(connectionID)
				}
				return
			}
			if onSuccess != nil {
				onSuccess(connectionID)
			}
		case <-ticker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				if onFailure != nil {
					onFailure(connectionID)
				}
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if onFailure != nil {
					onFailure(connectionID)
				}
				return
			}
		}
	}
}

// SendWebSocketMessage sends a message over WebSocket
func SendWebSocketMessage(conn *websocket.Conn, data any, onSuccess func()) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	err := conn.WriteJSON(data)
	if err == nil && onSuccess != nil {
		onSuccess()
	}
	return err
}

// CloseWithMessage closes the WebSocket connection with a specific close code and message
func CloseWithMessage(conn *websocket.Conn, code int, message string) error {
	if len(message) > 123 {
		message = message[:123]
	}
	deadline := time.Now().Add(writeWait)
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return err
	}
	closeMsg := websocket.FormatCloseMessage(code, message)
	err := conn.WriteMessage(websocket.CloseMessage, closeMsg)
	if err != nil {
		// If we can't write the close message, try WriteControl
		_ = conn.WriteControl(websocket.CloseMessage, closeMsg, deadline)
	}
	return err
}
