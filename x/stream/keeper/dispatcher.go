package keeper

import (
	"sync"
	"time"

	"cosmossdk.io/log"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// Dispatcher handles the routing of state events to subscriptions
type Dispatcher struct {
	intake      <-chan types.StreamEvent
	registry    *SubscriptionRegistry
	logger      log.Logger
	stopCh      chan struct{}
	stopOnce    sync.Once
	stopped     chan struct{} // Signals when dispatcher has stopped
	stoppedOnce sync.Once     // Ensures stopped channel is only closed once

	// Backpressure metrics
	droppedEvents uint64
}

// NewDispatcher creates a new event dispatcher
func NewDispatcher(intake <-chan types.StreamEvent, registry *SubscriptionRegistry, logger log.Logger) *Dispatcher {
	return &Dispatcher{
		intake:   intake,
		registry: registry,
		logger:   logger.With("component", "dispatcher"),
		stopCh:   make(chan struct{}),
		stopped:  make(chan struct{}),
	}
}

// Start begins the dispatcher event loop
func (d *Dispatcher) Start() {
	// panic recovery to prevent any panic from crashing the node
	defer func() {
		if r := recover(); r != nil {
			d.logger.Error("panic recovered in dispatcher", "panic", r)
			// Ensure stopped channel is closed
			d.stoppedOnce.Do(func() {
				close(d.stopped)
			})
		}
	}()

	d.logger.Info("starting event dispatcher")

	// Stats ticker for metrics
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-d.intake:
			if !ok {
				d.logger.Info("intake channel closed, stopping dispatcher")
				// Ensure stopped channel is closed
				d.stoppedOnce.Do(func() {
					close(d.stopped)
				})
				return
			}
			d.processEvent(event)

		case <-ticker.C:
			d.logStats()
			d.registry.UpdateMetrics()

		case <-d.stopCh:
			d.logger.Info("stopping event dispatcher")
			// Close all active subscriptions gracefully
			d.registry.CloseAll()
			// Close stopped channel only once
			d.stoppedOnce.Do(func() {
				close(d.stopped)
			})
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

// WaitForStop waits for the dispatcher to fully stop
func (d *Dispatcher) WaitForStop() {
	<-d.stopped
}

// processEvent processes a single state event
func (d *Dispatcher) processEvent(event types.StreamEvent) {
	// Add panic recovery for individual event processing
	defer func() {
		if r := recover(); r != nil {
			d.logger.Error("panic recovered while processing event",
				"panic", r,
				"event", event)
		}
	}()

	d.logger.Debug("processing event",
		"module", event.Module,
		"type", event.EventType,
		"address", event.Address,
		"secondary_address", event.SecondaryAddress,
		"denom", event.Denom,
		"block_height", event.BlockHeight)

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
