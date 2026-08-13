package common

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	proto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/log"

	"github.com/CosmosContracts/juno/v31/x/stream/types"
	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

func TestHandlerServeConnectionConnectionLimit(t *testing.T) {
	logger := log.NewNopLogger()
	registry := types.NewSubscriptionRegistry(logger, types.StreamConfig{
		WsMaxConnections:          1,
		MaxSubscriptionsPerClient: 1,
	})
	handler := &Handler{
		Logger:   logger,
		Registry: registry,
		Upgrader: GetUpgrader([]string{"*"}),
	}

	occupiedID, err := registry.RegisterConnection(types.ConnectionKindWebsocket, types.ConnectionMetadata{RemoteAddr: "existing"})
	require.NoError(t, err)
	t.Cleanup(func() {
		registry.UnregisterConnection(occupiedID)
	})

	resolved := &types.ResolvedStream{
		Descriptor: &encoding.MethodDescriptor{
			Module:     "bank",
			StreamName: "balance",
		},
		Fields: map[string]string{},
		Key: encoding.StreamEvent{
			Module: "bank",
			Method: "balance",
			Params: map[string]string{},
		},
	}

	wsHandler := func(w http.ResponseWriter, r *http.Request) {
		_ = handler.ServeConnection(ConnectionParams{
			Writer:   w,
			Request:  r,
			Resolved: resolved,
			Fetch: func(context.Context) (proto.Message, error) {
				return nil, nil
			},
		})
	}

	server := httptest.NewServer(http.HandlerFunc(wsHandler))
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	dialer := websocket.Dialer{HandshakeTimeout: time.Second}
	_, resp, err := dialer.Dial(wsURL, nil)
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	total, _ := registry.ConnectionStats()
	require.Equal(t, uint32(1), total)
}
