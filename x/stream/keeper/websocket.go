package keeper

import (
	"context"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

const (
	// WebSocket configuration
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development - should be restricted in production
		return true
	},
}

// getQueryContextOrSendError gets the query context or sends an error message and returns false
func (k *Keeper) getQueryContextOrSendError(conn *websocket.Conn) (context.Context, bool) {
	queryCtx, err := k.GetQueryContext()
	if err != nil {
		k.logger.Error("failed to get query context", "error", err)
		errorMsg := map[string]string{"error": "internal server error"}
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if sendErr := conn.WriteJSON(errorMsg); sendErr != nil {
			k.logger.Error("failed to send error message", "error", sendErr)
		}
		return nil, false
	}
	return queryCtx, true
}

// HandleBalanceSubscription handles balance subscription WebSocket connections
func (k *Keeper) HandleBalanceSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	denom := vars["denom"]

	// Validate address
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !k.connectionManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	if !k.connectionManager.RegisterConnection(remoteAddr) {
		conn.Close()
		return
	}
	defer func() {
		k.connectionManager.UnregisterConnection(remoteAddr)
		conn.Close()
	}()

	// Add subscription to this connection
	if !k.connectionManager.AddSubscription(remoteAddr) {
		return
	}
	defer k.connectionManager.RemoveSubscription(remoteAddr)

	ctx, cancel := context.WithCancel(k.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := k.getQueryContextOrSendError(conn)
	if !ok {
		return
	}

	// Send initial balance
	balance := k.bankKeeper.GetBalance(queryCtx, sdk.MustAccAddressFromBech32(address), denom)
	if err := k.sendWebSocketMessage(conn, balance); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, address, "", denom)
	sendCh := make(chan any, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() any {
		queryCtx, err := k.GetQueryContext()
		if err != nil {
			return map[string]string{"error": "failed to get context"}
		}
		return k.bankKeeper.GetBalance(queryCtx, sdk.MustAccAddressFromBech32(address), denom)
	})
}

// HandleAllBalancesSubscription handles all balances subscription WebSocket connections
func (k *Keeper) HandleAllBalancesSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !k.connectionManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	if !k.connectionManager.RegisterConnection(remoteAddr) {
		conn.Close()
		return
	}
	defer func() {
		k.connectionManager.UnregisterConnection(remoteAddr)
		conn.Close()
	}()

	// Add subscription to this connection
	if !k.connectionManager.AddSubscription(remoteAddr) {
		return
	}
	defer k.connectionManager.RemoveSubscription(remoteAddr)

	ctx, cancel := context.WithCancel(k.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := k.getQueryContextOrSendError(conn)
	if !ok {
		return
	}

	// Send initial balances
	balances := k.bankKeeper.GetAllBalances(queryCtx, addr)
	if err := k.sendWebSocketMessage(conn, map[string]any{"balances": balances}); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, address, "", "")
	sendCh := make(chan any, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() any {
		queryCtx, err := k.GetQueryContext()
		if err != nil {
			return map[string]string{"error": "failed to get context"}
		}
		return map[string]any{"balances": k.bankKeeper.GetAllBalances(queryCtx, addr)}
	})
}

// HandleDelegationsSubscription handles delegations subscription WebSocket connections
func (k *Keeper) HandleDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !k.connectionManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	if !k.connectionManager.RegisterConnection(remoteAddr) {
		conn.Close()
		return
	}
	defer func() {
		k.connectionManager.UnregisterConnection(remoteAddr)
		conn.Close()
	}()

	// Add subscription to this connection
	if !k.connectionManager.AddSubscription(remoteAddr) {
		return
	}
	defer k.connectionManager.RemoveSubscription(remoteAddr)

	ctx, cancel := context.WithCancel(k.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := k.getQueryContextOrSendError(conn)
	if !ok {
		return
	}

	// Send initial delegations
	delegations := k.getDelegationResponses(queryCtx, delAddr)
	if err := k.sendWebSocketMessage(conn, delegations); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, delegatorAddress, "", "")
	sendCh := make(chan any, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() any {
		queryCtx, err := k.GetQueryContext()
		if err != nil {
			return []stakingtypes.DelegationResponse{}
		}
		return k.getDelegationResponses(queryCtx, delAddr)
	})
}

// HandleDelegationSubscription handles delegation subscription WebSocket connections
func (k *Keeper) HandleDelegationSubscription(w http.ResponseWriter, r *http.Request) {
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
	if !k.connectionManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	if !k.connectionManager.RegisterConnection(remoteAddr) {
		conn.Close()
		return
	}
	defer func() {
		k.connectionManager.UnregisterConnection(remoteAddr)
		conn.Close()
	}()

	// Add subscription to this connection
	if !k.connectionManager.AddSubscription(remoteAddr) {
		return
	}
	defer k.connectionManager.RemoveSubscription(remoteAddr)

	ctx, cancel := context.WithCancel(k.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := k.getQueryContextOrSendError(conn)
	if !ok {
		return
	}

	// Send initial delegation
	delegation := k.getDelegationResponse(queryCtx, delAddr, valAddr)
	if err := k.sendWebSocketMessage(conn, delegation); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, delegatorAddress, validatorAddress, "")
	sendCh := make(chan any, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() any {
		queryCtx, err := k.GetQueryContext()
		if err != nil {
			return map[string]any{"found": false}
		}
		return k.getDelegationResponse(queryCtx, delAddr, valAddr)
	})
}

