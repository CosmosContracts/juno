package types

import (
	"context"
	"errors"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/gogoproto/types"

	"cosmossdk.io/log"

	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

func TestRunStreamSkipsIdenticalPayloads(t *testing.T) {
	registry := NewSubscriptionRegistry(log.NewNopLogger(), StreamConfig{
		SubscriptionBufferSize:    4,
		WsMaxConnections:          math.MaxUint32,
		GrpcMaxConnections:        math.MaxUint32,
		MaxSubscriptionsPerClient: math.MaxUint32,
	})

	key := encoding.StreamEvent{Module: "bank", Method: "balances"}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	responses := make(chan proto.Message, 3)
	responses <- &types.Any{Value: []byte("v1")}
	responses <- &types.Any{Value: []byte("v1")}
	responses <- &types.Any{Value: []byte("v2")}
	close(responses)

	var sendCount atomic.Int32
	params := RunStreamParams{
		Registry:       registry,
		Key:            key,
		ConnectionKind: ConnectionKindWebsocket,
		ConnectionMeta: ConnectionMetadata{RemoteAddr: "127.0.0.1:1"},
		Fetch: func(context.Context) (proto.Message, error) {
			msg, ok := <-responses
			if !ok {
				return nil, context.Canceled
			}
			return proto.Clone(msg), nil
		},
		Send: func(proto.Message) error {
			sendCount.Add(1)
			return nil
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunStream(ctx, params)
	}()

	// Wait for the initial snapshot to be delivered
	require.Eventually(t, func() bool { return sendCount.Load() == 1 }, time.Second, 10*time.Millisecond)

	// First event has identical payload, should be skipped
	registry.FanOut(key, key)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, int32(1), sendCount.Load())

	// Second event changes payload, should be forwarded
	registry.FanOut(key, key)
	require.Eventually(t, func() bool { return sendCount.Load() == 2 }, time.Second, 10*time.Millisecond)

	cancel()
	require.Eventually(t, func() bool {
		select {
		case err := <-errCh:
			return errors.Is(err, context.Canceled)
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}
