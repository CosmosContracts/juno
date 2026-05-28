# 08a — Security review findings (2026-05-28)

Output of the security review described in `08-security-review.md`, run on
branch `jakehartnell/v30` at HEAD `57141788`. Methodology: three parallel
LLM-driven reviews (deps + code + wasmbindings) plus `govulncheck`, with
every flagged line verified against source before reporting.

This file lists **what changed in this pass** (the easy fixes already
landed as code edits — see git diff against `57141788`) and **what's left
as a plan for the harder items**.

## 1. Fixes landed in this pass

All shipped as edits against the working tree; build/lint/tests green.

| File | Change | Rationale |
|---|---|---|
| `wasmbindings/message_plugin.go:163-165` | `errorsmod.Wrapf(nil, …)` → `errorsmod.Wrapf(sdkerrors.ErrUnauthorized, …)` | Old code returned nil → fell through into unguarded `SendCoins`, defeating the blocked-address check. Contracts could mint tokenfactory denoms straight into module accounts. |
| `wasmbindings/message_plugin.go:14` | Add `sdkerrors` import | Required by the line above. |
| `x/cw-hooks/keeper/msg_server.go:83-101` | Replaced `isContractSenderAuthorized` else-if chain with feeshare-style "admin if set, else creator" logic | Old chain required `sender == admin == creator` simultaneously. Any factory-instantiated contract (admin = DAO core, creator = factory) could not register for cw-hooks at all. |
| `x/cw-hooks/keeper/staking_hooks.go:258` | `ExecuteMessageOnContracts(...)` → `dispatchHookMessage(ctx, ..., "AfterValidatorBeginUnbonding")` | Every sibling hook uses `dispatchHookMessage` (which swallows contract errors). This one didn't — a bug in any registered contract could propagate into the staking unbonding queue and halt the chain. |
| `app/upgrades/v30/upgrades.go:71-78` | Guard `consensusParams.Block.MaxGas <= 0` before uint64 cast | `MaxGas == -1` (CometBFT "unbounded") would cast to 2^64-1 and seed feemarket AIMD with a nonsense ceiling. juno-1 has positive max_gas so this is operator-mode hardening for other deployments. |
| `app/ante/msg_filter.go:28-49` | `hasInvalidMsgs` now recurses into `authz.MsgExec.GetMessages()` | Blocked-msg list (currently just `MsgTimeoutOnClose`) was bypassable by wrapping in MsgExec. |
| `app/keepers/acceptedQueries.go:14-15,53-58` | Added `/cosmos.gov.v1.Query/Vote` entry; corrected existing `/cosmos.gov.v1beta1.Query/Vote` to use `govv1beta1.QueryVoteResponse{}` instead of `govv1.QueryVoteResponse{}` | Old code paired v1beta1 path with v1 response type — contracts querying gov votes via this binding decoded garbage. Now both paths are correctly typed. |
| `go.mod:478-486` | jwt-v4 replace `v4.4.2` → `v4.5.1` | Clears GHSA-29wx-vh33-7x7r (CVE-2024-51744). `v4.5.1` is chosen rather than `v4.5.2` to avoid a "same version, two paths" go.mod conflict with go-ethereum's direct require of `v4.5.2`. Both versions contain the CVE fix; `v4.5.2` is documentation-only. |
| `go.mod` (transitive) | `ulikunitz/xz v0.5.14` → `v0.5.15` | Clears GO-2025-3922 (LZMA memory leak), flagged reachable by govulncheck. |
| `.mise.toml:29` | Go `1.25.2` → `1.25.10` | Clears 15+ reachable stdlib advisories in `net/url`, `crypto/tls`, `crypto/x509`, `html/template`, `archive/tar`, `os`, `net`, `net/http` — all flagged reachable by govulncheck and all fixed in the 1.25.3 → 1.25.10 sequence. |
| `go.mod` (direct: `x/net`, `grpc`, `go-jose/v4`; transitive: `x/crypto`, `x/sys`, etc.) | `x/net v0.49.0 → v0.55.0`, `grpc v1.79.1 → v1.79.3`, `go-jose/v4 v4.1.3 → v4.1.4`, `x/crypto v0.47.0 → v0.51.0` (transitive) | Clears GO-2026-5026 (x/net), GO-2026-4762 (grpc), GO-2026-4945 (go-jose), and lifts x/crypto to current. |

