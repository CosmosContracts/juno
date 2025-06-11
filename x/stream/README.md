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
4. **Connection Manager** - Enforces connection and subscription limits
5. **gRPC Streaming** - Provides streaming endpoints for clients
6. **WebSocket Server** - Alternative WebSocket interface for browser clients

## Features

- **Automatic Configuration** - Reads connection limits from CometBFT's `config.toml`
- **Connection Management** - Enforces configurable connection and subscription limits
- **Prometheus Metrics** - Comprehensive metrics for monitoring
- **Graceful Degradation** - Returns appropriate HTTP status codes when limits are exceeded

## Usage

### gRPC Streaming with grpcurl

First, make sure your node has gRPC enabled in `app.toml`:

```toml
[grpc]
enable = true
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

### WebSocket Streaming

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

**Balance:**

```json
{
  "denom": "ujuno",
  "amount": "1000000"
}
```

**All Balances:**

```json
{
  "balances": [
    {"denom": "ujuno", "amount": "1000000"},
    {"denom": "uatom", "amount": "500000"}
  ]
}
```

**Delegation:**

```json
{
  "delegation": {
    "delegator_address": "juno1...",
    "validator_address": "junovaloper1...",
    "shares": "1000000.000000000000000000"
  },
  "balance": {
    "denom": "ujuno",
    "amount": "1000000"
  }
}
```

## Production Deployment

### Configuration

The Stream module automatically reads connection limits from your CometBFT `config.toml`:

```toml
[rpc]
# Maximum number of simultaneous connections (including WebSocket).
max_open_connections = 900

# Maximum number of unique queries a given client can /subscribe to
max_subscriptions_per_client = 5
```

These values are applied to the Stream module's WebSocket connections. If not configured, defaults of 900 connections and 5 subscriptions per client are used.

### Monitoring

Enable Prometheus metrics in `app.toml`:

```toml
[telemetry]
enabled = true
prometheus-retention-time = 0
```

Available metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `stream_connections` | Gauge | Number of active WebSocket connections |
| `stream_subscriptions{type="..."}` | Gauge | Active subscriptions by type |
| `stream_messages_sent{type="..."}` | Counter | Total messages sent by type |
| `stream_buffer_overflow{subscription_type="..."}` | Counter | Buffer overflow events |
| `stream_connection_duration` | Histogram | Connection duration in seconds |
| `stream_connection_rejected{reason="..."}` | Counter | Rejected connections by reason |

### Performance Tuning

1. **Buffer Sizes** - The module uses a 10,000 event buffer. Monitor `stream_buffer_overflow` metric.

2. **Connection Limits** - Adjust based on your hardware:
   - Small nodes: 100-500 connections
   - Medium nodes: 500-1000 connections
   - Large nodes: 1000+ connections

3. **Subscription Limits** - Default of 5 per client is reasonable for most use cases.

### Troubleshooting

1. **Connection Rejected (503)**
   - Check `max_open_connections` in config.toml
   - Monitor `stream_connection_rejected{reason="max_connections"}`

2. **Subscription Rejected (429)**
   - Check `max_subscriptions_per_client` in config.toml
   - Monitor `stream_connection_rejected{reason="max_subscriptions"}`

3. **High Memory Usage**
   - Check active connections: `stream_connections`
   - Check subscription count: `stream_subscriptions`
   - Consider reducing connection limits

4. **Messages Not Received**
   - Check `stream_buffer_overflow` metric
   - Indicates clients not reading fast enough
