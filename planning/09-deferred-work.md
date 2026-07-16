# 09 — Deferred work

Things this PR explicitly does *not* do, with a plan for picking each
one up. Sequenced into three buckets: blockers (must close before v30
mainnet halt-height), v30.x (post-tag patches, not-blocking), and v31
(next consensus break).

## Bucket A — Blockers (close before v30 mainnet halt-height)

### A1. DAO DAO v2.7.0 contract compatibility on wasmvm v3

**Why blocking.** Juno mainnet's biggest live contract surface is DAO
DAO. wasmvm v2 → v3 is consensus-breaking; upstream's compat statement
says cosmwasm-std v1 + v2 contracts continue to run, but "should
continue to run" is not "verified to run on the specific contracts our
chain has." If a popular DAO breaks at v30 height, it's a userspace
catastrophe even if the chain itself is healthy.

**Scope.** The DAO DAO module set: core (`dao-dao-core`), voting
(`cwd-voting-cw20-staked`, `cwd-voting-cw4`, `cwd-voting-token-staked`,
`cwd-voting-cw721-staked`, the new staked-juno path that consumes the
voting-snapshot bindings), proposal modules (`cwd-proposal-single`,
`cwd-proposal-multiple`, `cwd-proposal-condorcet`), pre-propose modules
(`cwd-pre-propose-single`, `cwd-pre-propose-approval-single`),
`cw-filter`, `cw4-group`, the cw20/cw721 building blocks. v2.7.0 is
the current mainnet release per `memory/2026-05-08.md` reconnaissance.

**Concrete plan.**

1. **Scaffold `interchaintest/tests/dao-dao/`** — new ictest suite. Runs:
   - Spin up a local v30 chain
   - Deploy current `dao-dao` workspace-optimized wasm artifacts
     (built from `dao-contracts` at the v2.7.0 tag)
   - Instantiate a DAO with cw4-group voting + proposal-single (cheapest,
     covers core happy path)
   - Open + vote + execute a proposal that does nothing
     (`ProposalAction::Nothing`)
   - Open + vote + execute a proposal that mints a tokenfactory denom
     (covers the `BankMsg::Send` + custom-bindings interaction)
   - Tear down and re-instantiate with cw20-staked voting
     (covers a heavier path)
   - Tear down and re-instantiate with the new staked-juno voting module
     (covers `wasmbindings::VotingPowerAt` + `TotalVotingPowerAt`)

2. **Mainnet snapshot replay** — separate exercise. Take a recent juno-1
   snapshot, extract the wasm-state subset, instantiate it on a fresh
   v30 devnet, and run a no-op execute against the live DAO DAO core
   contracts. This catches what synthetic ictest can't (state shape,
   storage layout, indirect compatibility). Requires a state-export
   utility — basic version is `junod export | jq` against a halted node.

3. **Wasmbindings smoke** — invoke `VotingPowerAt` from a small
   single-file test contract (compiled from a Rust scratch project)
   and verify the response shape (`{"power": "12345"}`) round-trips
   through DAO DAO's StdResult layer. This catches the wasm-side
   adapter that lives in `dao-contracts/packages/juno-bindings`.

**Owner.** This crosses repo boundaries: ictest scaffold lives in
`juno/interchaintest/tests/dao-dao/`, the wasm-binding adapter lives in
`dao-contracts/packages/juno-bindings`. Either Juno AI does a draft and
the DAO DAO team reviews, or Jake routes the request directly through
the dao-contracts maintainers. The ictest scaffold itself can land in
this repo without dao-contracts cooperation; the binding adapter can't.

**Blocking criterion.** Before v30 halt-height proposal goes to
mainnet governance: all three legs above pass on a v30-rc1 testnet
deploy. If any DAO DAO contract instantiation fails or any vote tally
disagrees with the live contract's prior behavior, that's a Sev-1.

### A2. `ictest-upgrade` against forked mainnet state

**Why blocking.** This is the highest-signal test we have. It forks
juno-1 state into a Docker chain, runs the v30 upgrade handler, and
verifies post-upgrade chain liveness. CI runs this on push, but it
hasn't run yet because the PR isn't open against upstream.

