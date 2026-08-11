# Juno v31 release plan

Status: draft for implementation  
Prepared: 2026-08-11  
Baseline: `v30.0.0` (`c0b3a8d2`)

## Executive decision

V31 should be a hardening and compatible-patch release, not the Cosmos SDK
0.54 / IBC-Go 11 / store-v2 migration.

The recommended dependency target is:

| Component | v30 | v31 target | Decision |
|---|---:|---:|---|
| Cosmos SDK | `v0.53.7` | `v0.53.8` | Upgrade |
| wasmd | `v0.61.11` | `v0.61.14` | Upgrade |
| wasmvm | `v3.0.4` | `v3.0.7` | Upgrade; consensus-sensitive |
| CometBFT | `v0.38.23` | `v0.38.25` | Upgrade |
| IBC-Go | `v10.6.0` | `v10.7.0` | Upgrade |
| packet-forward middleware | `v10.6.0` | `v10.6.0` | Keep; latest compatible release |
| ibc-hooks | `v10.0.0` | `v10.0.0` | Keep; latest v10 release |
| Go | inconsistent `1.25.2` / `1.25.10` | one supported patch everywhere | Resolve before RC |

This stack preserves the module and store architecture already exercised by
v30 while taking important malformed-transaction, CosmWasm, networking, IBC,
and race-condition fixes. Cosmos SDK 0.54.4, wasmd 0.70.3, IBC-Go 11.2.0,
CometBFT 0.39.4, and ibc-hooks 11.0.0 form a plausible future stack, but
packet-forward middleware still has no `/v11` release. Mixing IBC-Go v11 and
PFM v10 is not supported because their module paths and types differ. Juno
should not drop or privately fork PFM merely to force this migration into v31.

Revisit the 0.54/store-v2 migration for v32 when all of these gates are met:

- upstream PFM publishes a compatible `/v11` release;
- indexers, explorers, snapshots, state sync, and export/import are proven on
  store-v2;
- the SDK, IBC, CometBFT, and wasmd migration guides have been reviewed against
  every custom Juno module;
- a mainnet-state rehearsal produces the expected app hash and survives
  restart, state sync, and export/import.

Do not target SDK 0.55 in v31. It is a new, unmatched line with additional
mempool, sign-mode, store-path, and toolchain changes.

## Why these patches matter

- [Cosmos SDK v0.53.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.8)
  fixes several malformed-transaction panic paths, signature/signer bounds and
  key validation, distribution accounting edge cases, a full-share
  redelegation edge case, and a cachemulti concurrency race.
