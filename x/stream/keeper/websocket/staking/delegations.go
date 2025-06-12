package staking

import (
	"context"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
)

// DelegationsHandler handles delegations subscription WebSocket connections
type DelegationsHandler struct {
	keeper         KeeperInterface
	config         *common.StreamConfig
	logger         common.Logger
	connManager    common.ConnectionManager
	registry       common.SubscriptionRegistry
	circuitBreaker common.CircuitBreaker
	appContext     context.Context
	upgrader       *websocket.Upgrader
}

// NewDelegationsHandler creates a new delegations handler
func NewDelegationsHandler(keeper KeeperInterface, config *common.StreamConfig, logger common.Logger, connManager common.ConnectionManager, registry common.SubscriptionRegistry, circuitBreaker common.CircuitBreaker, appContext context.Context, upgrader *websocket.Upgrader) *DelegationsHandler {
	return &DelegationsHandler{
		keeper:         keeper,
		config:         config,
		logger:         logger,
		connManager:    connManager,
		registry:       registry,
		circuitBreaker: circuitBreaker,
		appContext:     appContext,
		upgrader:       upgrader,
	}
}

// Handle handles delegations subscription WebSocket connections
func (h *DelegationsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !h.connManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	connectionID := h.connManager.RegisterConnectionWithHeaders(remoteAddr, xForwardedFor)
	if connectionID == "" {
		conn.Close()
		return
	}

	// Check circuit breaker if enabled
	if !checkCircuitBreaker(h.config, h.circuitBreaker, h.logger, h.connManager, conn, connectionID) {
		return
	}
	defer func() {
		h.connManager.UnregisterConnection(connectionID)
		if err := conn.Close(); err != nil {
			h.logger.Error("failed to close websocket connection", "error", err, "connection_id", connectionID)
		}
	}()

	// Add subscription to this connection
	if !h.connManager.AddSubscription(connectionID) {
		return
	}
	defer h.connManager.RemoveSubscription(connectionID)

	ctx, cancel := context.WithCancel(h.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := getQueryContextOrSendError(h.keeper, h.logger, conn)
	if !ok {
		return
	}

	// Send initial delegations
	delegations := getDelegationResponses(h.keeper, queryCtx, delAddr)
	if err := sendWebSocketMessage(conn, delegations); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, delegatorAddress, "", "")
	sendCh := make(chan any, h.config.SubscriptionBufferSize)
	subscriber := h.registry.Subscribe(common.NewSubscriptionKeyAdapter(subKey), ctx, sendCh)
	defer h.registry.Unsubscribe(subscriber)

	handleWebSocketConnection(h.config, h.circuitBreaker, h.logger, conn, ctx, sendCh, connectionID, func() any {
		queryCtx, err := h.keeper.GetQueryContext()
		if err != nil {
			return []stakingtypes.DelegationResponse{}
		}
		return getDelegationResponses(h.keeper, queryCtx, delAddr)
	})
}

