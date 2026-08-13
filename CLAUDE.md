# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Juno is a sovereign Cosmos SDK / CometBFT chain with CosmWasm smart-contract support (via `wasmd`). The binary is `junod`. Go module path is `github.com/CosmosContracts/juno/v31` — the `/v31` suffix tracks the current chain upgrade name and is bumped each consensus-breaking release (see `RELEASES.md`).

## Toolchain

Go 1.25.10 is pinned in `.mise.toml` (also: `buf`, `yq`). Run `mise install` once to provision them. Dev tools (`golangci-lint`, `gofumpt`, `buf`, protoc plugins) are declared as Go tool dependencies in `go.mod` and invoked via `go tool …` — do not install them separately.

## Common commands

Run from the `juno/` submodule root:

```bash
make build              # → ./bin/junod
make install            # → $GOPATH/bin/junod
make lint               # go tool golangci-lint run --config ./.golangci.yml
make format             # go tool gofumpt -l -w .
make init               # sh scripts/init.sh — wipes ~/.juno and bootstraps a single-validator chain
go test ./...           # unit tests
go test ./x/clock/keeper -run TestEndBlocker -v   # single test pattern
```

Local devnet (Docker, single node, exposes 1317/26656/26657/9090):

```bash
STAKE_TOKEN=ujunox UNSAFE_CORS=true TIMEOUT_COMMIT=1s docker-compose up
```

`docker-compose.yml` runs `scripts/init.sh && scripts/start.sh` inside the container; `scripts/init.sh` is also the source of truth for genesis bootstrap (default denom `ujunox`, chain-id `testchain-1`, keyring `test`).

## Interchaintest (e2e)

End-to-end tests live in `interchaintest/tests/<suite>/` and use Strangelove's interchaintest framework — they spawn real chains in Docker and are slow. The repo is a separate Go module (`interchaintest/go.mod`). Each suite has a `make ictest-<name>` target that runs `go test -race -v -run <Suite> .` after `go clean -testcache`:

`ictest-basic`, `ictest-cw`, `ictest-node`, `ictest-feemarket`, `ictest-fees`, `ictest-upgrade`, `ictest-ibc`, `ictest-ibc-hooks`, `ictest-pfm`, `ictest-tokenfactory`, `ictest-drip`, `ictest-burn`, `ictest-fixes`.

Most suites need a local image — build it first with `make local-image` (uses `docker buildx`).

## Protobuf

Proto generation runs inside a dedicated Docker image — there is no host-toolchain path:

```bash
make proto-image        # build juno-protobuilder:latest (one-time / when proto/Dockerfile changes)
make proto-gen          # proto-gogo + proto-pulsar + proto-openapi
make proto-check        # proto-format + proto-lint
make proto-breaking     # diffs ./proto against origin/main
```

`.proto` sources live in `proto/juno/` and `proto/osmosis/` (tokenfactory inherits Osmosis package paths). OpenAPI specs are wired into the chain's REST layer via `app/endpoints/openapi*.go`.

## Architecture

### App wiring (`app/`)

`app/app.go` defines `App` and assembles everything. The wiring is split into a few intentional pieces — change the matching one rather than `app.go`:

- `app/keepers/keepers.go` — `AppKeepers` struct and `NewAppKeepers(...)`. All module keepers (SDK, IBC, wasm, custom Juno modules) are constructed here, store keys live in `keys.go`, and `wasm_config.go` / `acceptedQueries.go` configure the wasm VM and stargate query allow-list.
- `app/modules.go` — module manager registration, ordering of `BeginBlocker` / `EndBlocker` / `InitGenesis`, and ModuleAccount permissions. Adding a module means touching this file.
- `app/ante/ante.go` — custom `AnteHandler` chaining SDK ante decorators with `wasm`, IBC, and the Juno fee stack (`feemarket`, `feepay`, `feeshare`). `app/ante/decorators/` and `app/ante/msg_filter.go` hold the Juno-specific decorators. `app/post.go` has post-handlers.
- `app/upgrades/` — one subpackage per named upgrade (currently `v30/`), each exporting an `Upgrade` value combining `UpgradeName`, a `CreateUpgradeHandler` constructor, and `StoreUpgrades` (added/deleted/renamed module stores). `app.Upgrades` lists them and `app.go` registers handlers via the upgrade keeper. `v30` deletes legacy `globalfee`, `crisis`, `params`, `nft` stores and adds `feemarket`.
- `app/endpoints/` — REST/OpenAPI/Scalar docs and the websocket endpoint (`endpoints/websocket/`) bolted onto the cosmos-sdk API server.
- `cmd/junod/` — CLI entry (root command, server commands, genesis subcommands like `add-ica-config`).

