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

### WebSocket with wscat

The WebSocket server needs to be started separately. It's not automatically started with the node.

#### Starting the WebSocket Server

The WebSocket server can be started programmatically or through a separate command. By default, it runs on port 8080.

#### Connecting with wscat

```bash
# Install wscat
npm install -g wscat

# Subscribe to balance changes
wscat -c ws://localhost:8080/subscribe/bank/balance/juno1address.../ujuno

# Subscribe to all balances
wscat -c ws://localhost:8080/subscribe/bank/balances/juno1address...

# Subscribe to delegations
wscat -c ws://localhost:8080/subscribe/staking/delegations/juno1address...

# Subscribe to specific delegation
wscat -c ws://localhost:8080/subscribe/staking/delegation/juno1address.../junovaloper1...
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

## Configuration

The module can be configured through parameters:

```bash
# Query current parameters
grpcurl -plaintext \
  localhost:9090 \
  juno.stream.v1.Query/Params

# Update parameters (requires governance proposal in production)
junod tx stream update-params \
  --disabled-module-streams "distribution,gov" \
  --from validator
```

## Implementation Status

### ✅ Completed
- ABCI Listener implementation
- Event dispatcher and subscription registry
- gRPC streaming endpoints
- WebSocket server implementation
- Parameter management
- Genesis import/export
- Basic keeper tests

### ⚠️ Important Notes

1. **WebSocket Server**: The WebSocket server is implemented but needs to be manually started. It's not automatically launched with the node.

2. **Event Detection**: The ABCI listener's `OnWrite` method needs to be called by the store during state changes. This requires the streaming manager to be properly configured in app.go.

3. **Production Considerations**:
   - Add authentication/authorization for WebSocket connections
   - Implement rate limiting
   - Add connection limits
   - Consider using a message queue for high-volume scenarios
   - Add metrics and monitoring

## Testing

Run the keeper tests:
```bash
go test ./x/stream/keeper/...
```

Run integration tests:
```bash
go test ./x/stream/keeper/... -run Integration
```

## Troubleshooting

### No Events Received

1. Check if the stream module is enabled in params
2. Verify the ABCI listener is registered in app.go
3. Ensure the dispatcher is started
4. Check logs for any errors

### WebSocket Connection Failed

1. Ensure the WebSocket server is started
2. Check the port is not blocked
3. Verify the address format in the URL

### gRPC Stream Closed

1. Check node logs for errors
2. Verify gRPC is enabled in app.toml
3. Ensure the address is correct