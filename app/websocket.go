package app

import (
	"errors"
	"net/http"

	"github.com/cosmos/cosmos-sdk/server/api"
)

// RegisterWebSocketRoutes registers WebSocket routes for the stream module
func (app *App) RegisterWebSocketRoutes(apiSvr *api.Server) error {
	// Get the websocket handler from stream keeper
	wsHandler := app.AppKeepers.StreamKeeper.WebSocketHandler()
	if wsHandler == nil {
		return errors.New("websocket handler not initialized")
	}

	// Automatically register all WebSocket routes from modules
	wsHandler.RegisterRoutes(func(pattern string, handler http.Handler) {
		apiSvr.Router.Handle(pattern, handler)
	})

	return nil
}