### Custom `x/` modules

Each follows the standard SDK module layout (`keeper/`, `module/`, `types/`, often `spec/` and a top-level `README.md`). Module summary:

- `x/clock` — registers smart contracts to be `sudo`-executed each `EndBlock` (no external bots).
- `x/cw-hooks` — fires staking/validator lifecycle events into registered contracts.
- `x/feemarket` — AIMD EIP-1559 dynamic base-fee module (originally Skip's). Wired into the ante chain.
- `x/feepay` — gasless UX: contract devs fund a balance that pays user tx fees on their contract.
- `x/feeshare` — directs a share of contract execution fees to the contract's registered withdraw address (Juno Prop 51; based on Evmos x/revenue). Has its own ante decorator at `x/feeshare/ante/`.
- `x/tokenfactory` — Osmosis-derived; permissionless `factory/{creator}/{subdenom}` denoms with admin mint/burn/force-transfer.
- `x/mint` — Juno-specific inflation/issuance fork (note: `x/burn` exists *only* because of this fork — see its README).
- `x/burn` — funnels burned tokens to a fixed module address compatible with the mint fork. Do **not** copy this module to other chains.
- `x/drip` — lets allow-listed senders push tokens into the fee_pool to airdrop stakers.
- `x/stream` — token streaming.
- `x/wrappers/gov` — gov-related wrapper used for migration/compat.

When adding a module: register it in `app/keepers/keepers.go`, in `app/modules.go` (manager + InitGenesis order + permissions), generate proto via `make proto-gen`, and add a store in the next upgrade's `StoreUpgrades.Added` rather than mutating an existing upgrade.

### CosmWasm integration

- `wasmbindings/` exposes Juno-specific custom queries and messages to contracts. `RegisterCustomPlugins` in `wasmbindings/wasm.go` returns `wasmkeeper.Option`s wired in `app/keepers/keepers.go`. Currently surfaces bank metadata and tokenfactory operations to contracts; extend `query_plugin.go` / `message_plugin.go` to add more.
- `app/keepers/acceptedQueries.go` is the stargate query allow-list — anything contracts query via `Stargate{}` must be listed here.
- `wasm_config.go` configures the wasm VM (memory cache size, gas limits, etc.).

### Cross-cutting fee pipeline

Tx fee handling is *not* a single decorator; the order in `app/ante/ante.go` matters: `feemarket` decides the base fee, `feepay` may pay it on behalf of the sender for registered contracts, and `feeshare` (post-handler / dedicated decorator) splits collected fees to contract devs. Touching any one almost always means re-reading the others.

### Versioned upgrades and module path

The Go module is `…/juno/v31`. Internal imports use `github.com/CosmosContracts/juno/v31/...`. For each consensus-breaking major release the entire repo is mass-rewritten: `go.mod` major version, every import, and the new upgrade package/name. Don't pin the current major version outside module/import identity and the corresponding upgrade package.

## Conventions worth knowing

- `make lint` uses the *Go tool* form of golangci-lint declared in `go.mod` — do not assume a system `golangci-lint` is on PATH or matches the pinned version.
- The repo is consumed as a git submodule from a parent meta-repo. CI/builds run from this directory, but be aware top-level `git status` from the parent is not the right place to commit.
- Pre-`v3.0.0` (Lupercalia) genesis is effectively a different chain — `juno-phoenix2-genesis.tar.gz` and `uni-5-1785500-directory.tar.lz4` are historical artifacts kept for sync-from-genesis tooling, not active code.
