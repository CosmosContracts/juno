# Stream Module

The Stream module provides real-time streaming of blockchain state changes through gRPC streaming and WebSocket endpoints.

## Overview

The Stream module allows clients to subscribe to real-time state changes including:

- Token balance changes
- Delegation changes
- Unbonding delegation changes

## Architecture

The module uses an event-driven architecture:

1. **ABCI Listener** - Intercepts state changes during block processing
2. **Event Dispatcher** - Routes events to appropriate subscribers
3. **Subscription Registry** - Manages active subscriptions
4. **gRPC Streaming** - Provides streaming endpoints for clients
5. **WebSocket Server** - Alternative WebSocket interface for browser clients

## Usage

### gRPC Streaming with grpcurl

First, make sure your node has gRPC enabled in `app.toml`:

```toml
[grpc]
enable = true
address = "0.0.0.0:9090"
```

#### Stream Balance Changes

```bash
# Stream balance changes for a specific address and denom
grpcurl -plaintext \
  -d '{"address":"juno1address...", "denom":"ujuno"}' \
  localhost:9090 \
  juno.stream.v1.Query/StreamBalance

# Stream all balance changes for an address
grpcurl -plaintext \
  -d '{"address":"juno1address..."}' \
  localhost:9090 \
  juno.stream.v1.Query/StreamAllBalances
```

#### Stream Delegation Changes

```bash
# Stream all delegations for a delegator
grpcurl -plaintext \
  -d '{"delegator_address":"juno1address..."}' \
  localhost:9090 \
  juno.stream.v1.Query/StreamDelegations

# Stream a specific delegation
grpcurl -plaintext \
  -d '{"delegator_address":"juno1address...", "validator_address":"junovaloper1..."}' \
  localhost:9090 \
  juno.stream.v1.Query/StreamDelegation
```

#### Stream Unbonding Delegations

```bash
# Stream all unbonding delegations
grpcurl -plaintext \
  -d '{"delegator_address":"juno1address..."}' \
  localhost:9090 \
  juno.stream.v1.Query/StreamUnbondingDelegations

# Stream a specific unbonding delegation
grpcurl -plaintext \
  -d '{"delegator_address":"juno1address...", "validator_address":"junovaloper1..."}' \
  localhost:9090 \
  juno.stream.v1.Query/StreamUnbondingDelegation
```

#### Connecting with websocat

```bash
# Install websocat (macOS)
brew install websocat
# or (Linux, requires rust)
cargo install websocat

# Subscribe to balance changes
websocat ws://localhost:1317/ws/subscribe/bank/balance/juno1address.../ujuno

# Subscribe to all balances
websocat ws://localhost:1317/ws/subscribe/bank/balances/juno1address...

# Subscribe to delegations
websocat ws://localhost:1317/ws/subscribe/staking/delegations/juno1address...

# Subscribe to specific delegation
websocat ws://localhost:1317/ws/subscribe/staking/delegation/juno1address.../junovaloper1...
```

#### WebSocket Message Format

Messages are sent in JSON format:

```json
{
  "type": "balance",
  "data": {
    "denom": "ujuno",
    "amount": "1000000"
  }
}
```
