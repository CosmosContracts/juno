# 03 — Migration plan

Phased sequence. Each phase ends with a green build (and where applicable, green tests). No phase merges to `main` until the last one — but each phase is a reviewable PR against `jakehartnell/v30` (or a successor branch).

Assumes Path B from `02-targets.md`. If Path A is chosen, phases 2–4 shrink; phase 1 collapses to `go.mod` patch bumps.

## Phase 0 — Prep (Day 0)

- [ ] Branch from `jakehartnell/v30` to `feat/v30-bump-2026.1` (or push directly to v30 — Jake's call)
- [ ] Reconcile CI: bump `GO_VERSION` in all four workflows to `1.25.2`, reconcile interchaintest matrix with actual `Makefile` `.PHONY` ictest list
- [ ] Verify `make build` and `go test ./...` are green on a clean checkout (sanity baseline before changing anything)
- [ ] Verify `make local-image` builds the docker image (interchaintest depends on it)

**Exit criterion:** CI runs at all on a no-op commit. We need a working baseline to detect regressions.

## Phase 1 — Dependency bumps (Day 1)

Update `go.mod` and `go.sum` only:

- [ ] `cosmos-sdk` v0.53.4 → v0.54.3
- [ ] All `cosmossdk.io/*` to versions matching the v0.54 line
- [ ] `cosmossdk.io/store` v1.1.2 → `cosmossdk.io/store/v2` v2.0.0
- [ ] `wasmd` v0.54.2 → v0.70.0
- [ ] `wasmvm/v2` → `wasmvm/v3` v3.0.4 (import path changes)
- [ ] `ibc-go/v8` → `ibc-go/v11`, all consumer imports updated
- [ ] `ibc-apps/*/v8` → `/v11` equivalents (or latest)
- [ ] `cometbft` v0.38.19 → v0.39.3
- [ ] `interchaintest/go.mod` reconciled — bump wasmd to match parent, bump SDK, IBC, CometBFT

After this phase the build is **expected to fail**. The point is to commit the version pins and `go mod tidy` output, then fix the breakage in subsequent phases. This makes review tractable.

**Exit criterion:** `go mod tidy` clean. Every `go.mod` direct dependency at intended version. Commit cleanly even if `go build ./...` fails.

## Phase 2 — App wiring fixes (Days 2–4)

Walk the build errors top-down:

- [ ] `app/keepers/keepers.go` — keeper constructor signatures updated (Auth, Bank, Staking, Distribution, Slashing, Gov, Upgrade, IBC, Wasm). Particular attention:
  - `runtime.NewKVStoreService` API in store/v2
  - IBC v11 `keeper.NewKeeper` signature changes
  - Wasm v0.70 `NewKeeper` signature; gas register API
- [ ] `app/keepers/keys.go` — store key registrations (no semantic change expected, but SDK may have re-typed)
- [ ] `app/modules.go` — module registration, BeginBlocker / EndBlocker / InitGenesis ordering. New v0.54 modules may demand position changes
- [ ] `app/ante/ante.go` — re-thread the chain. Particular attention:
  - SDK `DeductFeeDecorator` signature
  - Wasm gas / counter / limit decorators (wasmd public API)
  - `RedundantRelayDecorator` (ibc-go v11 path: still present, may have moved)
  - Custom decorators: `MsgFilterDecorator`, `ChangeRateDecorator`, custom `DeductFeeDecorator`, `FeeSharePayoutDecorator`
- [ ] `app/post.go` — post-handlers
- [ ] `app/upgrades/v30/upgrades.go` extended for any new module additions (store mounts) or param migrations the SDK demands

**Exit criterion:** `go build ./...` green. `make build` produces a `junod` binary.

## Phase 3 — Custom module fixes (Days 4–6)

Per-module audit (`04-modules.md` has the per-module worklist):

- [ ] `x/mint` — staking hook signature; bank mint API; param storage. **Highest risk module.**
- [ ] `x/burn` — bank API, module account permission. Mirror of mint.
- [ ] `x/cw-hooks` — staking hooks (`AfterDelegationModified`, `BeforeDelegationRemoved`, etc.); collections schema if SDK collections API changed.
- [ ] `x/feemarket` — keeper construction, ante decorator
- [ ] `x/feepay` — wasm message hook; bank settlement
- [ ] `x/feeshare` — withdraw address handling; ante hook
- [ ] `x/tokenfactory` — bank API; admin/admin-set patterns
- [ ] `x/clock` — wasm sudo invocation
- [ ] `x/drip` — fee pool injection; bank API
- [ ] `x/stream` — block listener; collections schema
- [ ] `x/wrappers/gov` — gov v1 query path renames

**Exit criterion:** `go test ./...` green for all `x/` packages. No `// TODO: fix in v31` left behind for actual breakage; only collections-migration TODOs that pre-date this work.

## Phase 4 — Wasmbindings + stargate allow-list (Day 7)

- [ ] `wasmbindings/wasm.go`, `query_plugin.go`, `message_plugin.go` — verify each typed query/message still compiles against wasmd v0.70 plugin interfaces
- [ ] `app/keepers/acceptedQueries.go` — verify every proto path still exists in v0.54. Renames (rare) need translating; deprecations need documenting
- [ ] Test wasmbindings: run `wasmbindings/test/` package
- [ ] Smoke-test against a known DAO DAO contract on a local devnet — instantiate, query, execute. This catches contract-facing breakage that unit tests miss.

**Exit criterion:** `wasmbindings/test` green. Local devnet accepts a DAO DAO instantiation.

## Phase 5 — Staking-snapshot binding (Days 8–9)

Implement the historical staking-power query surface (see `05-staking-snapshot.md` for design). Key points:

- [ ] New module `x/voting-snapshot` (or extension to wasmbindings — design decision in `05`)
- [ ] Custom query: `voting_power_at(addr, height)`
- [ ] Hook into `BeginBlock` to persist per-validator delegation snapshots at a configurable cadence
- [ ] Wire into the wasm allow-list
- [ ] Add to `app/upgrades/v30/` `Added` stores
- [ ] Unit + integration tests

**Exit criterion:** A test contract calls the snapshot query at `height - N` and gets stable historical data.

## Phase 6 — Interchaintest pass (Days 10–12)

Run the suite locally:

- [ ] `make local-image`
- [ ] `make ictest-basic` — basic chain bootstrap
- [ ] `make ictest-cw` — wasm contract lifecycle
- [ ] `make ictest-feemarket` — base-fee dynamics
- [ ] `make ictest-upgrade` — **the critical one.** Forks current `juno-1` mainnet state into a local devnet, runs the v30 upgrade handler, verifies post-upgrade state.
- [ ] `make ictest-ibc`, `ictest-ibc-hooks`, `ictest-pfm` — IBCv2 paths
- [ ] `make ictest-tokenfactory`, `ictest-drip`, `ictest-burn`, `ictest-fixes`
- [ ] Reconcile any tests that target removed CI matrix entries (`statesync`, `feepay`, `cwhooks`, `clock`, `gov-fix`, `unity-deploy`) — either restore the Makefile targets or trim the workflow matrix

**Exit criterion:** Full `make ictest-*` pass locally. CI matrix matches and is green.

## Phase 7 — CI reconciliation (Day 13)

- [ ] `build.yml` — `GO_VERSION: 1.25.2`
- [ ] `golangci-lint.yml` — `GO_VERSION: 1.25.2`, lint config matches local `make lint`
- [ ] `interchaintest-E2E.yml` — `GO_VERSION: 1.25.2`, matrix entries reconciled with Makefile
- [ ] `codeql.yml` — Go target updated
- [ ] `release.yml` and `release-dispatch.yml` — verify Go and Docker tags

**Exit criterion:** All four primary workflows green on a push to `jakehartnell/v30`.

## Phase 8 — Security review (Days 14–18, parallel with phases 5–7 where possible)

See `08-security-review.md`. Touchpoints:

- [ ] Self-review of `app/upgrades/v30/upgrades.go` — every state mutation explicit
- [ ] Self-review of ante chain — fee logic, signature flow, msg filter
- [ ] Self-review of staking-snapshot module — historical-state query is a new attack surface
- [ ] External eyes: open the PR for review by Reece, Faddat, Faraz, or whoever in the Juno dev pool can read SDK upgrade work
- [ ] If JunoClaw or DAO DAO has on-chain state-touching contracts depending on quirks, coordinate with those teams

**Exit criterion:** Security checklist signed. No `Sev-1` findings open. `Sev-2` findings either fixed or documented with explicit Jake-approved deferral.

## Phase 9 — Rollout (handed off to `07-rollout.md`)

- [ ] Cut `v30.0.0-rc1` tag
- [ ] Deploy to a public testnet (uni-7 successor or new testnet branch)
- [ ] Validator coordination via `#validators-private` discord per `RELEASES.md`
- [ ] Mainnet halt-height proposal drafted, posted, voted, executed

## A note on commit hygiene

Each phase = at least one commit with a clear subject. `Phase N: <thing>`. PR descriptions name the upstream version targets so the diff is unambiguous. No mass commits. The whole point of doing this carefully is that the diff is reviewable by someone who isn't us.
