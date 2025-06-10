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

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// StartWebSocketServer starts the WebSocket server
func (k *Keeper) StartWebSocketServer(addr string) error {
	router := mux.NewRouter()

	// Bank endpoints
	router.HandleFunc("/subscribe/bank/balance/{address}/{denom}", k.handleBalanceSubscription)
	router.HandleFunc("/subscribe/bank/balances/{address}", k.handleAllBalancesSubscription)

	// Staking endpoints
	router.HandleFunc("/subscribe/staking/delegations/{delegator}", k.handleDelegationsSubscription)
	router.HandleFunc("/subscribe/staking/delegation/{delegator}/{validator}", k.handleDelegationSubscription)
	router.HandleFunc("/subscribe/staking/unbonding-delegations/{delegator}", k.handleUnbondingDelegationsSubscription)
	router.HandleFunc("/subscribe/staking/unbonding-delegation/{delegator}/{validator}", k.handleUnbondingDelegationSubscription)

	k.logger.Info("starting WebSocket server", "address", addr)
	return http.ListenAndServe(addr, router)
}

// handleBalanceSubscription handles balance subscription WebSocket connections
func (k *Keeper) handleBalanceSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	denom := vars["denom"]

	// Validate address
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send initial balance
	balance := k.bankKeeper.GetBalance(sdk.UnwrapSDKContext(ctx), sdk.MustAccAddressFromBech32(address), denom)
	if err := k.sendWebSocketMessage(conn, "balance", balance); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, address, "", denom)
	sendCh := make(chan interface{}, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() interface{} {
		return k.bankKeeper.GetBalance(sdk.UnwrapSDKContext(ctx), sdk.MustAccAddressFromBech32(address), denom)
	})
}

// handleAllBalancesSubscription handles all balances subscription WebSocket connections
func (k *Keeper) handleAllBalancesSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send initial balances
	balances := k.bankKeeper.GetAllBalances(sdk.UnwrapSDKContext(ctx), addr)
	if err := k.sendWebSocketMessage(conn, "balances", balances); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, address, "", "")
	sendCh := make(chan interface{}, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() interface{} {
		return k.bankKeeper.GetAllBalances(sdk.UnwrapSDKContext(ctx), addr)
	})
}

// handleDelegationsSubscription handles delegations subscription WebSocket connections
func (k *Keeper) handleDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send initial delegations
	delegations := k.getDelegationResponses(delAddr)
	if err := k.sendWebSocketMessage(conn, "delegations", delegations); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, delegatorAddress, "", "")
	sendCh := make(chan interface{}, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() interface{} {
		return k.getDelegationResponses(delAddr)
	})
}

// handleDelegationSubscription handles delegation subscription WebSocket connections
func (k *Keeper) handleDelegationSubscription(w http.ResponseWriter, r *http.Request) {
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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send initial delegation
	delegation := k.getDelegationResponse(delAddr, valAddr)
	if err := k.sendWebSocketMessage(conn, "delegation", delegation); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, delegatorAddress, validatorAddress, "")
	sendCh := make(chan interface{}, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() interface{} {
		return k.getDelegationResponse(delAddr, valAddr)
	})
}

// handleUnbondingDelegationsSubscription handles unbonding delegations subscription WebSocket connections
func (k *Keeper) handleUnbondingDelegationsSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delegatorAddress := vars["delegator"]

	// Validate address
	delAddr, err := sdk.AccAddressFromBech32(delegatorAddress)
	if err != nil {
		http.Error(w, "invalid delegator address", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send initial unbonding delegations
	unbondingDelegations, err := k.stakingKeeper.GetAllUnbondingDelegations(sdk.UnwrapSDKContext(ctx), delAddr)
	if err != nil {
		k.logger.Error("failed to get unbonding delegations", "error", err)
		unbondingDelegations = []stakingtypes.UnbondingDelegation{}
	}
	if err := k.sendWebSocketMessage(conn, "unbonding_delegations", unbondingDelegations); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, delegatorAddress, "", "")
	sendCh := make(chan interface{}, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() interface{} {
		unbondingDelegations, err := k.stakingKeeper.GetAllUnbondingDelegations(sdk.UnwrapSDKContext(ctx), delAddr)
		if err != nil {
			return []stakingtypes.UnbondingDelegation{}
		}
		return unbondingDelegations
	})
}

// handleUnbondingDelegationSubscription handles unbonding delegation subscription WebSocket connections
func (k *Keeper) handleUnbondingDelegationSubscription(w http.ResponseWriter, r *http.Request) {
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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		k.logger.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send initial unbonding delegation
	unbondingDelegation, err := k.stakingKeeper.GetUnbondingDelegation(sdk.UnwrapSDKContext(ctx), delAddr, valAddr)
	data := map[string]interface{}{"found": err == nil}
	if err == nil {
		data["unbonding_delegation"] = unbondingDelegation
	}
	if err := k.sendWebSocketMessage(conn, "unbonding_delegation", data); err != nil {
		return
	}

	// Create subscription
	subKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, delegatorAddress, validatorAddress, "")
	sendCh := make(chan interface{}, 32)
	subscriber := k.registry.Subscribe(subKey, ctx, sendCh)
	defer k.registry.Unsubscribe(subscriber)

	k.handleWebSocketConnection(conn, ctx, sendCh, func() interface{} {
		unbondingDelegation, err := k.stakingKeeper.GetUnbondingDelegation(sdk.UnwrapSDKContext(ctx), delAddr, valAddr)
		data := map[string]interface{}{"found": err == nil}
		if err == nil {
			data["unbonding_delegation"] = unbondingDelegation
		}
		return data
	})
}

// handleWebSocketConnection handles the WebSocket connection lifecycle
func (k *Keeper) handleWebSocketConnection(conn *websocket.Conn, ctx context.Context, sendCh <-chan interface{}, queryFunc func() interface{}) {
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sendCh:
			// Re-query data and send update
			data := queryFunc()
			if err := k.sendWebSocketMessage(conn, "update", data); err != nil {
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
func (k *Keeper) sendWebSocketMessage(conn *websocket.Conn, msgType string, data interface{}) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	message := WebSocketMessage{
		Type: msgType,
		Data: data,
	}
	return conn.WriteJSON(message)
}

// getDelegationResponses gets delegation responses for a delegator
func (k *Keeper) getDelegationResponses(delAddr sdk.AccAddress) []stakingtypes.DelegationResponse {
	ctx := context.Background()
	delegations, err := k.stakingKeeper.GetAllDelegatorDelegations(sdk.UnwrapSDKContext(ctx), delAddr)
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	bondDenom, err := k.stakingKeeper.BondDenom(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return []stakingtypes.DelegationResponse{}
	}

	var delegationResponses []stakingtypes.DelegationResponse
	for _, delegation := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			continue
		}

		validator, err := k.stakingKeeper.GetValidator(sdk.UnwrapSDKContext(ctx), valAddr)
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
func (k *Keeper) getDelegationResponse(delAddr sdk.AccAddress, valAddr sdk.ValAddress) interface{} {
	ctx := context.Background()
	delegation, err := k.stakingKeeper.GetDelegation(sdk.UnwrapSDKContext(ctx), delAddr, valAddr)

	data := map[string]interface{}{"found": err == nil}
	if err == nil {
		bondDenom, bondErr := k.stakingKeeper.BondDenom(sdk.UnwrapSDKContext(ctx))
		if bondErr != nil {
			return data
		}

		validator, valErr := k.stakingKeeper.GetValidator(sdk.UnwrapSDKContext(ctx), valAddr)
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
