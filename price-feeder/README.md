# Oracle Price Feeder

The `price-feeder` tool is an extension of Juno's `x/oracle` module, both of
which are based on Terra's [x/oracle](https://github.com/terra-money/classic-core/tree/main/x/oracle) and Umee [x/oracle](https://github.com/umee-network/tree/umee/x/oracle) module.
core differences are as follows:

- All exchange rates must be quoted in USD or USD stablecoins.
- No need or use of reference exchange rates (e.g. Luna).
- No need or use of Tobin tax.
- The `price-feeder` combines both `feeder` and `price-server` into a single
  Golang-based application for better UX, testability, and integration.

## Background

The `price-feeder` tool is responsible for performing the following:

1. Fetching and aggregating exchange rate price data from various providers, e.g.
   Binance and Osmosis, based on operator configuration. These exchange rates
   are exposed via an API and are used to feed into the main oracle process.
2. Taking aggregated exchange rate price data and submitting those exchange rates
   on-chain to Juno's `x/oracle` module following Juno's [Oracle](https://github.com/CosmosContracts/juno/tree/main/x/oracle#readme)
   specification.

<!-- markdown-link-check-disable -->
## Providers

The list of current supported providers:

- [Binance](https://www.binance.com/en)
- [Bitget](https://www.bitget.com/)
- [Coinbase](https://www.coinbase.com/)
- [Crypto](https://crypto.com/)
- [FTX](https://ftx.com/)
- [Gate](https://www.gate.io/)
- [Huobi](https://www.huobi.com/en-us/)
- [Kraken](https://www.kraken.com/en-us/)
- [Mexc](https://www.mexc.com/)
- [Okx](https://www.okx.com/)
- [Osmosis](https://app.osmosis.zone/)
<!-- markdown-link-check-enable -->

## Usage

The `price-feeder` tool runs off of a single configuration file. This configuration
file defines what exchange rates to fetch and what providers to get them from.
In addition, it defines the oracle's keyring and feeder account information.
The keyring's password is defined via environment variables or user input.
More information on the keyring can be found [here](#keyring)
Please see the [example configuration](price-feeder.example.toml) for more details.

```shell
$ price-feeder /path/to/price_feeder_config.toml
```

## Configuration

### `telemetry`

A set of options for the application's telemetry, which is disabled by default. An in-memory sink is the default, but Prometheus is also supported. We use the [cosmos sdk telemetry package](https://github.com/cosmos/cosmos-sdk/blob/3689d6f41ad8afa6e0f9b4ecb03b4d7f2d3a9e94/docs/docs/core/09-telemetry.md).

### `deviation`

Deviation allows validators to set a custom amount of standard deviations around the median which is helpful if any providers become faulty. It should be noted that the default for this option is 1 standard deviation.

### `provider_endpoints`

The provider_endpoints option enables validators to setup their own API endpoints for a given provider.

### `server`

The `server` section contains configuration pertaining to the API served by the
`price-feeder` process such the listening address and various HTTP timeouts.

### `currency_pairs`

The `currency_pairs` sections contains one or more exchange rates along with the
providers from which to get market data from. It is important to note that the
providers supplied in each `currency_pairs` must support the given exchange rate.

For example, to get multiple price points on ATOM, you could define `currency_pairs`
as follows:

```toml
[[currency_pairs]]
base = "ATOM"
providers = [
  "binance",
]
quote = "USDT"

[[currency_pairs]]
base = "ATOM"
providers = [
  "kraken",
  "osmosis",
]
quote = "USD"
```

Providing multiple providers is beneficial in case any provider fails to return
market data. Prices per exchange rate are submitted on-chain via pre-vote and
vote messages using a time-weighted average price (TVWAP).

### `provider_min_override`

At startup the amount of possible providers for a currency is checked by querying the
CoinGecko API to enforce an acceptable minimum providers for a given currency pair. If
this request fails and `provider_min_override` is set to true, the minimum is not enforced
and the `price-feeder` is allowed to run irrespective of how many providers are provided
for a given currency pair. `provider_min_override` will not take effect if CoinGecko
requests are successful.

### `account`

The `account` section contains the oracle's feeder and validator account information.
These are used to sign and populate data in pre-vote and vote oracle messages.

### `keyring`

The `keyring` section contains Keyring related material used to fetch the key pair
associated with the oracle account that signs pre-vote and vote oracle messages.

### `rpc`

The `rpc` section contains the Tendermint and Cosmos application gRPC endpoints.
These endpoints are used to query for on-chain data that pertain to oracle
functionality and for broadcasting signed pre-vote and vote oracle messages.

## Keyring

Our keyring must be set up to sign transactions before running the price feeder.
Additional info on the different keyring modes is available [here](https://docs.cosmos.network/v0.46/run-node/keyring.html).
**Please note that the `test` and `memory` modes are only for testing purposes.**
**Do not use these modes for running the price feeder against mainnet.**

### Setup

The keyring `dir` and `backend` are defined in the config file.
You may use the `PRICE_FEEDER_PASS` environment variable to set up the keyring password.

Ex :
`export PRICE_FEEDER_PASS=keyringPassword`

If this environment variable is not set, the price feeder will prompt the user for input.