// HandleUnbondingDelegationsSubscription handles unbonding delegations subscription WebSocket connections
func (k *Keeper) HandleUnbondingDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}

	// Check connection limits
	if !k.connectionManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	if !k.connectionManager.RegisterConnection(remoteAddr) {
		conn.Close()
		return
	}
	defer func() {
		k.connectionManager.UnregisterConnection(remoteAddr)
		conn.Close()
	}()

	// Add subscription to this connection
	if !k.connectionManager.AddSubscription(remoteAddr) {
		return
	}
	defer k.connectionManager.RemoveSubscription(remoteAddr)

	ctx, cancel := context.WithCancel(k.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := k.getQueryContextOrSendError(conn)
	if !ok {
		return
	}

	// Send initial unbonding delegations
	unbondingDelegations, err := k.stakingKeeper.GetAllUnbondingDelegations(queryCtx, delAddr)
	if err != nil {
		k.logger.Error("failed to get unbonding delegations", "error", err)
		unbondingDelegations = []stakingtypes.UnbondingDelegation{}
	}
	if err := k.sendWebSocketMessage(conn, unbondingDelegations); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, delegatorAddress, "", "")
	sendCh := make(chan any, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() any {
		queryCtx, err := k.GetQueryContext()
		if err != nil {
			return []stakingtypes.UnbondingDelegation{}
		}
		unbondingDelegations, err := k.stakingKeeper.GetAllUnbondingDelegations(queryCtx, delAddr)
		if err != nil {
			return []stakingtypes.UnbondingDelegation{}
		}
		return unbondingDelegations
	})
}

// HandleUnbondingDelegationSubscription handles unbonding delegation subscription WebSocket connections
func (k *Keeper) HandleUnbondingDelegationSubscription(w http.ResponseWriter, r *http.Request) {
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
	if !k.connectionManager.CheckConnectionLimits(w, r) {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Register connection
	remoteAddr := r.RemoteAddr
	if !k.connectionManager.RegisterConnection(remoteAddr) {
		conn.Close()
		return
	}
	defer func() {
		k.connectionManager.UnregisterConnection(remoteAddr)
		conn.Close()
	}()

	// Add subscription to this connection
	if !k.connectionManager.AddSubscription(remoteAddr) {
		return
	}
	defer k.connectionManager.RemoveSubscription(remoteAddr)

	ctx, cancel := context.WithCancel(k.appContext)
	defer cancel()

	// Get query context with proper SDK context
	queryCtx, ok := k.getQueryContextOrSendError(conn)
	if !ok {
		return
	}

	// Send initial unbonding delegation
	unbondingDelegation, err := k.stakingKeeper.GetUnbondingDelegation(queryCtx, delAddr, valAddr)
	data := map[string]any{"found": err == nil}
	if err == nil {
		data["unbonding_delegation"] = unbondingDelegation
	}
	if err := k.sendWebSocketMessage(conn, data); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, delegatorAddress, validatorAddress, "")
	sendCh := make(chan any, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() any {
		queryCtx, err := k.GetQueryContext()
		if err != nil {
			return map[string]any{"found": false}
		}
		unbondingDelegation, err := k.stakingKeeper.GetUnbondingDelegation(queryCtx, delAddr, valAddr)
		data := map[string]any{"found": err == nil}
		if err == nil {
			data["unbonding_delegation"] = unbondingDelegation
		}
		return data
	})
}

// handleWebSocketConnection handles the WebSocket connection lifecycle
func (k *Keeper) handleWebSocketConnection(conn *websocket.Conn, ctx context.Context, sendCh <-chan any, queryFunc func() any) {
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	// Ensure we send a close message on exit
	defer func() {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"))
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sendCh:
			// Re-query data and send update
			data := queryFunc()
			if err := k.sendWebSocketMessage(conn, data); err != nil {
				k.logger.Error("failed to send websocket message", "error", err)
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendWebSocketMessage sends a message over WebSocket
func (k *Keeper) sendWebSocketMessage(conn *websocket.Conn, data any) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	err := conn.WriteJSON(data)
	if err == nil {
		IncrementMessagesSent()
	}
	return err
}

// getDelegationResponses gets delegation responses for a delegator
func (k *Keeper) getDelegationResponses(ctx context.Context, delAddr sdk.AccAddress) []stakingtypes.DelegationResponse {
	delegations, err := k.stakingKeeper.GetAllDelegatorDelegations(ctx, delAddr)
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	var delegationResponses []stakingtypes.DelegationResponse
	for _, delegation := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			continue
		}

		validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
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
func (k *Keeper) getDelegationResponse(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) any {
	delegation, err := k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)

	data := map[string]any{"found": err == nil}
	if err == nil {
		bondDenom, bondErr := k.stakingKeeper.BondDenom(ctx)
		if bondErr != nil {
			return data
		}

		validator, valErr := k.stakingKeeper.GetValidator(ctx, valAddr)
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
