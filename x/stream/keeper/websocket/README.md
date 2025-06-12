# WebSocket Package Structure

This package contains all WebSocket-related functionality for the stream module.

## Directory Structure

```text
websocket/
├── bank/                # Bank module WebSocket handlers
│   ├── handler.go      # Main bank handler
│   ├── balance.go      # Balance subscription handler
│   ├── all_balances.go # All balances subscription handler
│   └── types.go        # Bank-specific interfaces
├── staking/            # Staking module WebSocket handlers
│   ├── handler.go      # Main staking handler
│   ├── delegations.go  # Delegations subscription handlers
│   ├── unbonding.go    # Unbonding subscription handlers
│   └── types.go        # Staking-specific interfaces
├── common/             # Common utilities and types
├── middleware/         # Circuit breaker and other middleware
├── handler.go          # Main WebSocket handler that combines all modules
└── init.go             # Initialization helpers and keeper interface
```

## Architecture

### Key Components

- **init.go**: Defines the `KeeperInterface` that the main keeper must implement to avoid import cycles
- **handler.go**: Combines all module handlers (bank, staking, etc.) into a single handler
- **Module-specific types.go**: Each module defines its own interfaces for what it needs from the keeper

## Adding New Modules

When adding support for new Cosmos SDK modules:

1. Create a new directory for the module (e.g., `gov/`, `distribution/`)
2. Add module-specific handlers in that directory
3. Define module-specific interfaces in `types.go`
4. Update `handler.go` to include the new module handler
5. Add the necessary methods to `KeeperInterface` in `init.go` if needed
6. Register new routes in `app/websocket.go`

## Example: Adding Gov Module

```text
1. Create websocket/gov/ directory
2. Add gov/types.go:
   - Define GovKeeperInterface with methods like GetProposal(), GetVotes(), etc.
3. Add gov/handler.go:
   - Implement the main Gov handler
4. Add specific handlers like gov/proposals.go, gov/votes.go
5. Update handler.go to include govHandler
6. Update init.go KeeperInterface if gov needs new common methods
7. Add routes in app/websocket.go
```

## Design Principles

- **All WebSocket logic stays within this package** - The main keeper only provides data access
- **Each module has its own subdirectory** - Keeps code organized and maintainable
- **Interfaces avoid import cycles** - The websocket package cannot import the keeper package
- **Module-specific interfaces** - Each module defines exactly what it needs, avoiding a giant interface
- **Handler is initialized once** - No lazy initialization, created during keeper construction
- **Direct route registration** - Routes are registered directly on the handler in `app/websocket.go`
