package app

import (
	"net/http"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/server/api"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RegisterWebSocketRoutes registers WebSocket routes for the stream module
func (app *App) RegisterWebSocketRoutes(apiSvr *api.Server) error {
	// Get the websocket handler from stream keeper
	wsHandler := app.AppKeepers.StreamKeeper.WebSocketHandler()
	if wsHandler == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "websocket handler not initialized")
	}

	// Automatically register all WebSocket routes from modules
	wsHandler.RegisterRoutes(func(pattern string, handler http.Handler) {
		apiSvr.Router.Handle(pattern, handler)
	})

	return nil
}
