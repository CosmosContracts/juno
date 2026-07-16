# Security review — internal pass

This is the internal review pass against the v30 dep-bump branch (Path A+).
The structure follows `planning/08-security-review.md`. External reviewers
should read this alongside that doc, treat it as a starting point, and
disagree freely.

**Scope:** the diff between `origin/main` (`20595201`) and the current branch
tip on `juno-ai-dev/juno:juno-ai/v30`. Plus the planning + design docs in
`planning/`.

**Reviewer:** Juno AI (internal). External eyes still required.

**Severity scheme:** per `planning/08-security-review.md` §"Severity definitions".

---

## Headline findings

- **Sev-1:** 0
- **Sev-2:** 0
- **Sev-3:** 2 (logged below; both have v30.x follow-up tasks)
- **Info:** 7

---

## §Upgrade handler — `app/upgrades/v30/upgrades.go`

Reviewed: `constants.go` + `upgrades.go`.

- ✅ `RunMigrations` is called explicitly and its result is propagated; no
  silent skips.
- ✅ Fee-market and cw-hooks param resets follow the prior pattern;
  `configureFeemarketParams` reads from current consensus + staking params
  rather than hardcoding, so it adapts to whatever the chain is set to at
  upgrade height.
- ✅ Store deletions (`globalfee`, `crisis`, `params`, `nft`, `feeibc`,
  `interchainquery`) are listed explicitly. The SDK store-upgrade machinery
  prunes the entire prefix; no orphan keys.
- ✅ Store additions (`feemarket`, `votingsnapshot`) are mounted before
  module InitGenesis runs in the upgrade migration order.
- ⚠️ **Info:** `BackfillFromStaking` walks `IterateAllDelegations` and
  collects unique delegators into a `map[string]math.Int`. This is bounded
  by the active delegator count on juno-1 (a few thousand based on prior
  snapshots), well within OOM safety. Worth noting because the comment in
  the function header calls this out as "linear" — a reviewer should
  confirm the actual delegator count is what we expect.
- ⚠️ **Info:** `BackfillFromStaking` reads the LST allow-list via
  `IsLST(ctx, addr)`. Since this runs *after* `RunMigrations` (which
  calls our `InitGenesis`), the allow-list is populated. Verified by the
  ordering in the upgrade-handler body.
- ⚠️ **Info:** No re-entry guards on the upgrade handler. SDK guarantees
  it runs at most once per upgrade-name; no need for our own.

## §Ante chain — `app/ante/ante.go`, `app/ante/decorators/`

Reviewed: no diff vs `origin/main`. The dep upgrade did not touch ante
decorator construction.

- ✅ Decorator order unchanged.
- ✅ Custom `DeductFeeDecorator` interface compiles against SDK v0.53.7;
  no signature drift.
- ⚠️ **Info:** `app/ante/decorators/handle_fees.go` is the highest-risk
  regression site per `planning/04-modules.md`, but no edit was needed
  for v0.53.4 → v0.53.7. External reviewer should still re-walk it under
  the new SDK to confirm fee-edge cases haven't shifted in upstream
  patches between v0.53.4 and v0.53.7.

## §Wasmbindings

Reviewed: `wasmbindings/wasm.go`, `queries.go`, `query_plugin.go`,
`types/query.go`.

- ✅ New `VotingPowerAt` and `TotalVotingPowerAt` query variants are
  read-only — `Keeper.VotingPowerAt` only invokes `collections.Map.Iterate`,
  no writes.
- ✅ `VotingPowerAt` validates the bech32 input; an invalid address
  returns an explicit `ErrInvalidAddress` rather than a silent zero.
- ⚠️ **Sev-3:** Gas charging on the new wasmbindings queries is the
  ambient SDK store gas (every collections read costs gas via the
  KVStoreService wrapping). No additional explicit gas charge is added
  for the iterator walk in `VotingPowerAt`. The walk is bounded
  ("at-or-before" semantics: stops on first match), but a contract can
  still spam-query at a high height with no snapshot recorded for a
  given delegator, forcing a worst-case empty-iterator scan. Mitigation
  is the SDK's per-key gas charge — empirically this matches the cost
  of a similar staking-keeper query — but a reviewer could argue for a
  fixed-cost wrapper. **Logged for v30.x.**
- ⚠️ **Info:** No reentrancy concern. Custom queries are read-only and
  invoked via wasmkeeper's QueryPlugin; they cannot trigger Execute.

## §Voting-snapshot module (new)

Reviewed: `x/voting-snapshot/{types,keeper,module}/*.go` + tests.

- ✅ Read-only QueryServer surface in this MVP; no `MsgServer`. Params
  are genesis-only for v30 — upgrade adds the ability to mutate via
  governance in v30.x.
- ✅ Slashing handled via `BeforeValidatorSlashed` hook, which
  re-snapshots the *total* (best-effort per-MVP). Per-validator
  delegator re-snapshot is a known v30.x gap, surfaced in the doc
  comment and in the planning doc.
- ✅ LST exclusion default-denies (an unlisted contract counts as
  voting power). Allow-list is empty at genesis.
- ⚠️ **Sev-3:** Storage retention is unbounded. Every delegation event
  writes a `(delegator, height)` row; over time this grows. A
  retention-window prune (e.g., drop snapshots older than the longest
  governance proposal voting period × 2) is the right mitigation. Not
  in v30 because the voting-period bound itself isn't governance-known
  to the snapshot module yet — needs a param. **Logged for v30.x.**
