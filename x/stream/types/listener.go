package types

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

// StreamingListener implements the ABCIListener interface for the stream module
type StreamingListener struct {
	logger log.Logger
	intake chan<- encoding.StreamEvent
	// decoders map[string]schema.ModuleCodec
}

// NewStreamingListener creates a new streaming listener
func NewStreamingListener(
	intake chan<- encoding.StreamEvent,
	logger log.Logger,
	// decoders map[string]schema.ModuleCodec,
) *StreamingListener {
	return &StreamingListener{
		logger: logger,
		intake: intake,
		// decoders: decoders,
	}
}

// ListenFinalizeBlock implements the ABCIListener interface
func (*StreamingListener) ListenFinalizeBlock(_ context.Context, _ abci.RequestFinalizeBlock, _ abci.ResponseFinalizeBlock) error {
	return nil
}

// ListenCommit implements the ABCIListener interface
func (l *StreamingListener) ListenCommit(_ context.Context, _ abci.ResponseCommit, _ []*storetypes.StoreKVPair) error {
	// Event-based decoding is in a weird limbo because of the way the SDK handles collections versus
	// grpc endpoints so there is no clean translation between kv state changes and actual query endpoints.
	// So we just emit a single event to refresh all active subscriptions every block for the time being.
	//
	// for _, kvPair := range changeSet {
	//     if kvPair == nil {
	//         continue
	//     }

	// 	events, err := l.parseStoreEvents(kvPair.StoreKey, kvPair.Key)
	// 	if err != nil {
	// 		l.logger.Error("failed to parse store event", "error", err, "store", kvPair.StoreKey)
	// 		continue
	// 	}

	//     l.emitEvents(events)
	// }

	// temp solution until automatic event decoding is better supported
	l.emitEvents([]encoding.StreamEvent{{
		Module: "new_block",
		Method: "temp",
		Params: map[string]string{},
	}})

	return nil
}

func (l *StreamingListener) emitEvents(events []encoding.StreamEvent) {
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		select {
		case l.intake <- event:
			l.logger.Debug("sent event to intake", "key", event.String())
		default:
			channelLen := len(l.intake)
			channelCap := cap(l.intake)
			fillPercent := 0.0
			if channelCap > 0 {
				fillPercent = float64(channelLen) / float64(channelCap) * 100
			}

			if fillPercent >= 95 {
				l.logger.Warn("intake channel full, dropping event",
					"key", event.String(),
					"channel_len", channelLen,
					"channel_cap", channelCap,
					"fill_percent", fillPercent)
				continue
			}

			timer := time.NewTimer(5 * time.Millisecond)
			select {
			case l.intake <- event:
				timer.Stop()
				l.logger.Debug("sent event to intake after wait", "key", event.String())
			case <-timer.C:
				l.logger.Warn("intake channel full after wait, dropping event",
					"key", event.String(),
					"channel_len", channelLen,
					"channel_cap", channelCap,
					"fill_percent", fillPercent)
			}
		}
	}
}