**Concrete plan.** Open the upstream PR; let CI run the matrix. If
`ictest-upgrade` fails, the failure is in the upgrade handler,
in the new module's InitGenesis, or in the store-deletion list.
Fix forward.

### A3. Async-icq counterparty audit

**Why blocking.** v30 deletes the `interchainquery` store. If any
counterparty chain has an active ICQ packet flow open against juno-1,
the upgrade-height transition will leave them with closed channels
and stuck packets. Need to confirm none exist before halt-height.

**Concrete plan.** Query juno-1 mainnet via `junod q ibc channel
channels` filtered by port-id `icqhost`. Cross-reference with major
counterparty chains' channel state. If there's a live channel, either
hold the v30 proposal for an `ibc-apps`/`v10` async-icq release, or
coordinate a graceful channel close before halt-height.

**Probable outcome.** Per `memory/cosmwasm-ecosystem-state.md`, async-icq
on Juno was never load-bearing. Expect zero live channels. But verify.

### A4. CometBFT v0.38.19 → v0.38.23 changelog scan

**Why blocking-ish.** Three patch releases between our prior pin and
Path A+'s pin. CometBFT changelog should be reviewed for any consensus-
or networking-relevant fixes that change validator behavior. None
flagged in `planning/SECURITY-REVIEW-NOTES.md` in passing, but a
deliberate scan is owed.

**Concrete plan.** Read CometBFT's `CHANGELOG.md` between `v0.38.19`
and `v0.38.23`. Note anything tagged `[CONSENSUS]`, `[NETWORKING]`, or
`[BUG FIX]`. Land the result as a comment block in the v30 upgrade
proposal text on Commonwealth.

## Bucket B — v30.x (post-tag patches, not blocking)

### B1. x/voting-snapshot — snapshot retention pruning

**Why deferred.** The keeper writes a snapshot on every delegation
event with no retention policy. Storage grows linearly with
delegation activity over time. Two-year horizon is fine; ten-year
is not.

**Concrete plan.** Add a `RetentionWindowHeights` param to
`x/voting-snapshot/types.Params`. Default to a sensible value (e.g.
1 year of blocks ≈ 12_614_400 at 2.5s blocks). Add an EndBlocker
hook that prunes `VotingPower` entries older than the window. Add
a governance `MsgUpdateParams` so the window is tunable.

**Severity.** Sev-3 per the security review. Not user-facing, no
correctness impact within the retention window.

### B2. x/voting-snapshot — fixed-cost gas wrapper for wasmbindings

**Why deferred.** Currently `VotingPowerAt` charges only the SDK
ambient store gas. For pathological queries (high height, no
snapshot for the delegator), the iterator walks empty storage with
no fixed-cost ceiling.

**Concrete plan.** Wrap `GetVotingPowerAt` and `GetTotalVotingPowerAt`
in `wasmbindings/queries.go` with an explicit `ctx.GasMeter().ConsumeGas(N, "voting_power_query")` charge sized to bound a DDoS-shaped contract pattern.

**Severity.** Sev-3. Unlikely to be exploitable in practice given
contract gas limits, but principled.

### B3. x/voting-snapshot — proto + gRPC + REST

**Why deferred.** v30 ships wasmbindings as the only delivery
surface. CLI / REST / gRPC consumers (block explorers, indexers,
scripts) can't query historical voting power directly.

**Concrete plan.** Define proto types in
`proto/juno/votingsnapshot/v1/{query,types}.proto`. Run `make
proto-gen`. Implement `QueryServer` returning the same data as
the wasmbinding. Register in `module.RegisterServices`. Add CLI
subcommands.

### B4. x/voting-snapshot — per-validator delegator re-snapshot on slash

**Why deferred.** `BeforeValidatorSlashed` currently only re-records
the chain-wide total. Individual delegators under the slashed
validator have stale per-delegator snapshots until their next
delegation event.

**Concrete plan.** In `Hooks.BeforeValidatorSlashed`, iterate all
delegations under `valAddr` and re-record each delegator's power.
Adds a write per delegator per slash; bounded by the slashed
validator's delegator count.

