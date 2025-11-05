package keeper

import (
	"context"
	"sync"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/schema"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/CosmosContracts/juno/v30/x/stream/types/encoding"
)

// Keeper defines the stream module keeper
type Keeper struct {
	baseApp *baseapp.BaseApp

	// State listening components
	registry       *types.SubscriptionRegistry
	dispatcher     *types.Dispatcher
	invoker        *types.RouterInvoker
	methodRegistry *encoding.DynamicRegistry

	// App lifecycle signalling
	appDone   <-chan struct{}
	appCancel context.CancelFunc
	stopOnce  sync.Once

	moduleCodecs map[string]schema.ModuleCodec

	logger log.Logger
}

// NewKeeper creates a new stream keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	homePath string,
	baseApp *baseapp.BaseApp,
) *Keeper {
	logger := baseApp.Logger()

	streamCfg, err := types.LoadStreamConfig(homePath)
	if err != nil {
		logger.Error("failed to load stream config, using defaults", "error", err)
	}

	registry := types.NewSubscriptionRegistry(logger, streamCfg)
	dispatcher := types.NewDispatcher(int(streamCfg.SubscriptionBufferSize), registry, logger)

	invoker, err := types.NewRouterInvoker(baseApp, cdc)
	if err != nil {
		logger.Error("create router invoker: %w", err)
	}

	methodRegistry, err := encoding.NewDynamicRegistry(baseApp)
	if err != nil {
		logger.Error("create dynamic method registry: %w", err)
	}

	appCtx, appCancel := context.WithCancel(context.Background())

	k := &Keeper{
		baseApp:        baseApp,
		registry:       registry,
		dispatcher:     dispatcher,
		invoker:        invoker,
		methodRegistry: methodRegistry,
		appDone:        appCtx.Done(),
		appCancel:      appCancel,
		moduleCodecs:   make(map[string]schema.ModuleCodec),
		logger:         logger.With("module", "x/stream"),
	}

	if err := k.methodRegistry.Refresh(k.baseApp); err != nil {
		k.logger.Error("failed to refresh method registry", "error", err)
	}

	return k
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger() log.Logger {
	return k.logger
}

// Dispatcher returns the event dispatcher for the stream module.
func (k *Keeper) Dispatcher() *types.Dispatcher {
	return k.dispatcher
}

// TODO: we need to wait until ALL cosmos sdk and juno modules FULLY implement SDK collections
// until we can simplify the state decode by A LOT. Keeping this here for future reference.
//
// // RegisterModuleSchema builds and registers a module codec for the provided store key.
// func (k *Keeper) RegisterModuleSchema(storeKey string, schema collections.Schema, opts collections.IndexingOptions) error {
// 	if storeKey == "" {
// 		return fmt.Errorf("store key cannot be empty")
// 	}

// 	moduleCodec, err := schema.ModuleCodec(opts)
// 	if err != nil {
// 		return fmt.Errorf("build module codec for store %s: %w", storeKey, err)
// 	}

// 	k.moduleCodecs[storeKey] = moduleCodec
// 	return nil
// }

// // ModuleCodecs returns a snapshot copy of the registered module codecs keyed by store name.
// func (k *Keeper) ModuleCodecs() map[string]schema.ModuleCodec {
// 	out := make(map[string]schema.ModuleCodec, len(k.moduleCodecs))
// 	maps.Copy(out, k.moduleCodecs)
// 	return out
// }

// StartDispatcher starts the event dispatcher goroutine
func (k *Keeper) StartDispatcher() {
	k.dispatcher.Start(newDoneContext(k.appDone))
	k.logger.Info("stream dispatcher started")
}

// StopDispatcher stops the event dispatcher
func (k *Keeper) StopDispatcher() {
	if k == nil {
		// This shouldn't happen, but let's be defensive
		return
	}

	k.stopOnce.Do(func() {
		k.logger.Info("stopping stream dispatcher", "keeper", k != nil, "dispatcher", k.dispatcher != nil)
		// Cancel app context to signal all streams to stop
		if k.appCancel != nil {
			k.appCancel()
		}
		// Stop the dispatcher
		if k.dispatcher != nil {
			k.dispatcher.Stop()
			k.dispatcher.WaitForStop()
		}
		// Don't close the intake channel here - let the listener handle it
		k.logger.Info("stream dispatcher stopped")
	})
}

// Registry returns the subscription registry
func (k *Keeper) Registry() *types.SubscriptionRegistry {
	return k.registry
}

// MethodRegistry returns the dynamically discovered gRPC method registry.
func (k *Keeper) MethodRegistry() *encoding.DynamicRegistry {
	return k.methodRegistry
}

// Invoker returns the router invoker used by the stream module.
func (k *Keeper) Invoker() *types.RouterInvoker {
	return k.invoker
}

// GetAppContext returns the app context used for lifecycle management
func (k *Keeper) GetAppContext() context.Context {
	return newDoneContext(k.appDone)
}

type doneContext struct {
	done <-chan struct{}
}

func newDoneContext(done <-chan struct{}) context.Context {
	if done == nil {
		return context.Background()
	}
	return doneContext{done: done}
}

func (d doneContext) Deadline() (time.Time, bool) {
	_ = d
	return time.Time{}, false
}

func (d doneContext) Done() <-chan struct{} {
	return d.done
}

func (d doneContext) Err() error {
	if d.done == nil {
		return nil
	}
	select {
	case <-d.done:
		return context.Canceled
	default:
		return nil
	}
}

func (d doneContext) Value(_ any) any {
	_ = d
	return nil
}
