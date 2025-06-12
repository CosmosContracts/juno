# X/Stream Module Refactoring - Phase 3 Handover Document

## Overview

The x/stream module has grown to 600+ lines in `websocket.go` with all handlers mixed together. To support adding all Cosmos SDK queries (bank, staking, distribution, gov, auth, mint, etc.), we need a scalable architecture with clear separation of concerns.

### Current Structure Issues

- All WebSocket handlers in single 600+ line file
- Difficult to add new query types
- Code duplication across handlers
- No clear pattern for extending to other modules

### Existing Handlers

- `HandleBalanceSubscription` - Single balance query
- `HandleAllBalancesSubscription` - All balances for address
- `HandleDelegationsSubscription` - All delegations
- `HandleDelegationSubscription` - Single delegation
- `HandleUnbondingDelegationsSubscription` - All unbonding
- `HandleUnbondingDelegationSubscription` - Single unbonding

## Proposed Architecture

### Directory Structure

```text
x/stream/
├── keeper/
│   ├── websocket/
│   │   ├── handler.go         # Base WebSocket handler interface
│   │   ├── router.go          # WebSocket request router
│   │   ├── middleware/        # WebSocket-specific middleware
│   │   │   ├── middleware.go  # Circuit breaker
│   │   ├── bank/
│   │   │   ├── handler.go     # Bank-specific handlers
│   │   │   ├── balance.go     # Balance query handler
│   │   │   └── types.go       # Bank-specific types (move to x/stream/types/bank.go)
│   │   ├── staking/
│   │   │   ├── handler.go     # Staking-specific handlers
│   │   │   ├── delegations.go # Delegation handlers
│   │   │   ├── unbonding.go   # Unbonding handlers
│   │   │   └── types.go       # Staking-specific types (move to x/stream/types/staking.go)
│   ├── grpc/
│   │   ├── server.go          # gRPC stream server
│   │   ├── handler.go         # Base gRPC handler interface
│   │   ├── bank/
│   │   ├── staking/
│   ├── connection_manager.go  # (existing)
│   ├── dispatcher.go          # (existing)
│   ├── metrics.go             # (moved to x/stream/types/metrics.go)
│   ├── config.go              # (existing)
│   └── keeper.go              # (existing)
```

## Implementation Plan

### Phase 3.1: Core Infrastructure

#### 1. Define Base Interfaces

```go
// keeper/common/handler.go
type QueryHandler interface {
    // Validate validates the request parameters
    Validate(ctx context.Context, req interface{}) error

    // Execute performs the query
    Execute(ctx context.Context, req interface{}) (interface{}, error)

    // GetQueryRoute returns the query route (e.g., "bank/balance")
    GetQueryRoute() string

    // GetSubscriptionKey generates subscription key for this query
    GetSubscriptionKey(req interface{}) types.SubscriptionKey
}

// keeper/websocket/handler.go
type WebSocketHandler interface {
    QueryHandler

    // HandleSubscription manages the WebSocket connection lifecycle
    HandleSubscription(w http.ResponseWriter, r *http.Request)

    // GetInitialData fetches initial data for the subscription
    GetInitialData(ctx context.Context, req interface{}) (interface{}, error)
}
```

#### 2. Create Router Infrastructure

```go
// keeper/websocket/router.go
type Router struct {
    handlers map[string]WebSocketHandler
    keeper   *Keeper
    logger   log.Logger
}

func (r *Router) RegisterHandler(route string, handler WebSocketHandler)
func (r *Router) Route(pattern string) http.HandlerFunc
```

#### 3. Implement Base Handler

```go
// keeper/websocket/handler.go
type BaseWebSocketHandler struct {
    keeper         *Keeper
    queryRoute     string
    validateFunc   ValidateFunc
    executeFunc    ExecuteFunc
    subscriptionFunc SubscriptionFunc
}

// Implements common WebSocket lifecycle management
func (h *BaseWebSocketHandler) HandleSubscription(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request parameters
    // 2. Validate request
    // 3. Check connection limits
    // 4. Upgrade to WebSocket
    // 5. Check circuit breaker
    // 6. Send initial data
    // 7. Create subscription
    // 8. Handle connection lifecycle
}
```

use the implementation in `keeper/websocket.go` and migrate it to the new structure. make sure to remove as much code duplication as possible.

### Phase 3.2: Module Migration

#### 1. Bank Module

Create bank-specific handlers:

- `BalanceHandler` - Single balance query
- `AllBalancesHandler` - All balances query
- `SupplyHandler` - Token supply query (new)
- `DenomMetadataHandler` - Denom metadata query (new)

```go
// keeper/websocket/bank/balance.go
type BalanceHandler struct {
    BaseWebSocketHandler
    bankKeeper bankkeeper.Keeper
}

func NewBalanceHandler(k *Keeper) *BalanceHandler {
    return &BalanceHandler{
        BaseWebSocketHandler: BaseWebSocketHandler{
            keeper:     k,
            queryRoute: "bank/balance",
            validateFunc: validateBalanceRequest,
            executeFunc:  executeBalanceQuery,
        },
        bankKeeper: k.bankKeeper,
    }
}
```

#### 2. Staking Module

Migrate existing staking handlers:

- `DelegationsHandler`
- `DelegationHandler`
- `UnbondingDelegationsHandler`
- `UnbondingDelegationHandler`
- `ValidatorsHandler` (new)
- `ValidatorHandler` (new)
- `RedelegationsHandler` (new)

### Phase 3.3: Advanced Features

#### 1. Query Registration System

```go
// Auto-register queries at startup
func (k *Keeper) RegisterQueries() {
    router := websocket.NewRouter(k)

    // Bank queries
    router.RegisterHandler("bank/balance", bank.NewBalanceHandler(k))
    router.RegisterHandler("bank/all-balances", bank.NewAllBalancesHandler(k))

    // Staking queries
    router.RegisterHandler("staking/delegations", staking.NewDelegationsHandler(k))
    // ... etc
}
```

#### 2. Dynamic Route Generation

```go
// Generate routes from registered handlers
func (k *Keeper) GetWebSocketRoutes() []Route {
    var routes []Route
    for path, handler := range k.router.handlers {
        routes = append(routes, Route{
            Path:    "/ws/" + path + "/{params}",
            Handler: handler.HandleSubscription,
        })
    }
    return routes
}
```

## Migration Strategy

### Step 1: Create New Structure (Non-Breaking)

1. Create new directory structure
2. Implement base interfaces and handlers by migrating the existing handlers in `keeper/websocket.go` to the new structure. afterwards remove the old implementations.
3. Create router infrastructure

### Step 2: Gradual Migration

1. Implement new bank module handlers
2. Route new endpoints through new system
3. Remove old endpoints
4. Migrate one module at a time

### Step 3: Cutover

1. Remove old handler code
2. Clean up legacy code

## Code Examples

### Example: New Balance Handler

```go
// keeper/websocket/bank/balance.go
package bank

import (
    "context"
    "fmt"

    sdk "github.com/cosmos/cosmos-sdk/types"
    banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

    "github.com/CosmosContracts/juno/v30/x/stream/keeper/common"
    "github.com/CosmosContracts/juno/v30/x/stream/types"
)

type BalanceRequest struct {
    Address string `json:"address"`
    Denom   string `json:"denom"`
}

type BalanceHandler struct {
    *common.BaseHandler
    bankKeeper banktypes.QueryServer
}

func NewBalanceHandler(bankKeeper banktypes.QueryServer) *BalanceHandler {
    return &BalanceHandler{
        BaseHandler: &common.BaseHandler{
            QueryRoute: "bank/balance",
        },
        bankKeeper: bankKeeper,
    }
}

func (h *BalanceHandler) Validate(ctx context.Context, req interface{}) error {
    balReq, ok := req.(*BalanceRequest)
    if !ok {
        return fmt.Errorf("invalid request type")
    }

    if _, err := sdk.AccAddressFromBech32(balReq.Address); err != nil {
        return fmt.Errorf("invalid address: %w", err)
    }

    if err := sdk.ValidateDenom(balReq.Denom); err != nil {
        return fmt.Errorf("invalid denom: %w", err)
    }

    return nil
}

func (h *BalanceHandler) Execute(ctx context.Context, req interface{}) (interface{}, error) {
    balReq := req.(*BalanceRequest)

    queryReq := &banktypes.QueryBalanceRequest{
        Address: balReq.Address,
        Denom:   balReq.Denom,
    }

    return h.bankKeeper.Balance(ctx, queryReq)
}

func (h *BalanceHandler) GetSubscriptionKey(req interface{}) types.SubscriptionKey {
    balReq := req.(*BalanceRequest)
    return types.GenerateSubscriptionKey(
        types.SubscriptionTypeBalance,
        balReq.Address,
        "",
        balReq.Denom,
    )
}
```

### Example: Router Usage

```go
// In app.go or module initialization
router := websocket.NewRouter(streamKeeper)

// Register all bank handlers
router.RegisterHandler("bank/balance", bank.NewBalanceHandler(app.BankKeeper))
router.RegisterHandler("bank/all-balances", bank.NewAllBalancesHandler(app.BankKeeper))
router.RegisterHandler("bank/supply", bank.NewSupplyHandler(app.BankKeeper))

// Register all staking handlers
router.RegisterHandler("staking/delegations", staking.NewDelegationsHandler(app.StakingKeeper))
router.RegisterHandler("staking/validators", staking.NewValidatorsHandler(app.StakingKeeper))

// Mount routes
mux.HandleFunc("/ws/query/{module}/{query}", router.Route())
```

## Benefits

1. **Scalability**: Easy to add new modules and queries
2. **Maintainability**: Clear separation of concerns
3. **Testability**: Each handler can be tested independently
4. **Consistency**: Uniform patterns across all modules
5. **Discoverability**: Self-documenting query routes
6. **Performance**: Reusable middleware and optimized routing
7. **Type Safety**: Strong typing for requests/responses (use imports from cosmossdk.io/api/cosmos/moduleName/v1 or v1beta1)

## Testing Strategy

1. **Unit Tests**: Test each handler in isolation
2. **Integration Tests**: Test full query flow
3. **Load Tests**: Ensure performance under load

## Next Steps

1. Create Todo List for this plan
2. Implement core infrastructure
3. Begin incremental migration of previous code
4. Hook up new code in app/websocket.go
5. Check IDE diagnostics, fix all issues and warnings and remove all unused imports
6. Update documentation