### B5. x/voting-snapshot — VotingPowerOverRange + per-validator filter

**Why deferred.** Future-proofing for time-decay voting schemes
(plural voting, conviction voting) and validator-collective sub-DAOs.
Not used by DAO DAO v2.7.0.

**Concrete plan.** Add `VotingPowerOverRange(addr, fromHeight,
toHeight)` returning a slice of `(height, power)` pairs.
Add an optional `validator` filter parameter to `VotingPowerAt`.

### B6. x/stream — MethodRegistry refresh timing

**Why deferred.** The keeper's gRPC method registry is built during
`NewKeeper`, before module manager init wires bank's hybrid
handlers. The unit test was patched to call `Refresh()` after first
block, which matches production runtime, but means the registry is
empty for the first block of any chain run.

**Concrete plan.** Move the initial `Refresh()` call from `NewKeeper`
into a finalize-block hook (or a one-shot `BeginBlock(height==1)`
trigger). Verify the test fix can then be reverted.

### B7. Test parallelism — feemarket ante flake

**Why deferred.** `go test ./...` (parallel) shows intermittent
failures in `x/feemarket/ante TestEscrowFunds` due to cross-package
global-state leakage (proto registry init, etc.). `-p 1` sequential
is consistently green.

**Concrete plan.** Identify the global state. Likely a proto
descriptor registration race or an `init()` that mutates a process-
global. Fix at the source rather than papering over with test
isolation.

### B8. CI — switch unit-test job to `-p 1`

**Why deferred.** Until B7 is fixed, CI parallel runs will be
flaky. Sequential is slower but deterministic.

**Concrete plan.** In `.github/workflows/build.yml`, the `test` job:
`go test -p 1 ./...`. Or move the affected package to its own job.

### B9. wasmbindings — bank `DenomMetadata` proto-path mismatch

**Why deferred.** `app/keepers/acceptedQueries.go` maps
`/cosmos.bank.v1beta1.Query/DenomMetadata` → `QueryDenomsMetadataResponse`
(plural type for a singular path). Pre-existing bug, not v30-introduced.

**Concrete plan.** One-line fix:
`/cosmos.bank.v1beta1.Query/DenomMetadata` →
`QueryDenomMetadataResponse`. Land as a small follow-up PR.

### B10. async-icq reintroduction (once `ibc-apps` publishes `/v10`)

**Why deferred.** `ibc-apps/modules/async-icq/v10` doesn't exist as
a published module. Latest commit on `/v8` (April 2026) still pins
ibc-go v8.

**Concrete plan.** Watch the ibc-apps repo for a v10 publish.
When it lands, re-add the keeper, store, and module wiring; v30.x
patch release with a new store mount via a v30.1 upgrade.
Counterparty coordination (channel reopen) is the user-side cost.

## Bucket C — v31 (next consensus break)

### C1. Cosmos SDK v0.54 family + IBCv2 + store/v2

The originally-targeted Path B from `planning/02-targets.md`. Ships
when:

- `ibc-apps` publishes a `/v11` line for PFM, ibc-hooks, async-icq
- We've absorbed v30 in production for a release cycle
- Indexer / explorer team confirms readiness for store/v2

Likely 6–9 months out. Sequenced as v31 to avoid bundling a
high-risk store migration with a higher-risk dep upgrade.

### C2. cometbft v0.39+ (libp2p networking optional)

Bundled with C1 since the SDK + IBC + cometbft tend to land together.

### C3. `module.AppModule` → `appmodule.AppModule` migration

Currently silenced via `//nolint:staticcheck` in `app/modules.go` and
`x/voting-snapshot/module/module.go`. Cosmos-SDK upstream + most
ecosystem modules will migrate together; we follow.

## How this list gets used

- Bucket A items are PR-blockers — they have to close before tagging
  v30.0.0.
- Bucket B items are GitHub issues filed against this repo with the
  `v30.x` milestone. Each one is small enough to ship as a single
  PR.
- Bucket C is `v31` design work — kept here so the v30.x maintainers
  remember why certain things are silenced or deferred.

This file gets pruned as items close. An empty Bucket A is the gate
for v30 mainnet halt-height.