### Net effect

- 6 source bugs fixed (1 ship-blocker `PerformMint`, 1 ship-blocker `cw-hooks` auth, 1 ship-blocker `acceptedQueries` proto mismatch, 1 chain-halt foot-gun, 1 operator hardening, 1 ante-filter bypass).
- ~21 reachable CVEs/GHSAs cleared by dep/toolchain bumps (xz, x/net, grpc, go-jose/v4, plus 15+ stdlib fixes from the planned Go 1.25.10 toolchain bump once `mise install` is re-run).
- 1 false positive identified and documented (GO-2024-2584).
- 2 advisories (GO-2026-4513, GO-2026-4740) remain reachable in `shamaton/msgpack/v2` — pulled in only via `wasmvm/v3` `types/`. Naturally clears when the wasmd v0.61.13 / wasmvm v3.0.6 coordinated bump in §2C lands.

### Still landing in this PR (TBD)

None — everything in §1 is the scope of this pass. The items in §2 below are the proposed follow-ups, broken out so they can be reviewed and merged independently.

## 2. Plan for harder items

### A. `voting-snapshot` slashing semantics

**Problem.** `BeforeValidatorSlashed` (the only slash hook the SDK fires globally) runs *before* `RemoveValidatorTokens` deducts the slash. `GetDelegatorBonded` inside that hook returns un-slashed power. A DAO vote opened in the same block as a slash carries a stale numerator for the slashed validator's delegators.

**Why the obvious fix is wrong.** Moving the snapshot to `BeforeDelegationSharesModified` doesn't work: that hook isn't fired during a slash. The SDK mutates `validator.Tokens` directly via `RemoveValidatorTokens`; `Delegation.Shares` are unchanged; per-delegator hooks fire only for unbondings/redelegations *in flight* across the slash, not the full delegator set.

**Recommended approach — defer-to-EndBlocker hybrid with marker fallback.**

1. In `BeforeValidatorSlashed`, append `(height, valAddr)` to a transient `PendingSlashes` queue and write `SlashMarker[height, valAddr] = fraction` to a persistent map.
2. In `EndBlocker` (new), drain the queue. By this point `RemoveValidatorTokens` has executed and `GetDelegatorBonded` returns post-slash values. For each `(height, valAddr)`:
   - If `len(GetValidatorDelegations) ≤ MaxDelegatorsPerSlash` (proposed `5000`): walk every delegator, call `recordDelegatorPower`, remove the `SlashMarker` entry.
   - If `len > MaxDelegatorsPerSlash`: keep the marker. The read path applies the slash fraction post-hoc.
3. `VotingPowerAt` reads gain a tail step: iterate the delegator's *current* validators (typical N=1–5), check `SlashMarker[height, valAddr]`, apply the fraction multiplicatively to the slashed validator's share-equivalent of the raw snapshot value.
4. Always call `recordTotal(ctx)` once after the drain so the denominator snapshot is post-slash.

**Why this shape.**
- Eager re-snapshot is the right default — read path stays simple for the 99% case.
- The overflow valve handles the whale-validator scenario without enforcing an eager walk that wouldn't fit a block (today's biggest juno-1 validator has ~6k delegators; the cap leaves headroom).
- The fallback path costs one map lookup per validator in the reader's portfolio — cheap relative to a tally read from a DAO contract.

**Files.**
- `x/voting-snapshot/types/keys.go` — add prefixes for `PendingSlashes`, `SlashMarker`
- `x/voting-snapshot/keeper/keeper.go` — declare the collections fields, wire them into NewKeeper
- `x/voting-snapshot/keeper/hooks.go:42-62` — rewrite `BeforeValidatorSlashed`
- `x/voting-snapshot/keeper/snapshot.go` — add `FlushPendingSlashes`, modify `VotingPowerAt` to apply `SlashMarker`
- `x/voting-snapshot/module/module.go` — add EndBlocker

**Open question.** Whether to constraint-tag `MaxDelegatorsPerSlash` (currently a `const`) as a Params field so it can be tuned post-launch. Recommend yes if the validator-cap delegate-count distribution shifts — but defer to v30.x. Hard-code for v30 ship.

