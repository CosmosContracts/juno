# WebSocket Package Documentation

This package contains all WebSocket-related functionality for the stream module.

## Directory Structure

```text
websocket/
├── bank/                # Bank module WebSocket handlers
│   ├── handler.go      # Main bank handler
│   ├── module.go       # Bank module implementation
│   └── types.go        # Bank-specific interfaces
├── staking/            # Staking module WebSocket handlers
│   ├── handler.go      # Main staking handler
│   ├── module.go       # Staking module implementation
│   └── types.go        # Staking-specific interfaces
├── common/             # Common utilities and types
│   ├── module_handler.go # Generic module handler
│   ├── query_wrapper.go  # Query wrapping utilities
│   ├── registry.go       # Module registry system
│   └── ...
├── middleware/         # Circuit breaker and other middleware
├── handler.go          # Main WebSocket handler that combines all modules
└── init.go             # Initialization helpers and keeper interface
```

## Architecture

### Key Components

- **init.go**: Defines the `KeeperInterface` that the main keeper must implement to avoid import cycles
- **handler.go**: Combines all module handlers (bank, staking, etc.) into a single handler with module registry
- **Module-specific types.go**: Each module defines its own interfaces for what it needs from the keeper
- **module.go**: Each module implements the common Module interface for consistent behavior
- **common/module_handler.go**: Provides generic handler functionality to reduce boilerplate
- **common/query_wrapper.go**: Simplifies query function wrapping with error handling
- **common/registry.go**: Module registry for dynamic module management

### Key Improvements

1. **Generic Module Handler**: Reduces code duplication across modules
2. **Query Wrapper Utilities**: Simplifies repetitive query wrapping patterns
3. **Circuit Breaker Cleanup**: Automatic cleanup of stale connections
4. **Module Registry**: Dynamic module registration and route management
5. **Automatic Route Registration**: Routes are automatically discovered and registered from modules

## Automatic Route Registration

The WebSocket system now supports automatic route registration, eliminating the need to manually register each route in `app/websocket.go`.

### How It Works

1. **Modules define their routes** with full HTTP patterns:

    ```go
    m.routeDefinitions = []common.RouteDefinition{
        {Pattern: "/ws/subscribe/bank/balance/{address}/{denom}", Handler: m.handleBalance},
        {Pattern: "/ws/subscribe/bank/balances/{address}", Handler: m.handleAllBalances},
    }
    ```

2. **The handler discovers all routes** from registered modules:

    ```go
    func (h *Handler) RegisterRoutes(registerFunc func(pattern string, handler http.HandlerFunc)) {
        routes := h.registry.GetAllRouteDefinitions()
        for _, route := range routes {
            registerFunc(route.Pattern, http.HandlerFunc(route.Handler))
        }
    }
    ```

3. **App registers routes with one call**:

    ```go
    wsHandler.RegisterRoutes(func(pattern string, handler http.HandlerFunc) {
        apiSvr.Router.Handle(pattern, handler)
    })
    ```

## Available WebSocket Routes

### Bank Module Routes

- `/ws/subscribe/bank/balance/{address}/{denom}` - Subscribe to balance updates for a specific address and denomination
- `/ws/subscribe/bank/balances/{address}` - Subscribe to all balance updates for an address
- `/ws/subscribe/bank/spendable-balances/{address}` - Subscribe to spendable balances
- `/ws/subscribe/bank/spendable-balance/{address}/{denom}` - Subscribe to spendable balance for specific denom
- `/ws/subscribe/bank/total-supply` - Subscribe to total supply updates
- `/ws/subscribe/bank/supply/{denom}` - Subscribe to supply of specific denom
- `/ws/subscribe/bank/params` - Subscribe to bank module parameters
- `/ws/subscribe/bank/denoms-metadata` - Subscribe to all denominations metadata
- `/ws/subscribe/bank/denom-metadata/{denom}` - Subscribe to specific denom metadata
- `/ws/subscribe/bank/denom-owners/{denom}` - Subscribe to denom owners
- `/ws/subscribe/bank/send-enabled` - Subscribe to send enabled status

