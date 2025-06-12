package keeper

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/middleware"
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
	registry   *types.SubscriptionRegistry
	dispatcher *Dispatcher

	// stores the node's context.Context for streaming queries
	// TODO: find a better solution for accessing the main node context
	// in the stream queries since grpc stream servers only have their own
	// context instance. This would also increase the chains performance
	// again because we don't need to run a preblocker before every block.
	queryContext atomic.Value

	// App context for shutdown handling
	appContext context.Context
	appCancel  context.CancelFunc
	stopOnce   sync.Once

	// Connection limits from CometBFT config
	maxConnections            int
	maxSubscriptionsPerClient int
	connectionManager         *ConnectionManager

	// Stream configuration
	config StreamConfig

	// CORS configuration
	allowAllOrigins bool
	corsOrigins     []string

	// Circuit breaker for connection protection
	circuitBreaker *middleware.CircuitBreaker

	// WebSocket handler
	wsHandler *websocket.Handler

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
	maxConnections int,
	maxSubscriptionsPerClient int,
) *Keeper {
	// Start with default config
	config := DefaultStreamConfig()

	// Create buffered intake channel for state events
	intake := make(chan types.StreamEvent, config.IntakeBufferSize)

	// Create subscription registry
	registry := types.NewSubscriptionRegistry(logger)

	// Create dispatcher
	dispatcher := NewDispatcher(intake, registry, logger)

	// Create app context for lifecycle management
	appCtx, appCancel := context.WithCancel(context.Background())

	// Set default values if not provided
	if maxConnections <= 0 {
		maxConnections = 900 // CometBFT default
	}
	if maxSubscriptionsPerClient <= 0 {
		maxSubscriptionsPerClient = 5 // CometBFT default
	}

	// Create connection manager
	connectionManager := NewConnectionManager(maxConnections, maxSubscriptionsPerClient, logger)

	// Create circuit breaker (will be configured when SetStreamConfig is called)
	circuitBreaker := middleware.NewCircuitBreaker(config.CircuitBreakerThreshold, config.CircuitBreakerTimeout)

	k := &Keeper{
		cdc:                       cdc,
		storeKey:                  storeKey,
		authority:                 authority,
		bankKeeper:                bankKeeper,
		stakingKeeper:             stakingKeeper,
		intake:                    intake,
		registry:                  registry,
		dispatcher:                dispatcher,
		appContext:                appCtx,
		appCancel:                 appCancel,
		maxConnections:            maxConnections,
		maxSubscriptionsPerClient: maxSubscriptionsPerClient,
		connectionManager:         connectionManager,
		config:                    config,
		circuitBreaker:            circuitBreaker,
		logger:                    logger.With("module", "x/stream"),
	}

	// Initialize WebSocket handler
	k.InitializeWebSocketHandler()

	return k
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

	// Start periodic cleanup for circuit breaker if enabled
	if k.config.CircuitBreakerEnabled && k.circuitBreaker != nil {
		go k.runCircuitBreakerCleanup()
	}
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
// Returns an error if context is not yet available
func (k *Keeper) GetQueryContext() (context.Context, error) {
	val := k.queryContext.Load()

	if val != nil {
		storedCtx := val.(context.Context)
		select {
		case <-storedCtx.Done():
			k.logger.Warn("stored query context is cancelled")
			return nil, fmt.Errorf("query context is no longer valid")
		default:
			return storedCtx, nil
		}
	}

	// This happens when no block has been processed yet
	k.logger.Debug("no query context available yet")
	return nil, types.ErrNoQueryContext
}

// GetAppContext returns the app context used for lifecycle management
func (k *Keeper) GetAppContext() context.Context {
	return k.appContext
}

// SetConnectionLimits updates the connection limits from config
func (k *Keeper) SetConnectionLimits(maxConnections, maxSubscriptionsPerClient int) {
	if maxConnections > 0 {
		k.maxConnections = maxConnections
		k.connectionManager.maxConnections = int32(maxConnections)
	}
	if maxSubscriptionsPerClient > 0 {
		k.maxSubscriptionsPerClient = maxSubscriptionsPerClient
		k.connectionManager.maxSubscriptionsPerClient = int32(maxSubscriptionsPerClient)
	}
	k.logger.Info("connection limits updated",
		"max_connections", k.maxConnections,
		"max_subscriptions_per_client", k.maxSubscriptionsPerClient)
}

// SetAllowAllOrigins sets whether to allow all origins for WebSocket connections
func (k *Keeper) SetAllowAllOrigins(allow bool) {
	k.allowAllOrigins = allow
	k.logger.Info("CORS configuration updated", "allow_all_origins", allow)

	// Reinitialize WebSocket handler with new CORS settings
	k.InitializeWebSocketHandler()
}

// SetStreamConfig updates the stream configuration
func (k *Keeper) SetStreamConfig(config StreamConfig) error {
	// Apply defaults to fill any zero values
	config.ApplyDefaults()

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid stream config: %w", err)
	}

	k.config = config

	// Recreate intake channel with new buffer size if it changed
	if cap(k.intake) != config.IntakeBufferSize {
		// Note: This is safe only during initialization before the dispatcher starts
		k.intake = make(chan types.StreamEvent, config.IntakeBufferSize)
		k.dispatcher = NewDispatcher(k.intake, k.registry, k.logger)
	}

	// Update connection manager settings
	if k.connectionManager != nil {
		k.connectionManager.SetEnableUUID(config.EnableConnectionUUID)
	}

	// Update circuit breaker configuration
	if config.CircuitBreakerEnabled && k.circuitBreaker == nil {
		k.circuitBreaker = middleware.NewCircuitBreaker(config.CircuitBreakerThreshold, config.CircuitBreakerTimeout)
	} else if config.CircuitBreakerEnabled && k.circuitBreaker != nil {
		// Update existing circuit breaker settings
		k.circuitBreaker.UpdateThreshold(config.CircuitBreakerThreshold)
		k.circuitBreaker.UpdateTimeout(config.CircuitBreakerTimeout)
	} else if !config.CircuitBreakerEnabled {
		// Disable circuit breaker
		k.circuitBreaker = nil
	}

	k.logger.Info("stream configuration updated",
		"intake_buffer_size", config.IntakeBufferSize,
		"subscription_buffer_size", config.SubscriptionBufferSize,
		"enable_connection_uuid", config.EnableConnectionUUID,
		"circuit_breaker_enabled", config.CircuitBreakerEnabled)

	// Reinitialize WebSocket handler with new config
	k.InitializeWebSocketHandler()

	return nil
}

