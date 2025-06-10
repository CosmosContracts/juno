package keeper

import (
	"context"
	"sync/atomic"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
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

	// Context for streaming queries - updated each block
	queryContext atomic.Value // stores context.Context

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

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		authority:     authority,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		intake:        intake,
		registry:      registry,
		dispatcher:    dispatcher,
		logger:        logger.With("module", "x/stream"),
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger
}

// Intake returns the intake channel for the state listener
func (k Keeper) Intake() chan<- types.StreamEvent {
	return k.intake
}

// StartDispatcher starts the event dispatcher goroutine
func (k Keeper) StartDispatcher() {
	go k.dispatcher.Start()
	k.logger.Info("stream dispatcher started")
}

// StopDispatcher stops the event dispatcher
func (k Keeper) StopDispatcher() {
	k.dispatcher.Stop()
	k.logger.Info("stream dispatcher stopped")
}

// Registry returns the subscription registry
func (k Keeper) Registry() *SubscriptionRegistry {
	return k.registry
}

// SetQueryContext updates the context used for streaming queries
// This should be called at the beginning of each block
func (k *Keeper) SetQueryContext(ctx context.Context) {
	k.queryContext.Store(ctx)
	k.logger.Info("query context updated", "height", ctx.Value("height"))
}

// GetQueryContext returns the current query context
// Falls back to a background context if not set
func (k *Keeper) GetQueryContext() (context.Context, error) {
	if ctx := k.queryContext.Load(); ctx != nil {
		return ctx.(context.Context), nil
	}
	// This should not happen in practice, but provide a fallback
	return context.Background(), nil
}
