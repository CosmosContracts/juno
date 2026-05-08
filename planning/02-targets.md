# 02 — Target stack

Two viable target paths. They differ in ambition, not in shape.

## Path A — Conservative (stay on the v0.53 line)

Hold the SDK at the v0.53 line, take the latest patches, jump wasmd to v0.61 to get wasmvm v3.

| Component | From | To |
|---|---|---|
| cosmos-sdk | v0.53.4 | **v0.53.7** |
| wasmd | v0.54.2 | **v0.61.11** |
| wasmvm | v2.2.4 | **v3.0.4** |
| ibc-go | v8.7.0 | **v8.8.0** |
| cometbft | v0.38.19 | **v0.38.23** |
| store | v1.1.2 | v1.1.2 (stay on v1) |

**Why this might be the right call:**

- Smallest surface change. Re-validates with maximum confidence.
- The wasmvm v2 → v3 jump alone is a meaningful upgrade (BN254 precompile lands here) and is enough to ship as `v30` even without the SDK minor bump.
- Keeps Juno on a release line that ibc-go v8 + ICA v8 already match. No middleware re-stacking.

**What's lost:**

- We re-do this work in 6–12 months when v0.53 goes EOL. The "stale branch" pattern repeats.
- Doesn't get the v0.54 throughput / stability improvements upstream advertises.
- store/v1 is increasingly the legacy path; future bumps get harder.

## Path B — Aggressive (cross to v0.54 + IBCv2)

Jump the SDK minor, take wasmd v0.70, take ibc-go v11, take cometbft v0.39.

| Component | From | To |
|---|---|---|
| cosmos-sdk | v0.53.4 | **v0.54.3** |
| wasmd | v0.54.2 | **v0.70.0** |
| wasmvm | v2.2.4 | **v3.0.4** |
| ibc-go | v8.7.0 | **v11.0.0** |
| cometbft | v0.38.19 | **v0.39.3** |
| store | v1.1.2 | **store/v2.0.0** |

**Why this is probably the right call:**

- v0.54 is the "2026.1 release family" — the line upstream will support and patch through 2027. Path A puts us on a maintenance branch.
- ibc-go v11 brings IBCv2 (Eureka), ICS27-GMP, attestation light client. IBCv2 is the primitive Juno needs for cross-chain agent-mandate to be real.
- wasmd v0.70 + wasmvm v3.0.4 is the matched pairing. Doing the wasmd jump halfway (v0.61.x) is the same review effort with less payoff.
- Aligns with `memory/juno-aggressive-planning.md`: the safety premium has collapsed; the contrarian thesis only pays if actually played.

**What's at risk:**

- store/v2 migration is non-trivial. Most downstream tooling (block explorers, indexers) needs to keep up.
- ibc-go v11 changes ICA middleware stacking — the wiring in `app/keepers/keepers.go:415–440` rewrites.
- cometbft v0.39 introduces optional libp2p networking. We don't enable it for v30, but the dependency adds surface.
- DAO DAO v2.7.0 contracts on chain were compiled against cosmwasm-std 2.x. They run on a wasmvm v3 chain per upstream's compat statement, but it's worth a real test pass before anyone relies on it.

## Path A+ — Pivot (session 2, 2026-05-08)

Discovered after Path B was locked: **the ibc-apps repo has no `/v11` line published** (latest is `/v10`, pinned to `ibc-go v10` + `cosmos-sdk v0.53`). Juno wires PFM, ibc-hooks, and async-icq into `app/keepers/keepers.go` + `app/modules.go` + `app/keepers/keys.go`; pure Path B would either lose all three or require forking ibc-apps to /v11 ourselves.

Pivoted to **Path A+**: take Path A's SDK + IBC line, but still pull in wasmvm v3 (the consensus-break that justifies a v30 upgrade height regardless).

| Component | From | To |
|---|---|---|
| cosmos-sdk | v0.53.4 | **v0.53.7** |
| wasmd | v0.54.2 | **v0.61.11** |
| wasmvm | v2.2.4 | **v3.0.4** (path change `/v2` → `/v3`) |
| ibc-go | v8.7.0 | **v10.6.0** (path change `/v8` → `/v10`) |
| ibc-apps PFM | /v8 v8.2.0 | **/v10 v10.6.0** |
| ibc-apps ibc-hooks | /v8 v8.0.0 | **/v10 v10.0.0** |
| ibc-apps async-icq | /v8 v8.0.1-pseudo | **/v8 (latest commit)** — no /v9/v10/v11 published yet |
| cometbft | v0.38.19 | **v0.38.23** |
| store | v1.1.2 | v1.1.2 (stay on v1; defer store/v2 to v31) |
| `cosmossdk.io/*` family | (current) | **unchanged** — v0.53.7 SDK keeps the same pins |

**What we still get:**

- wasmvm v3.0.4 — the consensus-break that ships BN254 precompile (per prop #374). This is the user-facing v30 payoff.
- cosmos-sdk v0.53.7 patches (security + minor improvements).
- wasmd v0.61.11 — matched against wasmvm v3 + sdk v0.53 + ibc-go v10.
- ibc-go v10 + PFM/ibc-hooks/v10 — current widely-deployed IBC stack.

**What we defer to v31:**

- cosmos-sdk v0.54 family (the "2026.1 release line").
- IBCv2 / Eureka via ibc-go v11.
- store/v2 migration.
- cometbft v0.39 with optional libp2p networking.

V31 is sequenced for "once ibc-apps publishes a /v11 line." Tracking task: monitor `cosmos/ibc-apps` for the v11 cut.

## Recommendation

**Path A+** as above, with a hard pre-commit to test cw-hooks + DAO DAO + JunoClaw contracts on a v0.53.7 + wasmvm v3 devnet before any mainnet halt-height proposal.

Path A+ doesn't ditch the rebirth thesis — it sequences it. We get the actually-shippable consensus-break (wasmvm v3 + BN254) now, and earn the right to push v31 the moment ibc-apps catches up. Forking ibc-apps to /v11 ourselves was rejected because (a) it adds a forked dep we'd have to re-converge later, (b) it expands security review surface, and (c) BN254 is the user-visible win — IBCv2 is infrastructure that nobody's blocked on yet.

## Whichever path: locked decisions

- **Go toolchain: 1.25.x.** Already on `.mise.toml` and `go.mod`. CI workflows must catch up.
- **Module path: `github.com/CosmosContracts/juno/v30`.** This is the v30 release; the next bump is v31 with full module path rewrite.
- **wasmvm v3.0.4 is in.** Both paths take it. This is the consensus-breaking change that justifies a `v30` upgrade height even on Path A.
- **BN254 precompile lands via wasmvm v3.0.x upstream.** Per `memory/bn254-precompile-tracking.md` and prop #374 (passed May 5, 2026), this means we don't need to maintain a CosmWasm fork to get BN254 — we get it for free with the wasmvm bump.

## Path B cost in person-days (rough)

- Dependency bumps + compile: ~1 day
- store/v1 → store/v2 keeper migrations: ~2 days
- ibc-go v8 → v11 middleware re-stacking: ~2 days
- ante chain (`DeductFeeDecorator`, fee pipeline): ~1 day
- Custom modules audit (mint/burn/cw-hooks, the rest): ~2–3 days
- Upgrade handler additions (store mounts, param migrations): ~1 day
- Staking-snapshot binding implementation: ~2 days
- Test pass + ictest fixes: ~3 days
- CI reconciliation + green: ~1 day
- Security review: ~3–5 days (parallel-track-able)

**~16–18 days of focused work**, sequenced. Not all blocking — staking-snapshot and security review run in parallel with later phases.
