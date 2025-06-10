package keeper

import (
	"sync"
	"time"

	"cosmossdk.io/log"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// Dispatcher handles the routing of state events to subscriptions
type Dispatcher struct {
	intake   <-chan types.StreamEvent
	registry *SubscriptionRegistry
	logger   log.Logger
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewDispatcher creates a new event dispatcher
func NewDispatcher(intake <-chan types.StreamEvent, registry *SubscriptionRegistry, logger log.Logger) *Dispatcher {
	return &Dispatcher{
		intake:   intake,
		registry: registry,
		logger:   logger.With("component", "dispatcher"),
		stopCh:   make(chan struct{}),
	}
}

// Start begins the dispatcher event loop
func (d *Dispatcher) Start() {
	d.logger.Info("starting event dispatcher")

	ticker := time.NewTicker(30 * time.Second) // Stats ticker
	defer ticker.Stop()

	for {
		select {
		case event := <-d.intake:
			d.processEvent(event)

		case <-ticker.C:
			d.logStats()

		case <-d.stopCh:
			d.logger.Info("stopping event dispatcher")
			return
		}
	}
}

// Stop stops the dispatcher
func (d *Dispatcher) Stop() {
	d.stopOnce.Do(func() {
		close(d.stopCh)
	})
}

// processEvent processes a single state event
func (d *Dispatcher) processEvent(event types.StreamEvent) {
	d.logger.Debug("processing event",
		"module", event.Module,
		"type", event.EventType,
		"address", event.Address,
		"secondary_address", event.SecondaryAddress,
		"denom", event.Denom,
		"block_height", event.BlockHeight)

	// Set block height from current context if available
	if event.BlockHeight == 0 {
		// TODO: Get current block height from context when available
		event.BlockHeight = time.Now().Unix() // Temporary fallback
	}

	// Fan out to subscribers - the registry will handle finding matching subscriptions
	d.registry.FanOut(event, event)
}

// logStats logs subscription statistics
func (d *Dispatcher) logStats() {
	stats := d.registry.GetStats()
	if stats["total"] > 0 {
		d.logger.Info("subscription stats", "stats", stats)
	}
}
