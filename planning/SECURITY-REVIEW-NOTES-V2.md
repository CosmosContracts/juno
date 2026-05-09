# Security review — second pass (post-Bucket-A+B)

This is the second internal review pass against the v30 dep-bump
branch, scoped to the **delta since `planning/SECURITY-REVIEW-NOTES.md`
was written.** It covers the Bucket A and B work added in
planning/09-deferred-work.md and the misc CI / Dockerfile / lint
hygiene that landed alongside.

**Scope:** everything between commits `fd6e08c3` (first review) and
the current branch tip on `juno-ai-dev/juno:juno-ai/v30`.

**Reviewer:** Juno AI (internal). External reviewers still required
for the upgrade handler and the new module per `planning/08`.

---

## Headline findings (delta)

- **Sev-1:** 0 (cumulative: 0)
- **Sev-2:** 0 (cumulative: 0)
- **Sev-3:** 1 new + 2 closed (cumulative: 1)
- **Info:** 6 new (cumulative: 13)

The two Sev-3s from the first pass closed:
- Sev-3 #1 (wasmbindings VotingPowerAt no fixed-cost gas) — closed
  by B2: explicit `ctx.GasMeter().ConsumeGas(5000, "wasmbindings/voting_power_*")`
  at the top of all three wasmbinding query helpers
- Sev-3 #2 (unbounded snapshot retention) — closed by B1: new
  `RetentionWindowHeights` param defaulting to ~1 year of blocks,
  EndBlocker pruning per-block

---

## §New module surfaces (B3 — proto/gRPC/REST)

`x/voting-snapshot` now exposes:
- gRPC `Query.{Params, VotingPowerAt, TotalVotingPowerAt, VotingPowerOverRange}` (read-only)
- gRPC `Msg.UpdateParams` (gov-only)
- REST gateway routes via `query.pb.gw.go`

Reviewed `keeper/grpc_query.go` and `keeper/msg_server.go`:

- ✅ `MsgUpdateParams` checks `msg.Authority != m.keeper.Authority()`
  and returns `codes.PermissionDenied`. Authority is plumbed from
  `keepers.go` as `govModAddress` (gov module account). Matches the
  pattern of every other governance-controlled module in the SDK.
- ✅ `Query.VotingPowerAt`/`TotalVotingPowerAt`/`VotingPowerOverRange`
  validate bech32 input where applicable and translate keeper errors
  to `codes.Internal`. Read-only — they only call keeper methods that
  iterate collections.
- ⚠️ **Info:** `Query.Params` returns `DefaultParams()` if Params is
  unset (e.g., a node that joined before InitGenesis ran). Matches
  the IsLST(ctx, addr) defensive read elsewhere; consistent.
- ⚠️ **Info:** No rate limiting / per-caller cost on
  `VotingPowerOverRange`. Range can be arbitrary; if a caller asks
  for `[0, max_int64]` the keeper iterates the entire delegator's
  history. Bounded in practice by the retention window (B1) and per-
  iteration store-read gas. Not a Sev-3 because the caller still
  pays gas, but flag for the wasmbinding wrapper review (B5a).

## §Pruning (B1) — new state-mutating EndBlocker

`x/voting-snapshot/keeper/prune.go` writes (deletes) state on every
block, gated by `RetentionWindowHeights`.

- ✅ Zero retention disables pruning entirely (verified by
  `TestPruneRetentionWindow` test).
- ✅ Cutoff math (`cutoff := height - int64(window)`) treats negative
  cutoffs as "no-op" (early `if cutoff <= 0 { return nil }`). Avoids
  the underflow case at chain start.
- ⚠️ **Info:** `pruneInterval` is hardcoded to 1 (every block).
  Comment says it's a knob for future batching. Per-block prune
  iteration is bounded by recent-snapshot count and is well within
  block-gas budget for any realistic Juno validator-set size, but if
  delegations spike (e.g., post-incentive launch), revisit.
- ⚠️ **Sev-3:** `pruneVotingPower` walks every (delegator, height)
  pair via `k.VotingPower.Iterate(ctx, rng)` with an unbounded range,
  then filters in Go memory. For a chain with millions of delegation-
  changing events accumulated over years, this is a long full-store
  scan every block. Should use a per-delegator prefix-aware range OR
  a per-block bookkeeping index. **Logged for v30.x** — does not
  break correctness in v30 since the retention window keeps row
  count bounded; performance issue at long-running scale.
- ⚠️ **Info:** Errors from Prune in EndBlock are silently swallowed
  (per the cosmos-sdk pattern for non-critical EndBlocker work). If
  Prune ever returns an error consistently, snapshots won't be
  pruned and storage grows unboundedly. Added a TODO comment in the
  module for future logging once the keeper grows a Logger() helper.

## §Per-validator slash re-snapshot (B4)

Hooks.BeforeValidatorSlashed iterates `stakingKeeper.GetValidatorDelegations`
and re-records each delegator's power.

- ✅ Bounded by the slashed validator's delegator count. Juno
  validators today have hundreds-to-low-thousands delegators; the
  iteration is well within block-gas headroom.
- ⚠️ **Info:** `recordDelegatorPower` for each delegator calls
  `stakingKeeper.GetDelegatorBonded(ctx, addr)` which itself iterates
  all of that delegator's delegations. So total cost of one slash
  hook is `O(validator_delegators × delegator_avg_delegations)`. In
  the worst case (many high-fan-out delegators), this could be
  meaningful. Worth a per-environment benchmark before mainnet halt-
  height.

