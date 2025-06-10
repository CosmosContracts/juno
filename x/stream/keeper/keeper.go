package keeper

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

// Keeper defines the stream module keeper
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	authority string

	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper

	// State listening components
	intake     chan types.StreamEvent
	registry   *SubscriptionRegistry
	dispatcher *Dispatcher

	queryContext atomic.Value // stores context.Context

	// App context for shutdown handling
	appContext context.Context
	appCancel  context.CancelFunc
	stopOnce   sync.Once

	logger log.Logger
}

// NewKeeper creates a new stream keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
	bankKeeper bankkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
	logger log.Logger,
) *Keeper {
	// Create buffered intake channel for state events
	intake := make(chan types.StreamEvent, 10000)

	// Create subscription registry
	registry := NewSubscriptionRegistry(logger)

	// Create dispatcher
	dispatcher := NewDispatcher(intake, registry, logger)

	// Create app context for lifecycle management
	appCtx, appCancel := context.WithCancel(context.Background())

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		authority:     authority,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		intake:        intake,
		registry:      registry,
		dispatcher:    dispatcher,
		appContext:    appCtx,
		appCancel:     appCancel,
		logger:        logger.With("module", "x/stream"),
	}
}

// GetAuthority returns the module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger() log.Logger {
	return k.logger
}

// Intake returns the intake channel for the state listener
func (k *Keeper) Intake() chan<- types.StreamEvent {
	return k.intake
}

// StartDispatcher starts the event dispatcher goroutine
func (k *Keeper) StartDispatcher() {
	go k.dispatcher.Start()
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
func (k *Keeper) Registry() *SubscriptionRegistry {
	return k.registry
}

// SetQueryContext updates the context used for streaming queries
// This should be called at the beginning of each block
func (k *Keeper) SetQueryContext(ctx context.Context) {
	if sdkCtx, ok := ctx.(sdk.Context); ok {
		k.queryContext.Store(ctx)
		k.logger.Info("query context updated with SDK context",
			"height", sdkCtx.BlockHeight(),
			"has_multistore", sdkCtx.MultiStore() != nil)
	}
}

// GetQueryContext returns the current query context
// Falls back to a background context if not set
func (k *Keeper) GetQueryContext() (context.Context, error) {
	val := k.queryContext.Load()

	if val != nil {
		storedCtx := val.(context.Context)
		return storedCtx, nil
	}

	// This happens when no block has been processed yet
	k.logger.Error("no query context available - PreBlocker hasn't run yet")
	return nil, fmt.Errorf("query context not initialized - no blocks processed yet")
}

// GetAppContext returns the app context used for lifecycle management
func (k *Keeper) GetAppContext() context.Context {
	return k.appContext
}