### B. `voting_power_over_range` unbounded iteration

**Problem.** `wasmbindings/queries.go:102-118` calls `VotingPowerOverRange(del, from, to)` with a flat 5000 wasm gas charge. SDK store-read gas is the only ceiling on iteration size; a malicious contract can pass `from=0, to=MaxInt64` and force the keeper to iterate every snapshot ever written for a delegator.

**Recommended approach.**

```go
const (
    votingPowerBaseGas     uint64 = 1_000
    votingPowerPerRowGas   uint64 = 500
    votingPowerMaxRangeBlk int64  = 100_800 // ~1 week at 6s blocks
    votingPowerMaxRows     int    = 1024
)
```

Reject `to - from > votingPowerMaxRangeBlk` at the binding layer. Add a `VotingPowerOverRangeCapped` keeper variant that breaks the loop at `votingPowerMaxRows`. Charge `votingPowerBaseGas + len(rows) * votingPowerPerRowGas` after the iteration.

**Why these numbers.**
- 1 week covers realistic DAO proposal windows (DAO DAO defaults 3–7 days).
- 1024 rows handles daily restaking over ~3 years; passive delegators produce zero rows; pathological per-block restaking trips the cap.
- Gas multipliers chosen to match the magnitude of SDK `GetDelegation` (~500 gas per row analogue). 1024 rows ≈ 513k SDK gas, ≈71M wasm gas at the 140k multiplier — meaningful but not block-killing.

**No real-world consumer exists yet** (search of `/workspace/dao-contracts` returned only `VotingPowerAtHeight` callers), so the cap is forward-design, not a survey. Optionally expose `MaxRangeWidth` and `MaxRows` as Params for post-launch governance lift if a use case materialises.

**Files.**
- `wasmbindings/queries.go` — add cap + per-row gas
- `x/voting-snapshot/keeper/snapshot.go` — add `VotingPowerOverRangeCapped`

### C. wasmd `v0.61.13` + wasmvm `v3.0.6` coordinated bump

**Status.** wasmd v0.61.12 = wasmvm v3.0.5 bump. wasmd v0.61.13 = wasmvm v3.0.6 bump. Neither patch touches Go source on either repo; the substance is inside `libwasmvm` → `cosmwasm-vm`.

**Consensus risk.** v3.0.5 contains a gas-accounting overflow fix in `packages/vm/src/{backend.rs,environment.rs}` — `u64 +` replaced with `saturating_add`, plus a `checked_add → GasDepletion` guard. Old behaviour on adversarial overflow was panic (node halt); new behaviour is clean `OutOfGas` error. Honest contracts cannot reach this path. v3.0.6 over v3.0.5 is doc/comment churn and a `Decimal::sqrt` optimisation — no host-function semantics change.

**Recommendation.** Land as a **separate coordinated upgrade**, not bundled into v30. Rationale: the gas-overflow change is technically consensus-touching (panic→error boundary). Rolling it in its own window — with explicit validator coordination and a clean stop-the-world height — is safer than hiding it inside the v30 upgrade handler where rollback is painful. v30 ships first; the wasmd/wasmvm patch lands as v30.1 (or a no-handler coordinated stop) two to four weeks later once mainnet is stable.

**Test plan when we do bump.** Rebuild juno against wasmd v0.61.13; `make test && make test-race && make ictest-wasm`. On a private mainnet state-sync, replay the last ~100k blocks under the new binary and diff app-hashes against juno-1 at each height. Any divergence is the overflow path; grep mainnet history for any tx with `gas_used` approaching u64::MAX (none expected).

### D. GO-2024-2584 — false positive, document & suppress

**Finding.** govulncheck flags `cosmos-sdk v0.53.7` as affected by GO-2024-2584 ("Slashing evasion") with `Fixed in: N/A`. The underlying advisory is ASA-2024-005 / GHSA-86h5-xcpx-cfqc, severity Low, fixed in SDK `0.47.10` and `0.50.5`. The v0.53.x line forked from a `v0.50.x` baseline already past v0.50.5, so the fix is inherited; the vuln DB simply doesn't list v0.53 as "Fixed in" because no explicit v0.53 fix tag exists.

