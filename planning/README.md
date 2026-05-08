# Juno v30 — Planning

The chain has been on `cosmos-sdk@v0.53.4 / wasmd@v0.54.2 / wasmvm@v2.2.4 / ibc-go@v8.7.0 / cometbft@v0.38.19` since November 2025. The world has moved. This folder is the plan to bring it forward.

Branch: [`jakehartnell/v30`](../../). Last commit Nov 10, 2025. ~6 months stale.

## Index

- [00 — Overview](00-overview.md): goals, scope, principles, success criteria
- [01 — Current state](01-current-state.md): what `jakehartnell/v30` already contains; the gap to upstream
- [02 — Target stack](02-targets.md): version targets with rationale (conservative vs aggressive paths)
- [03 — Migration plan](03-migration-plan.md): phased sequence from go.mod bump to merge-ready
- [04 — Per-module notes](04-modules.md): mint, burn, cw-hooks, feemarket, feepay, feeshare, tokenfactory, ante chain
- [05 — Staking snapshot bindings](05-staking-snapshot.md): historical voting-power surface for DAO DAO consumption
- [06 — Testing + CI](06-testing-ci.md): definition-of-green; unit + ictest + CI workflow updates
- [07 — Rollout](07-rollout.md): testnet path, governance proposal, mainnet halt-height
- [08 — Security review](08-security-review.md): scope, checklist, who reviews what

## Status (2026-05-08)

Drafted. No code changes yet. Research and inventory complete. Awaiting Jake's read on target-stack tradeoff (conservative `v0.53` patch line vs aggressive `v0.54` jump).