// HandleDelegation handles delegation subscription WebSocket connections
func (h *DelegationsHandler) HandleDelegation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]
	validatorAddress := vars["validator"]

	// Validate addresses
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}
	valAddr, err := sdk.ValAddressFromBech32(validatorAddress)
	if err != nil {
		http.Error(w, "invalid validator address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !h.connManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	connectionID := h.connManager.RegisterConnectionWithHeaders(remoteAddr, xForwardedFor)
	if connectionID == "" {
		conn.Close()
		return
	}

	// Check circuit breaker if enabled
	if !checkCircuitBreaker(h.config, h.circuitBreaker, h.logger, h.connManager, conn, connectionID) {
		return
	}
	defer func() {
		h.connManager.UnregisterConnection(connectionID)
		if err := conn.Close(); err != nil {
			h.logger.Error("failed to close websocket connection", "error", err, "connection_id", connectionID)
		}
	}()

	// Add subscription to this connection
	if !h.connManager.AddSubscription(connectionID) {
		return
	}
	defer h.connManager.RemoveSubscription(connectionID)

	ctx, cancel := context.WithCancel(h.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := getQueryContextOrSendError(h.keeper, h.logger, conn)
	if !ok {
		return
	}

	// Send initial delegation
	delegation := getDelegationResponse(h.keeper, queryCtx, delAddr, valAddr)
	if err := sendWebSocketMessage(conn, delegation); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, delegatorAddress, validatorAddress, "")
	sendCh := make(chan any, h.config.SubscriptionBufferSize)
	subscriber := h.registry.Subscribe(common.NewSubscriptionKeyAdapter(subKey), ctx, sendCh)
	defer h.registry.Unsubscribe(subscriber)

	handleWebSocketConnection(h.config, h.circuitBreaker, h.logger, conn, ctx, sendCh, connectionID, func() any {
		queryCtx, err := h.keeper.GetQueryContext()
		if err != nil {
			return map[string]any{"found": false}
		}
		return getDelegationResponse(h.keeper, queryCtx, delAddr, valAddr)
	})
}

// Helper functions
func getQueryContextOrSendError(keeper KeeperInterface, logger common.Logger, conn *websocket.Conn) (context.Context, bool) {
	queryCtx, err := keeper.GetQueryContext()
	if err != nil {
		logger.Error("failed to get query context", "error", err)
		errorMsg := map[string]string{"error": "service temporarily unavailable"}
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if sendErr := conn.WriteJSON(errorMsg); sendErr != nil {
			logger.Error("failed to send error message", "error", sendErr)
		}
		return nil, false
	}
	return queryCtx, true
}

func checkCircuitBreaker(config *common.StreamConfig, circuitBreaker common.CircuitBreaker, logger common.Logger, connManager common.ConnectionManager, conn *websocket.Conn, connectionID string) bool {
	if config.CircuitBreakerEnabled && circuitBreaker != nil {
		allowed, err := circuitBreaker.AllowRequest(connectionID)
		if !allowed {
			logger.Warn("circuit breaker blocked request", "connection_id", connectionID, "error", err)
			types.IncrementConnectionRejected("circuit_breaker")
			errorMsg := map[string]string{"error": "service temporarily unavailable"}
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			conn.WriteJSON(errorMsg)
			conn.Close()
			connManager.UnregisterConnection(connectionID)
			return false
		}
	}
	return true
}

func handleWebSocketConnection(config *common.StreamConfig, circuitBreaker common.CircuitBreaker, logger common.Logger, conn *websocket.Conn, ctx context.Context, sendCh <-chan any, connectionID string, queryFunc func() any) {
	common.HandleWebSocketConnection(conn, ctx, sendCh, connectionID, queryFunc,
		func(conn *websocket.Conn, data any) error {
			return sendWebSocketMessage(conn, data)
		},
		func(connectionID string) {
			logger.Error("failed to send websocket message", "connection_id", connectionID)
			if config.CircuitBreakerEnabled && circuitBreaker != nil {
				circuitBreaker.RecordFailure(connectionID)
			}
		},
		func(connectionID string) {
			if config.CircuitBreakerEnabled && circuitBreaker != nil {
				circuitBreaker.RecordSuccess(connectionID)
			}
		})
}

func sendWebSocketMessage(conn *websocket.Conn, data any) error {
	return common.SendWebSocketMessage(conn, data, func() {
		types.IncrementMessagesSent()
	})
}

// getDelegationResponses gets delegation responses for a delegator
func getDelegationResponses(keeper KeeperInterface, ctx context.Context, delAddr sdk.AccAddress) []stakingtypes.DelegationResponse {
	stakingKeeper := keeper.GetStakingKeeper()
	delegations, err := stakingKeeper.GetAllDelegatorDelegations(ctx, delAddr)
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	bondDenom, err := stakingKeeper.BondDenom(ctx)
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	var delegationResponses []stakingtypes.DelegationResponse
	for _, delegation := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			continue
		}

		validator, err := stakingKeeper.GetValidator(ctx, valAddr)
		if err == nil {
			delegationResponses = append(delegationResponses, stakingtypes.DelegationResponse{
				Delegation: delegation,
				Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
			})
		}
	}
	return delegationResponses
}

// getDelegationResponse gets a delegation response for a specific delegator-validator pair
func getDelegationResponse(keeper KeeperInterface, ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) any {
	stakingKeeper := keeper.GetStakingKeeper()
	delegation, err := stakingKeeper.GetDelegation(ctx, delAddr, valAddr)

	data := map[string]any{"found": err == nil}
	if err == nil {
		bondDenom, bondErr := stakingKeeper.BondDenom(ctx)
		if bondErr != nil {
			return data
		}

		validator, valErr := stakingKeeper.GetValidator(ctx, valAddr)
		if valErr == nil {
			delegationResponse := stakingtypes.DelegationResponse{
				Delegation: delegation,
				Balance:    sdk.NewCoin(bondDenom, validator.TokensFromShares(delegation.Shares).TruncateInt()),
			}
			data["delegation"] = delegationResponse
		}
	}
	return data
}