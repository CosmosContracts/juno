# WebSocket Package Structure

This package contains all WebSocket-related functionality for the stream module.

## Directory Structure

```text
websocket/
├── bank/                # Bank module WebSocket handlers
│   ├── handler.go      # Main bank handler
│   ├── module.go       # Bank module implementation
│   ├── balance.go      # Balance subscription handler
│   ├── all_balances.go # All balances subscription handler
│   └── types.go        # Bank-specific interfaces
├── staking/            # Staking module WebSocket handlers
│   ├── handler.go      # Main staking handler
│   ├── module.go       # Staking module implementation
│   ├── delegations.go  # Delegations subscription handlers
│   ├── unbonding.go    # Unbonding subscription handlers
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

## Adding New Modules

When adding support for new Cosmos SDK modules:

1. Create a new directory for the module (e.g., `gov/`, `distribution/`)
2. Add `types.go` with module-specific interfaces
3. Create `module.go` implementing the `common.Module` interface:
   ```go
   type Module struct {
       *common.ModuleHandler
       keeper KeeperInterface
       name   string
   }
   ```
4. Implement route handlers in `module.go`
5. Update `handler.go` to register the new module
6. Add the necessary methods to `KeeperInterface` in `init.go` if needed
7. Register new routes in `app/websocket.go`

## Example: Adding Gov Module

```go
// websocket/gov/module.go
func NewModule(keeper KeeperInterface, deps *common.HandlerDependencies) *Module {
    m := &Module{
        ModuleHandler: common.NewModuleHandler(deps, keeper),
        keeper:        keeper,
        name:          "gov",
    }
    
    // Register routes
    m.RegisterRoute("proposals", m.handleProposals)
    m.RegisterRoute("proposal", m.handleProposal)
    m.RegisterRoute("votes", m.handleVotes)
    
    return m
}
```

## Design Principles

- **All WebSocket logic stays within this package** - The main keeper only provides data access
- **Each module has its own subdirectory** - Keeps code organized and maintainable
- **Interfaces avoid import cycles** - The websocket package cannot import the keeper package
- **Module-specific interfaces** - Each module defines exactly what it needs, avoiding a giant interface
- **Handler is initialized once** - No lazy initialization, created during keeper construction
- **Direct route registration** - Routes are registered directly on the handler in `app/websocket.go`