// SetCORSOrigins updates the allowed CORS origins
func (k *Keeper) SetCORSOrigins(origins []string) {
	k.corsOrigins = origins
	k.allowAllOrigins = len(origins) == 1 && origins[0] == "*"
	k.logger.Info("CORS origins updated", "origins", origins, "allow_all", k.allowAllOrigins)
}

// GetConfig returns the current stream configuration
func (k *Keeper) GetConfig() StreamConfig {
	return k.config
}

// ValidateDenom validates if a denom is valid for streaming
func (k *Keeper) ValidateDenom(ctx context.Context, denom string) error {
	if denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}

	// Check if it's a valid denom format
	if err := sdk.ValidateDenom(denom); err != nil {
		return fmt.Errorf("invalid denom format: %w", err)
	}

	// Optionally check if the denom exists in bank module metadata
	// This is a more strict validation but might be too restrictive
	// as not all denoms have metadata

	return nil
}

// GetCircuitBreakerMetrics returns circuit breaker metrics for monitoring
func (k *Keeper) GetCircuitBreakerMetrics() map[string]interface{} {
	if k.circuitBreaker == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	metrics := k.circuitBreaker.GetMetrics()
	metrics["enabled"] = true
	return metrics
}

// runCircuitBreakerCleanup periodically cleans up stale circuit breaker entries
func (k *Keeper) runCircuitBreakerCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if k.circuitBreaker != nil && k.connectionManager != nil {
				activeConnections := k.connectionManager.GetActiveConnections()
				k.circuitBreaker.CleanupStaleConnections(activeConnections)

				// Update metrics
				metrics := k.circuitBreaker.GetMetrics()
				if openCircuits, ok := metrics["open_circuits"].(int); ok {
					if halfOpenCircuits, ok2 := metrics["half_open_circuits"].(int); ok2 {
						if closedCircuits, ok3 := metrics["closed_circuits"].(int); ok3 {
							types.UpdateCircuitBreakerMetrics(openCircuits, halfOpenCircuits, closedCircuits)
						}
					}
				}

				k.logger.Debug("circuit breaker cleanup completed", "active_connections", len(activeConnections))
			}
		case <-k.appContext.Done():
			k.logger.Info("stopping circuit breaker cleanup")
			return
		}
	}
}