## §VotingPowerOverRange (B5a)

New keeper method + wasmbinding + gRPC route. Reviewed for
consistency with the at-or-before semantics of VotingPowerAt.

- ✅ Returns rows in ascending height order (prefix iteration on
  `(delegator, height)` is naturally ascending).
- ✅ Inverted range (`fromHeight > toHeight`) returns nil cleanly.
- ⚠️ **Info:** Doc comment in `snapshot.go` explicitly warns callers
  that an empty result is *not* "zero power for the window" — could
  also mean "delegator didn't change stake during the window."
  Consumers must fall back to `VotingPowerAt(fromHeight)` for the
  baseline. Documented; trusting the caller-contract reads it.

## §Stream lazy refresh (B6)

`MethodRegistry()` is now lazy via `sync.Once`. Reviewed for
correctness under concurrent access.

- ✅ `sync.Once.Do` guarantees Refresh runs at most once per keeper
  instance, even under concurrent callers.
- ✅ The previous eager Refresh from NewKeeper is gone — first call
  to `MethodRegistry()` triggers it after module-manager init has
  fully wired bank / other gRPC handlers. Verified by the test
  passing without the explicit Refresh-after-Commit workaround.
- ⚠️ **Info:** If `Refresh()` returns an error on first call, it's
  logged but the registry remains empty. Subsequent calls won't
  retry. If the chain ends up in this state, all stream lookups
  fail silently. For v30 this is the same behavior as the old
  eager-Refresh path; documented but not a regression.

## §Feegrant test parallelism fix (B7)

Pure test-side change; the fix converts the `cases` map to a slice
and adds `s.SetupTest()` per subtest. No production-code changes.
No security implications.

## §Stargate path fix (B9)

`/cosmos.bank.v1beta1.Query/DenomMetadata` mapped to
`QueryDenomMetadataResponse` (singular) — was incorrectly mapped to
the plural type. Pre-existing bug, not v30-introduced. Fix is a one-
liner, no security implications.

## §CI infrastructure changes

- `make lint` invocation in `.github/workflows/golangci-lint.yml`
  replaces the standalone action. The Go-tool form (`go tool
  golangci-lint`) is built from go.mod's pinned v2.5.0 against the
  workflow's setup-go 1.25 toolchain. Cleaner version coupling, no
  security implications.
- `go test ./...` (parallel) restored in `build.yml` after B7 fix.

## §A1 DAO DAO ictest scaffold

`interchaintest/tests/dao-dao/dao_dao_test.go` ships the test
scaffold. No production-code surface. The test cases that aren't
yet implemented (`TestCw20StakedDao`, `TestWasmbindingsVotingPowerAt`)
call `t.Skip()` so CI doesn't fail on them. Not load-bearing.

## §What I did NOT review

- The interchaintest test runs themselves. ictest is blocked from
  inside SafeClaude's network namespace (see
  `planning/ICTEST-BLOCKER-V30.md`); CI runs handle this naturally.
- DAO DAO contract behavior under the new wasmvm v3 binary. Pending
  the testnet smoke described in `planning/ICTEST-BLOCKER-V30.md`
  and `planning/09-deferred-work.md` §A1.
- The proto-generated `.pb.go` files for x/voting-snapshot — these
  are output of `make proto-gen` and trusted as deterministic from
  the .proto source. Source is small + reviewed.

## §Updated open questions for external reviewer

Carrying forward from the first review with updates:

1. **Snapshot retention default.** Now set to `12_614_400` blocks
   (~1y at 2.5s). Closes Sev-3 #2 from the first pass. Open question
   becomes: is 1 year the right default, or should it follow the
   longest active governance proposal voting period × N?

2. **Pre-upgrade query semantics.** Unchanged from first pass: still
   returns zero (not error) for heights with no snapshot. Backfill
   covers `height >= upgrade_height`; pre-upgrade contracts shouldn't
   be querying historical power because the chain didn't have this
   primitive.

3. **wasmbindings gas accounting.** Closed by B2. Fixed cost of 5000
   gas per query is in line with comparable bank/staking ambient
   costs. External reviewer to gut-check the constant.

4. **Async-icq removal.** A3 confirmed zero ICQ counterparties on
   juno-1; closed.

5. **NEW: full-store scan in pruneVotingPower.** Sev-3 logged above.
   Per-delegator prefix iteration would be the cleaner shape but
   requires either knowing the delegator set up front or maintaining
   a "dirty heights" index. Defer to v30.x performance work.

6. **NEW: per-validator slash hook cost.** Bounded but worth
   benchmarking on a fork of juno-1 mainnet state with realistic
   slash event before halt-height.

## Sign-off (placeholder for external)

```
Internal (Juno AI): findings recorded inline above; cumulative open
   issues at this point: 0 Sev-1, 0 Sev-2, 1 Sev-3 (deferred to
   v30.x with reasons), 13 Info notes
- Reviewer A: not yet
- Reviewer B: not yet

Sev-2 deferrals: none
Sev-1 / open issues at sign-off: 0
```

External review still required for:
- `app/upgrades/v30/upgrades.go` (BackfillFromStaking is the new
  state-touching code there — the rest is unchanged from first pass)
- `x/voting-snapshot/` (entire module — gRPC + msg server + hooks +
  pruning all new code paths)
- `app/keepers/keepers.go` (ICA host/controller wiring + wasm IBC
  handler — unchanged since first pass, but external eyes on the
  29-fee removal cleanup are valuable)