**Recommendation.** Add `.govulncheck.yaml` suppression with a one-line justification pointing back to this file, or accept the noise. The second reachable trace (`testutil → vesting`) is test-scaffolding only — `testutil/` is not imported from `app/` or `cmd/`, so the path is absent from the compiled `junod` binary.

### E. CometBFT v0.38 → v1.x migration

**Status.** v0.38.23 is the latest patch on the v0.38 line; we are fully up-to-date including CSA-2026-001 ("Tachyon", fixed v0.38.21). CometBFT's published support policy is "two latest release lines"; with v0.39 and v1.x both out, v0.38 is on borrowed time. No published EOL date.

**Recommendation.** Make the v0.38 → v1.x migration a tracked v31/v32 epic. Out of scope for v30. Single biggest medium-term consensus-layer risk in the dep tree.

### F. `bdpiprava/scalar-go` supply chain

**Status.** Direct dep (`app/endpoints/scalar.go:6`). Personal-namespace package, single maintainer, 66 stars, last code commits July 2024 (since then only dependabot churn). Serves the `/scalar` OpenAPI UI at the LCD port (1317 by default).

**Risk.** Supply-chain compromise could inject JS into operators visiting their own node's API explorer. Not exploitable on consensus path; not exploitable for operators who don't expose 1317 publicly. Real but low-severity.

**Recommendation.** One of:
1. Vendor `bdpiprava/scalar-go` into `app/endpoints/scalar-vendor/` and pin the commit hash; revisit only if upstream Scalar publishes an official Go binding.
2. Replace with a static HTML+CDN of upstream Scalar's TypeScript, served from `app/endpoints/scalar.go`. Loses the live-reload-from-OpenAPI flow but removes the dep entirely.
3. Accept the dep with a comment in `scalar.go`.

Defer the decision; not a v30 blocker. If picking (1) or (2), do it in a separate PR after v30 ships.

## 3. Items intentionally NOT actioned

Documented here so future passes don't re-flag them:

