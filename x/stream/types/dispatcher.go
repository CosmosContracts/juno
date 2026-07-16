package types

import (
	"context"
	"sync"

	"cosmossdk.io/log"

	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

// Dispatcher handles the routing of state events to subscriptions
type Dispatcher struct {
	intake   chan encoding.StreamEvent
	registry *SubscriptionRegistry
	logger   log.Logger

	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewDispatcher creates a new event dispatcher
func NewDispatcher(bufferSize int, registry *SubscriptionRegistry, logger log.Logger) *Dispatcher {
	return &Dispatcher{
		intake:   make(chan encoding.StreamEvent, bufferSize),
		registry: registry,
		logger:   logger.With("component", "dispatcher"),
	}
}

// Intake returns the intake channel for producers to publish events.
func (d *Dispatcher) Intake() chan<- encoding.StreamEvent {
	return d.intake
}

// IntakeCapacity returns the capacity of the intake channel buffer.
func (d *Dispatcher) IntakeCapacity() int {
	return cap(d.intake)
}

// Start begins the dispatcher event loop using the provided context.
func (d *Dispatcher) Start(ctx context.Context) {
	d.startOnce.Do(func() {
		runCtx := ctx
		if runCtx == nil {
			runCtx = context.Background()
		}
		runCtx, cancel := context.WithCancel(runCtx)
		d.cancel = cancel

		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			d.run(runCtx)
		}()

		d.logger.Debug("event dispatcher started")
	})
}

// Stop stops the dispatcher
func (d *Dispatcher) Stop() {
	d.stopOnce.Do(func() {
		if d.cancel != nil {
			d.cancel()
		}
	})
}

// WaitForStop waits for the dispatcher to fully stop
func (d *Dispatcher) WaitForStop() {
	d.wg.Wait()
}

func (d *Dispatcher) run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			d.logger.Error("panic recovered in dispatcher", "panic", r)
		}
		d.registry.CloseAll()
		d.logger.Debug("event dispatcher stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-d.intake:
			if !ok {
				return
			}
			d.processEvent(event)
		}
	}
}

// processEvent processes a single state event
func (d *Dispatcher) processEvent(event encoding.StreamEvent) {
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
		"method", event.Method,
		"params", event.Params)

	// Temp solution to refresh all subs each block until we can better support event decoding.
	if event.Module == "new_block" && event.Method == "temp" {
		d.registry.Broadcast(event)
		return
	}

	// Fan out to subscribers, the registry will handle finding matching subscriptions
	d.registry.FanOut(event, event)
}