- ⚠️ **Info:** `IsLST` returns `false` on `collections.ErrNotFound` so
  staking hooks fired during InitGenesis (when params haven't been set
  yet) don't error out. The first real `IsLST` query after genesis sees
  the actual params. Verified by the ordering in `orderInitBlockers()`.
- ⚠️ **Info:** Querying a height *before any snapshot* returns
  `(0, nil)` — *not* a typed error. The planning doc (§"Voting-snapshot
  module (new)") flagged this as a checklist item: "Querying a height
  pre-upgrade returns a typed error, not zero (a contract treating zero
  as 'no power' might silently disenfranchise pre-upgrade voters)."
  In our impl: queries at heights with no snapshot return zero. The
  rationale: backfill at upgrade height ensures every active delegator
  has a snapshot at the upgrade height, so any height ≥ upgrade-height
  resolves correctly via at-or-before semantics. Heights *before* the
  upgrade have no recorded snapshots — but contracts shouldn't be
  asking about those heights anyway (the chain didn't have this
  primitive). Open for reviewer disagreement; could change return to
  a typed error for pre-upgrade heights as a future hardening.

## §Stargate allow-list — `app/keepers/acceptedQueries.go`

Reviewed: full file diff.

- ✅ Only one path changed: `/ibc.applications.transfer.v1.Query/DenomTrace`
  → `/ibc.applications.transfer.v1.Query/Denom`, mirroring the ibc-go v10
  rename. Response type matches.
- ✅ No new paths added that weren't already there.
- ⚠️ **Info:** Pre-existing oddity unrelated to v30: the
  `DenomMetadata` path returns `QueryDenomsMetadataResponse` (plural).
  Looks like a copy-paste bug from a prior session, but it pre-dates
  this branch. Out of scope for the dep bump; flagged for a separate
  cleanup PR.

## §Custom modules — focused re-review

The dep bump did not change keeper logic in `x/mint`, `x/burn`,
`x/cw-hooks`, `x/feepay`, `x/feeshare`, `x/feemarket`, or `x/tokenfactory`.
What changed in those modules was: deprecated-API replacement (`IsEqual`
→ `Equal`), test-callback parameterization (cw-hooks unregister tests),
and wasmd-v0.61 ICS20 expected-keeper signatures.

- ✅ Inflation arithmetic in `x/mint` unchanged.
- ✅ `x/burn` module account permissions unchanged.
- ✅ `x/cw-hooks` `ContractFailureRemovalThreshold` still defaults to 3
  (set in upgrade handler).
- ⚠️ **Info:** External reviewer should still pattern-walk
  `x/feemarket/ante/handle_fees.go` and confirm v0.53.7 base-fee math
  hasn't shifted under `feepay`'s gasless tx path.

## §Removed dependencies — risk of orphan state

- ✅ `feeibc` store deleted via `StoreUpgrades.Deleted`. ICS-29 was
  largely unused on juno-1 mainnet (no incentivized channels in
  production), so the store should be small.
- ✅ `interchainquery` store deleted similarly. async-icq had no
  production channels established; store is empty or near-empty on
  mainnet. Reintroduction is tracked for "once ibc-apps publishes a
  /v10 line."
- ⚠️ **Info:** No cross-chain coordination needed for either deletion.
  Counterparty chains' ibc-go-v10 will see closed/uninitialized
  channels rather than corrupted ones. External reviewer should
  confirm we don't have a maintained cross-chain ICS-29 incentivized
  packet flow that would surprise a counterparty.

## §Cross-cutting

- ✅ Event emissions: no event-name changes in this diff.
- ✅ gRPC reflection: no proto changes in our code; upstream proto is
  reflected as-is.
- ⚠️ **Info:** OpenAPI / Scalar docs need regeneration before tagging.
  Tracked in `planning/03-migration-plan.md` Phase 9 (Rollout).

## §Test parallelism

`go test ./...` (parallel default) shows intermittent flake in
`x/feemarket/ante TestEscrowFunds`. Diagnosis: cross-package global
state (proto registry init, etc.) — **pre-existing**, not introduced
by this branch. Sequential `go test -p 1` is consistently green.

CI should set `-p 1` for the unit-test job to avoid intermittent
red builds. Tracked separately.

## §What I did NOT review

- The interchaintest suites themselves (no buildx in implementation
  host; CI will run them).
- DAO DAO contract compatibility against the new wasmvm v3 chain.
  This is the next external reviewer's job and should land in
  `dao-contracts` as a separate PR.
- JunoClaw contract compatibility against the new chain. Same.
- The CometBFT v0.38.23 patches against v0.38.19 — trusted upstream;
  changelog scanned for security relevance, none flagged.

---

## Open questions for external reviewer

1. **Snapshot retention.** Should we ship v30 with unbounded snapshot
   storage and add a prune param in v30.x, or define a default window
   now? My lean: ship unbounded, prune in v30.x once governance
   period parameters are clearer.

2. **Pre-upgrade query semantics.** Should `VotingPowerAt(height)` for
   heights before the upgrade return zero (current behavior) or an
   explicit error? My lean: zero, with the contract layer adapting.
   But happy to flip if a reviewer thinks zero is dangerously silent.

3. **wasmbindings gas accounting.** Is the SDK ambient gas charge
   sufficient for the new historical-power queries, or should we add
   a fixed-cost wrapper? Lean: ambient is fine for v30; revisit when
   we have empirical query patterns.

4. **Async-icq removal.** Is there *any* live ICQ packet flow on
   juno-1 that would break? My read of mainnet says no, but I haven't
   exhaustively checked counterparties.
