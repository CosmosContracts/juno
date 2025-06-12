package keeper

import (
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

// InitializeWebSocketHandler creates and initializes the WebSocket handler
func (k *Keeper) InitializeWebSocketHandler() {
	// Convert StreamConfig to common.StreamConfig
	config := &common.StreamConfig{
		IntakeBufferSize:        k.config.IntakeBufferSize,
		SubscriptionBufferSize:  k.config.SubscriptionBufferSize,
		EnableConnectionUUID:    k.config.EnableConnectionUUID,
		CircuitBreakerEnabled:   k.config.CircuitBreakerEnabled,
		CircuitBreakerThreshold: k.config.CircuitBreakerThreshold,
		CircuitBreakerTimeout:   k.config.CircuitBreakerTimeout,
	}

	// Create adapters for common interfaces
	loggerAdapter := common.NewLoggerAdapter(k.logger)
	connManagerAdapter := common.NewConnectionManagerAdapter(k.connectionManager)
	registryAdapter := common.NewSubscriptionRegistryAdapter(k.registry)

	// Create the handler using the new CreateHandler function
	k.wsHandler = websocket.CreateHandler(
		k, // Keeper implements websocket.KeeperInterface
		config,
		loggerAdapter,
		connManagerAdapter,
		registryAdapter,
		k.circuitBreaker,
		k.appContext,
		k.allowAllOrigins,
	)
}

// GetBankKeeper returns the bank keeper for use by websocket package
func (k *Keeper) GetBankKeeper() websocket.BankKeeper {
	return k.bankKeeper
}

// GetStakingKeeper returns the staking keeper for use by websocket package
func (k *Keeper) GetStakingKeeper() websocket.StakingKeeper {
	return k.stakingKeeper
}

// WebSocketHandler returns the websocket handler for routing
func (k *Keeper) WebSocketHandler() *websocket.Handler {
	return k.wsHandler
}