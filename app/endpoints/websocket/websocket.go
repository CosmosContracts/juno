package websocket

import (
	"net/http"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/server/api"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v30/app/endpoints/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper"
)

// RegisterRoutes registers WebSocket routes for the stream module using the configuration
// loaded from the node's home directory.
func RegisterRoutes(apiSvr *api.Server, sk *keeper.Keeper, homePath string) error {
	if sk == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "stream keeper not initialized")
	}

	cfg, err := LoadConfig(homePath)
	if err != nil {
		sk.Logger().Debug("failed to load websocket config", "error", err)
		cfg = Config{}
	}

	appCtx := sk.GetAppContext()
	var appDone <-chan struct{}
	if appCtx != nil {
		appDone = appCtx.Done()
	}

	server := NewServer(ServerOptions{
		ContextDone:    appDone,
		MethodRegistry: sk.MethodRegistry(),
		Registry:       sk.Registry(),
		Invoker:        sk.Invoker(),
		Logger:         sk.Logger(),
		Upgrader:       common.GetUpgrader(cfg.CORSAllowedOrigins),
	})
	if server == nil || !server.HasRoutes() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "websocket server not initialized")
	}

	server.RegisterRoutes(func(pattern string, handler http.Handler) {
		apiSvr.Router.Handle(pattern, handler)
	})

	return nil
}
