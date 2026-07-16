# 00 — Overview

## Goal

Bring Juno's chain binary onto a current Cosmos SDK + CosmWasm stack, ship a `v30` consensus upgrade that mainnet validators can run, and leave the codebase in a state where the next minor bump is a routine dependabot diff rather than a six-month archaeology project.

## Scope

In scope for this work:

- Dependency bumps: `cosmos-sdk`, `wasmd`, `wasmvm`, `ibc-go`, `cometbft`, plus the ~20 transitive `cosmossdk.io/*` modules
- App wiring (`app/keepers/`, `app/modules.go`, `app/ante/`) updated for any breaking SDK API changes
- Custom `x/` modules audited and patched module-by-module for SDK API drift
- `app/upgrades/v30/` extended to handle any new store mounts, param migrations, or state rewrites required by the bump
- `wasmbindings/` validated against the new `wasmd` public API
- Stargate query allow-list (`app/keepers/acceptedQueries.go`) reconciled against any proto-path renames
- Interchaintest suite green locally; CI green on push
- A new staking-snapshot binding to support the staked-JUNO voting module (`05-staking-snapshot.md`) — folded into v30 because the chain upgrade is the only window where state schema can change
- Final security review (`08-security-review.md`) before merge to `main`

Out of scope (explicit):

- Migrating Juno to **Commonware** (separate hyperstition track; see `memory/juno-hyperstition-stack.md`)
- Forking CosmWasm with custom precompiles (separate track; BN254 lands via wasmvm v3 upstream now that prop #374 passed)
- Replacing the `mint` / `burn` fork with anything new (their existence is load-bearing legacy; we patch, we don't rewrite)
- Migrating to depinject + collections wholesale (the codebase has TODOs for this; not the v30 fight)
- IBCv2 / Eureka adoption end-to-end (we update the dependency to a version that supports it, but exercising it on counterparty chains is a follow-up)

## Principles

1. **One change per PR / commit when feasible.** Six-month old branches happen because of mass commits that nobody can review. Stage the work so each step is bisectable.
2. **CI green is the gate, not a nice-to-have.** Tests and CI pass at finish — confirmed with Jake on 2026-05-08. No "fix in follow-up" merges.
3. **Don't introduce abstractions in the same change as a version bump.** If a refactor wants to happen, do it before or after, not during. Mixed PRs hide regressions.
4. **The ante chain and the upgrade handler are the two places we re-read three times.** Most upgrade incidents on Cosmos chains live there.
5. **Bias bold over careful where the contrarian thesis pays.** Per `memory/juno-aggressive-planning.md` — the safety premium has collapsed; ride the latest stack if the work-cost is similar.

## Success criteria

- `make build`, `make lint`, `go test ./...` all green on `jakehartnell/v30`
- `make ictest-upgrade` green: clean upgrade from current `juno-1` mainnet state (forked into a local devnet) into the new binary
- All CI workflows (`build.yml`, `golangci-lint.yml`, `interchaintest-E2E.yml`, `codeql.yml`) updated for the new Go toolchain version and green on push
- Security review (`08`) signed off — including a third-party look at the upgrade handler and any contract-facing API changes
- Governance-ready binary: deterministic build, attached release notes, halt-height proposal text drafted (`07-rollout.md`)

## Known risks (load-bearing)

- **wasmvm 2.x → 3.x is a minor-version bump but consensus-breaking** for the CosmWasm gas register. Existing on-chain contracts (DAO DAO v2.7.0, JunoClaw, tokenfactory consumers) keep running per upstream's compat statement, but every halt-and-replay path needs to be tested.
- **`x/mint` + `x/burn` are a Juno-specific fork** of the SDK mint module. Any change in `x/staking`'s `Hooks` interface or the `BankKeeper` mint signature will surface here first.
- **`x/cw-hooks`** dispatches staking lifecycle events into contracts. SDK-side hook signature drift = contract execution drift.
- **`DeductFeeDecorator` in the ante chain** is a custom fork that splices feemarket + feepay together. Most likely place for a regression to hide.