- **Gin replace pin at v1.9.0** — the original GHSA (h395-qcrw-5vmq) is fixed in v1.9.1; the pin is stale. Dropping the replace is cosmetic. Not actioned because the call surface is irrelevant (gin is not in junod's runtime path).
- **goleveldb replace pin** — verify whether SDK v0.53 still requires this; the comment says "v47 upgrade guide". Not actioned because no active CVE; would need an SDK-team confirmation.
- **iavl v1.2.6 → v1.3.5** — five minors behind, no published GHSA; stability rather than security. Not actioned in this pass.

## 4. govulncheck residual

After this pass's bumps land (jwt + xz + x/net + grpc + go-jose) and once `mise install` brings the dev/CI Go toolchain to 1.25.10, expected residual call-reachable advisories are:

- **GO-2024-2584** — the cosmos-sdk false positive documented in §2D.
- **GO-2026-4513 + GO-2026-4740** (`shamaton/msgpack/v2`) — pulled via `wasmvm/v3/types`. Clears when §2C wasmd/wasmvm bump lands.

That should leave **zero genuinely unfixable reachable findings**. Re-run before tagging to confirm.

## 5. Sign-off checklist for v30 ship

- [x] Source bugs in §1 landed
- [x] Toolchain + dep bumps in §1 landed
- [x] `make lint` 0 issues
- [x] `make test` green (touched packages)
- [ ] `make ictest-wasm` green (full integration)
- [ ] `make ictest-upgrade` green from v29.0.0 base (covers v29 → HEAD)
- [ ] govulncheck residual matches §4
- [ ] Two external reviewers signed off on upgrade handler
- [ ] Two external reviewers signed off on x/voting-snapshot
- [ ] §2A and §2B follow-up PRs filed (post-v30 if not bundled)
- [ ] §2D suppression filed
- [ ] §2C scheduled as a v30.1 coordinated upgrade
- [ ] uni-7 pre-flight in §6 passes; uni-7 upgrade observed clean before mainnet schedule

## 6. uni-7 testnet upgrade flow — pre-flight

uni-7 is the rehearsal for the juno-1 v30 upgrade. Verifying it goes
cleanly is the single most important non-source check before mainnet.

### 6.1 Dep delta magnitude

v29.0.0 → v30 is the largest single-upgrade delta in Juno's history:

| Dep | v29.0.0 | v30 (HEAD) | Risk |
|---|---|---|---|
| `cosmos-sdk` | v0.50.13 | v0.53.7 | One full minor (v0.50 → v0.53). Multiple module ConsensusVersion bumps along the path. |
| `ibc-go` | v8.7.0 | v10.6.0 | **Two majors (v8 → v10).** v9 reshaped client-state types; v10 added IBC-eureka. Each major has consensus-breaking store migrations. |
| `wasmd` | v0.54.0 | v0.61.11 | Seven minors. Includes the wasmd-half of the wasmvm v2 → v3 transition. |
| `wasmvm` | v2.2.4 | v3.0.4 (`/v3` import path) | **Major.** CGO `libwasmvm.so` swap required on every validator host — common operator footgun. |
| `cometbft` | v0.38.17 | v0.38.23 | Patch-only, low risk. |

The `ictest-upgrade` suite spins a single juno chain from `v29.0.0` →
HEAD and verifies cw-hooks survives. It does **not** establish IBC
connections or ICS-27 ICA channels, so it will not surface bugs in:

- ibc-go v8 → v10 client/connection/channel store migration with active
  packets in flight
- residual ICS-29 `feeibc` state on channels that had fees enabled
  (the store is purged via `StoreUpgrades.Deleted` — any unresolved
  fee escrow disappears)
- residual `interchainquery` state for any registered ICQ queries

uni-7 is exactly the place to exercise these. Before scheduling the
upgrade proposal, take a snapshot of currently-active IBC connections
and ICQ queries from a uni-7 node:

```bash
junod q ibc connection connections --node <uni-7-rpc> --output json | jq '.connections | length'
junod q ibc channel channels        --node <uni-7-rpc> --output json | jq '.channels | length'
junod q interchain-query queries    --node <uni-7-rpc> --output json 2>/dev/null | head
```

Any non-empty result is a deliberate test surface — leave those
connections live across the upgrade height and re-verify them after.

### 6.2 max_gas precondition

`configureFeemarketParams` reads `consensusParams.Block.MaxGas` and
the new §1 guard halts the upgrade handler if it is `<= 0`. uni-7's
genesis sets `max_gas = 100000000` (verified by inspecting
`testnets/uni-7/genesis.zip`), so the genesis-state default is safe.

Risk: a consensus-params change proposal could have set it to `-1`
in the years since genesis. **Verify live state before scheduling:**

```bash
junod query consensus params --node <uni-7-rpc> --output json | jq '.params.block.max_gas'
# Must be a positive int. "-1" = upgrade will halt at handler.
```

If the live value is non-positive, queue a consensus-params proposal
to fix `max_gas` first, with a voting period that completes before
the v30 upgrade height.

### 6.3 Store-purge correctness (no source change needed)

`StoreUpgrades.Deleted` lists: `globalfee`, `crisis`, `params`,
`nft`, `feeibc`, `interchainquery`. Cross-checked against v29.0.0's
keys.go — every one of these is a real KV store currently allocated
on uni-7 (v29 baseline). Cross-checked against upstream `StoreKey`
string constants — every entry matches:

- `x/crisis` → `"crisis"` ✓
- `x/params` → `"params"` ✓
- `cosmossdk.io/x/nft` → `"nft"` ✓
- `ibc-go/v8/modules/apps/29-fee` → `"feeibc"` ✓
- `async-icq/types` → `"interchainquery"` ✓
- `x/globalfee` → `"globalfee"` ✓

uni-7 genesis spot-check confirms all six modules had data at chain
start. Purge will run as expected.

`StoreUpgrades.Added`: `feemarket`, `votingsnapshot`. Both new in v30,
no prior data — correct.

### 6.4 BackfillFromStaking — runs twice (latent, non-blocking)

The new `x/voting-snapshot` module's `InitGenesis` already calls
`BackfillFromStaking(ctx)` (see `x/voting-snapshot/keeper/genesis.go:22`).
`RunMigrations` triggers that `InitGenesis` for new modules during
the upgrade handler.

Then `app/upgrades/v30/upgrades.go:50` calls
`BackfillFromStaking(ctx)` **again**, explicitly.

Both calls run at the same block height, so the second
`VotingPower.Set` and `TotalPower.Set` overwrite the first with
identical values. Not a correctness bug — both are wasted O(n)
work on top of an already-O(n_delegations) BeginBlock at upgrade
height.

uni-7 has ~tens of validators and probably hundreds of delegators,
so the cost is invisible. juno-1 has ~50 active vals and ~tens of
thousands of delegators; double-work is still in the seconds-not-
minutes range but unnecessary.

**Recommended cleanup (non-blocking for uni-7):** drop the
explicit handler call, leaving the `InitGenesis` call as the
single backfill site. Save for a v30.1 cleanup if not bundled now.

### 6.5 wasmvm v3 binary-swap operator instructions

When validators swap the junod binary at the upgrade height, they
**must also swap `libwasmvm.so`** from the v2 series (`libwasmvm.x86_64.so`
shipped alongside `wasmvm v2.x`) to the v3 series (shipped alongside
`wasmvm v3.0.4`). The Go module path is now `wasmvm/v3` and the
shared library is ABI-incompatible with v2.

`make build` on the v30 binary will fail at link time without the v3
`.so`. But a worse failure mode: validators who copy the v30 binary
onto a host that already has `libwasmvm.x86_64.so` from v2 will see
the binary start, then panic at first wasm execution post-upgrade
(typically on the first contract tx after upgrade height) with
either a CGO ABI mismatch or `undefined symbol` errors.

**Add to the uni-7 upgrade announcement** (and reuse for juno-1):

> **Validator action required at upgrade height:**
>
> 1. Stop junod
> 2. Replace `junod` binary with `v30.0.0`
> 3. Replace `libwasmvm.x86_64.so` (or `libwasmvm.aarch64.so` on ARM)
>    with the version bundled in the v30.0.0 release tarball.
>    Old path is typically `/usr/lib/libwasmvm.x86_64.so` or
>    `$HOME/lib/libwasmvm.x86_64.so` — check with `ldd $(which junod)`.
> 4. Start junod
>
> If post-upgrade you see `undefined symbol: wasmvm_*` or a CGO panic
> on first contract execution, the `.so` was not swapped.

### 6.6 ibc-go v8 → v10 module ConsensusVersion gap

`mm.RunMigrations` is the load-bearing call. It reads the prior
ConsensusVersion map from `x/upgrade` store and steps each module
forward to the version reported by its current `AppModule.ConsensusVersion()`.
For the modules that crossed two majors (ibc-go's 02-client, 03-connection,
04-channel, ICA-host, ICA-controller), the migration registry inside
ibc-go itself owns the per-version migration functions.

Risk surface: if any Juno-side wiring forgot to register one of the
new keepers' AppModule via `module.NewManager(...)`, RunMigrations
will skip migrations for that module and the chain will halt at the
first state read that expects the new format. Verified in `app/modules.go`:
all ibc-go modules are registered in `NewManager` (`ibc.NewAppModule`,
`ibctransfer.NewAppModule`, ICA host + controller, packet-forward).

No source change needed, but **the uni-7 dry-run is the only place
this gets exercised end-to-end with real IBC packet history**.
Block out a validator on uni-7 to spam IBC transfers from before
upgrade height into after, and confirm packets that crossed the
boundary either acked or timed-out cleanly.

### 6.7 Recommended uni-7 dry-run timeline

1. **D-7**: re-tag candidate as `v30.0.0-rc1`; cut Docker image with
   the matching `libwasmvm.so`. Publish release notes including
   §6.5 operator instructions.
2. **D-5**: post upgrade proposal on uni-7 with `--upgrade-height` ≈
   D+1 block estimate. Voting period 12h (uni-7 genesis setting).
3. **D-1**: verify §6.2 max_gas live value, §6.3 expected purge,
   §6.1 snapshot of IBC/ICQ state.
4. **D**: upgrade height. Watch for the §1 guard panic (max_gas),
   the §6.5 CGO panic (validators who missed `.so` swap), and any
   IBC migration halt. If chain produces blocks for >10 min post-
   upgrade with normal tx throughput, the rehearsal passed.
5. **D+2**: replay §6.1 IBC/ICQ queries against new chain; verify
   the §6.4 backfill seeded sensible per-delegator power via
   `junod q voting-snapshot voting-power <addr> <upgrade-height>`.
6. **D+7**: if no anomalies, schedule juno-1 mainnet proposal.