- [wasmd v0.61.14](https://github.com/CosmWasm/wasmd/releases/tag/v0.61.14)
  pairs the current SDK line with wasmvm 3.0.7.
  [wasmvm v3.0.7](https://github.com/CosmWasm/wasmvm/releases/tag/v3.0.7)
  includes the CosmWasm overflow fix introduced in this patch series. VM
  changes can change deterministic execution outcomes and therefore require a
  coordinated upgrade and contract replay even though the major version is
  unchanged.
- [IBC-Go v10.7.0](https://github.com/cosmos/ibc-go/releases/tag/v10.7.0)
  adds IBC v2 transfer validation and fixes binary store-key parsing in packet
  commitment and acknowledgement queries.
- [CometBFT v0.38.25](https://github.com/cometbft/cometbft/releases/tag/v0.38.25)
  fixes a double-sign check off-by-one, blocksync response handling and
  validation, consensus/mempool races, and heap amplification. CometBFT's
  [versioning policy](https://github.com/cometbft/cometbft#versioning) permits
  state-compatible patch upgrades within v0.38, but mixed-version testing is
  still required.
- ibc-hooks has a [v11 release](https://github.com/cosmos/ibc-apps/releases/tag/modules%2Fibc-hooks%2Fv11.0.0),
  while the latest PFM release remains
  [v10.6.0](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv10.6.0).
  This is the blocking compatibility gap for the ambitious stack.

Before merging dependency changes, capture the selected modules' `go.mod`
compatibility in the PR and run `go mod graph` to ensure Minimal Version
Selection did not silently choose a different SDK, CometBFT, IBC, or wasmvm
version.

## Release scope

### P0 — immediate correctness and release-pipeline fixes

These fixes should land first and may be backported to a v30 patch when they
are non-consensus-breaking.

1. **Reject transaction gas limits above `math.MaxInt64`.**
   `app/ante/decorators/handle_fees.go` converts the transaction's `uint64` gas
   limit to `int64` in normal and FeePay fee arithmetic. Values above
   `MaxInt64` wrap negative and can reach decimal, coin, or priority math.
   Reject them once, early in ante, and test `MaxInt64` and `MaxInt64+1` for
   ordinary, feegrant, and FeePay transactions.

2. **Use the configured feemarket fee denom in FeePay.**
   Ordinary fees use `feemarket.Params.FeeDenom`, while `handleZeroFees` asks
   for the bond-denom price. Make the funding, pricing, refund, and accounting
   rules use one declared denom and add a test where fee denom differs from
   bond denom.

3. **Repair the `juno-std` release dispatch.**
   `.github/workflows/release-dispatch.yml` reads a tag only from manual
   `payload.inputs`; a normal `release` event supplies
   `context.payload.release.tag_name`. Its proto pins are also pre-v30 (SDK
   0.53.4, wasmd 0.54.2, CometBFT 0.38.19, IBC-Go 8.7.0). Derive the release
   tag correctly, update or derive dependency revisions from the released
   `go.mod`, add a recoverable `workflow_dispatch` path, and test the payload.

4. **Unify the Go toolchain.**
   `.mise.toml` is `1.25.10`, while both Go modules, `Dockerfile`, and CI use
   `1.25.2`. Select a patch supported by the dependency stack and pin it in all
   locations. Keep `.mise.toml`, `go.mod`, `interchaintest/go.mod`, Docker build
   args/`GOTOOLCHAIN`, and workflow variables identical.

5. **Fix the PFM third-hop escrow assertion.**
   `interchaintest/tests/pfm/pfm_test.go` builds Chain C's escrow address with
   `cdChan.PortID` but `abChan.ChannelID`. Use `cdChan.ChannelID` and assert
   balances independently at every hop so the test cannot mask routing or
   accounting regressions.

### P1 — dependency and upgrade implementation

1. Bump to the dependency matrix above one layer at a time, with focused
   commits and tests after SDK, wasm, IBC middleware, and CometBFT changes.
2. Rewrite the module path and internal imports from
   `github.com/CosmosContracts/juno/v30` to `/v31`, including the
   interchaintest replacement.
3. Add `app/upgrades/v31/` with:
   - `UpgradeName = "v31"`;
   - an upgrade handler that runs module migrations and only intentional state
     transformations;
   - an explicit `StoreUpgrades` table, even if empty;
   - unit tests for the version map, store list, params, and idempotence/error
     handling.
4. Register the v31 upgrade without modifying the already-released v30
   handler. Historical upgrade code and constants are immutable.
5. Audit upstream changelogs and migrations across `app/`, custom `x/`
   modules, ante/post handlers, IBC middleware order, custom Wasm bindings,
   state streaming, and export/import.
6. Treat wasmvm 3.0.4 to 3.0.7 as consensus-sensitive: replay real contracts,
   compare outcomes around overflow/error behavior, and require a coordinated
   upgrade rather than a rolling binary replacement.

### P1 — close v30 validation gaps

1. **Finish DAO DAO integration tests.** The three tests in
   `interchaintest/tests/dao-dao/dao_dao_test.go` skip unconditionally and the
   helpers are stubs. Implement:
   - cw4 DAO instantiate, propose, vote, execute;
   - cw20-staked voting;
   - a real contract querying `VotingPowerAt`, `TotalVotingPowerAt`, and range
     queries;
   - artifact source tags and checksums;
   - replay of representative currently deployed DAO contracts against a
     recent mainnet snapshot.

2. **Make state-sync testing real.** `TestStateSync` skips with the default
   zero-full-node chain spec. Configure the topology it requires or stop
   claiming it as a release gate.

3. **Audit genesis export/import completeness.** Voting-snapshot and FeePay
   have documented incomplete export behavior. Define whether full historical
   state must round-trip; implement it or explicitly document and test the
   supported invariant. Exercise export/import, snapshot restore, and state
   sync after the upgrade.

4. **Benchmark voting-snapshot worst cases.** Slash/status handling can
   materialize every delegation for a large validator and perform repeated
   validator lookups at EndBlock. Benchmark mainnet-sized (including >10,000
   delegator) cases and, if needed, add an iterator/index/cache or safely
   bounded work strategy without allowing stale same-block voting power.

5. **Reassess async-ICQ.** Confirm whether a compatible maintained release and
   real counterparty demand now exist. Reintroducing the deleted module/store
   is a consensus change and needs explicit store migration and channel
   coordination; otherwise record it as out of scope.

### P2 — maintainability and validation

- Validate the creator address in `wasmbindings/queries.go` and resolve the
  encoding TODO in `wasmbindings/message_plugin.go`; malformed contract input
  should return a precise error rather than an empty or misleading response.
- Make tokenfactory `MsgCreateDenom.ValidateBasic` validate sender/subdenom,
  and make force-transfer zero-amount behavior consistent with mint/burn.
- Reassess the deprecated `module.AppModule` to `appmodule.AppModule`
  migration. Do it only if every selected ecosystem module supports the same
  model; otherwise retain the compatibility layer with a documented blocker.
- Reassess collections schema registration in `app/keepers/keepers.go` and the
  stream keeper after the SDK bump.
- Remove vestigial capability/scoped-keeper wiring only after confirming that
  the selected IBC stack and upgrade history no longer require it.
- Archive or label `planning/` as v30 history. Its deferred-work list is stale:
  many voting-snapshot, stream, fee-bypass, contract-cap, and DenomMetadata
  items were completed before the v30 tag.

## CI and test gates

Build CI currently runs only for a narrow path filter. Expand it to pull
requests and changes to `go.mod`, both module sums, proto files, generated API,
Dockerfile, Makefile, scripts, and workflows. Verify both Go modules, validate
workflow YAML, and add a generated-proto clean-tree check.

Retries must not turn deterministic E2E failures into green runs. Retry only
classified infrastructure failures and preserve the first failure's logs.
Skipped or stubbed tests do not satisfy a release gate.

### Gate A — every PR

```sh
go version
go mod verify
go build ./...
go test ./...
go test -race ./...
go tool golangci-lint run --config ./.golangci.yml
go mod tidy
git diff --exit-code

(cd interchaintest && go build ./...)
(cd interchaintest && go test ./...)
(cd interchaintest && go mod tidy)
git diff --exit-code
```

Also run protobuf generation/checks when proto inputs or dependency revisions
change, and require a clean generated diff.

### Gate B — local image and E2E

Run the complete Makefile interchaintest matrix, not only a smoke subset. At a
minimum:

```sh
make local-image
make ictest-basic
make ictest-cw
make ictest-upgrade
make ictest-ibc
make ictest-ibc-hooks
make ictest-pfm
make ictest-feemarket
make ictest-fees
make ictest-dao-dao
make ictest-node
```

The upgrade test must start from the exact released `v30.0.0` image and move
to the candidate v31 image. It must assert:

- module version-map and store transitions;
- feemarket state plus normal, feegrant, FeePay, and feeshare paths;
- voting-snapshot history, params, and contract queries;
- clock/cw-hooks registrations and caps;
- IBC clients, channels, ICA, hooks, and multi-hop PFM packet flows;
- stored Wasm code and contract state execute successfully;
- restart, snapshot restore, state sync, and export/import after upgrade.

### Gate C — mainnet-state rehearsal

Before the first RC, run the v30-to-v31 handler on a recent `juno-1` snapshot.
Record snapshot height/hash, old and new binary commits, dependency graph,
upgrade height, handler duration, pre/post app hashes, store/module version
maps, invariant results, and all query/transaction probes. Restart multiple
times, state-sync a fresh node, and perform an export/import round trip.

The rehearsal is blocking if it panics, violates an invariant, loses a store,
changes contract behavior unexpectedly, cannot restart/state-sync/export, or
cannot explain an app-hash difference.

### Gate D — RC testnet

- Publish reproducible binaries, checksums, SBOM/provenance, and multi-arch
  container digests from the tagged commit.
- Exercise mixed v30/v31 peers before the halt and uniform v31 validators
  after it.
- Run normal traffic plus fee, Wasm, DAO, IBC, PFM, clock, cw-hooks, and
  voting-snapshot probes for at least 72 hours.
- Confirm explorer, indexer, RPC, REST, gRPC, websocket, relayer, wallet, and
  `juno-std` consumers.
- Complete an independent security review of the dependency diff, upgrade
  handler, ante/post fee arithmetic, Wasm boundary, and unbounded iteration.

## Rollout checklist

1. Freeze scope after Gate C; only release-blocking fixes enter the RC branch.
2. Publish upgrade proposal text with exact height, binary/tag/checksums,
   container digest, minimum hardware guidance, and validator commands.
3. Give validators the notice required by `RELEASES.md`; collect readiness and
   rehearse the halt procedure.
4. State clearly that a committed state migration cannot be rolled back by
   simply restarting v30. Prepare a forward-fix binary and coordinated recovery
   procedure.
5. At the halt, verify >2/3 readiness, app hash agreement, peer health, and
   first-block execution before declaring success.
6. Monitor for at least 24 hours: consensus rounds/missed blocks, mempool and
   RPC errors, block time/gas, Wasm failures, IBC queues/timeouts, relayer
   health, FeePay/feeshare accounting, and voting-snapshot EndBlock duration.
7. Publish a post-upgrade report with evidence for every gate and file follow-up
   issues with owners and milestones.

## Explicit non-goals

- Cosmos SDK 0.54/0.55, store-v2, IBC-Go 11, or CometBFT 0.39 unless PFM and
  all operational gates become ready before implementation begins and the
  scope is deliberately re-approved.
- A private long-lived fork of PFM or another core ecosystem dependency.
- New consensus modules unrelated to a demonstrated v31 requirement.
- Treating unimplemented, skipped, or retry-green tests as release evidence.

## Definition of done

V31 is ready to propose only when the selected dependency graph is documented,
all P0/P1 items are resolved or explicitly waived with rationale, Gates A-D
are green, a recent mainnet-state rehearsal is reproducible, independent review
has no open critical/high findings, release artifacts are reproducible, and
validator/operator instructions have been tested verbatim.