### Staking Module Routes

- `/ws/subscribe/staking/delegations/{delegator}` - Subscribe to all delegations for a delegator
- `/ws/subscribe/staking/delegation/{delegator}/{validator}` - Subscribe to a specific delegation
- `/ws/subscribe/staking/unbonding-delegations/{delegator}` - Subscribe to all unbonding delegations
- `/ws/subscribe/staking/unbonding-delegation/{delegator}/{validator}` - Subscribe to a specific unbonding delegation

## Adding New Modules

When adding support for new Cosmos SDK modules:

1. **Create a new directory** for the module (e.g., `gov/`, `distribution/`)

2. **Add `types.go`** with module-specific interfaces:

    ```go
    type KeeperInterface interface {
        GetGovQueryServer() govtypes.QueryServer
        // Add other required methods
    }
    ```

3. **Create `module.go`** implementing the `common.Module` interface:

    ```go
    type Module struct {
        *common.ModuleHandler
        keeper KeeperInterface
        name   string
        routeDefinitions []common.RouteDefinition
    }

    func NewModule(keeper KeeperInterface, deps *common.HandlerDependencies) *Module {
        m := &Module{
            ModuleHandler: common.NewModuleHandler(deps, keeper),
            keeper:        keeper,
            name:          "gov",
        }

        // Define routes with full HTTP patterns
        m.routeDefinitions = []common.RouteDefinition{
            {Pattern: "/ws/subscribe/gov/proposals", Handler: m.handleProposals},
            {Pattern: "/ws/subscribe/gov/proposal/{id}", Handler: m.handleProposal},
            {Pattern: "/ws/subscribe/gov/votes/{proposal_id}", Handler: m.handleVotes},
        }

        // Register internal routes (optional, for backward compatibility)
        m.RegisterRoute("proposals", m.handleProposals)
        m.RegisterRoute("proposal", m.handleProposal)
        m.RegisterRoute("votes", m.handleVotes)

        return m
    }
    ```

4. **Implement required methods**:

    ```go
    func (m *Module) Name() string { return m.name }
    func (m *Module) Routes() map[string]common.RouteHandler { return m.GetRoutes() }
    func (m *Module) RouteDefinitions() []common.RouteDefinition { return m.routeDefinitions }
    ```

5. **Register the module** in `handler.go`:

    ```go
    govModule := gov.NewModule(govKeeper, deps)
    if err := moduleRegistry.RegisterModule(govModule); err != nil {
        logger.Error("failed to register gov module", "error", err)
    }
    ```

6. **Update `KeeperInterface`** in `init.go` if needed

Routes will be automatically registered when the app calls `RegisterRoutes()`!

## Route Pattern Guidelines

- Use kebab-case for multi-word endpoints: `/ws/subscribe/bank/spendable-balances`
- Include module name in the path: `/ws/subscribe/{module}/{endpoint}`
- Use meaningful parameter names: `{address}`, `{denom}`, `{validator}`
- Keep patterns RESTful and intuitive

## Design Principles

- **All WebSocket logic stays within this package** - The main keeper only provides data access
- **Each module has its own subdirectory** - Keeps code organized and maintainable
- **Interfaces avoid import cycles** - The websocket package cannot import the keeper package
- **Module-specific interfaces** - Each module defines exactly what it needs, avoiding a giant interface
- **Handler is initialized once** - No lazy initialization, created during keeper construction
- **Automatic route discovery** - Routes are defined by modules and automatically registered

## Benefits of Automatic Registration

1. **No Manual Route Registration**: Routes are discovered from modules automatically
2. **Single Source of Truth**: Route patterns are defined in the modules that handle them
3. **Type Safety**: Route patterns and handlers are defined together at compile time
4. **Easy Discovery**: All routes for a module are defined in one place
5. **Consistent Patterns**: Enforces consistent URL patterns across modules
