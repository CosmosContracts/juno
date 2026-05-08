# 08 — Security review

The user's directive: a security review at the end of the process. This file is the scope, the checklist, and the structure of that review. It is read before review begins and signed off when review closes.

## Why this exists

A chain upgrade is a window where consensus rules change. State migrations, ante chains, and module additions are the surfaces where a bug becomes a chain halt or a value-extraction opportunity. The security review's job is to catch what unit tests and ictest can't:

- Subtle ordering bugs in BeginBlocker / EndBlocker / Migration sequences
- Wasm-binding API surfaces that contracts can call to escape their gas envelope
- Stargate query paths that leak more than they should
- Custom ante decorators that invert fee math under a specific tx shape
- Authorization checks that the SDK previously enforced and our forks now don't

## Scope

**In scope:**

- The migration diff (every change between `origin/main` and the tip of the migration branch)
- `app/upgrades/v30/upgrades.go` — the migration body, in particular state writes
- `app/ante/` and `app/post.go` — every decorator in the chain
- `app/keepers/keepers.go` — keeper construction, capability granting, IBC middleware stacking
- `app/keepers/acceptedQueries.go` — every stargate path
- `wasmbindings/` — every custom query and message
- The `x/voting-snapshot` module (new code; never been reviewed)
- `x/mint` and `x/burn` — re-review as a unit because they're paired
- `x/cw-hooks` — re-review the staking lifecycle dispatch (a panic in a hook is a chain halt)
- `x/feemarket`, `x/feepay`, `x/feeshare` as a fee-pipeline group

**Out of scope (or in scope only as cross-checks):**

- Upstream cosmos-sdk, wasmd, wasmvm, ibc-go, cometbft. We trust the upstream review. We verify *our integration* of them.
- Vendored modules we didn't change (e.g., `x/clock` if its diff is empty or trivial)
- DAO DAO contracts — separate review surface, separate team
- JunoClaw contracts — separate review surface, separate team

## Reviewer pool

External reviewers are non-negotiable for the upgrade handler and the new module. Internal eyes are not enough on either.

Suggested reviewers (per `memory/people.md` and prior collaboration):

- **Reece Williams (Reecepbcups)** — long-running Juno SDK contributor; knows the ante chain and the custom modules
- **Faddat / Jonathan Gimeno** — has reviewed prior Juno upgrade handlers
- **Faraz / Cosmos validator pool** — for the validator-facing upgrade procedure
- **Patrick (Commonware)** — out of scope, separate hyperstition track per `memory/jake-relational-norms.md` (don't pitch at idea stage)
- **Audit firm (Halborn / Oak / SCV / Informal)** — only if budget warrants; default no, given the diff is incremental and the upstream code is itself audited

The minimum viable review: **two external reviewers** sign off on the upgrade handler and the new module. Other surfaces can be self-reviewed plus one external.

## Checklist

### Upgrade handler (`app/upgrades/v30/upgrades.go`)

- [ ] Every `RunMigrations(...)` step is observed to complete; no silent skips
- [ ] State writes are bounded — no unbounded loops over delegations / accounts / contracts that could OOM the upgrade
- [ ] Param resets (feemarket, cw-hooks) use sane defaults under all chain conditions; verify against current mainnet staking params and consensus params
- [ ] Module-account permissions in `app/modules.go` match what the migrations expect
- [ ] Store deletions (`globalfee`, `crisis`, `params`, `nft`) are total — no orphan keys left behind
- [ ] Store additions (`feemarket`, `voting-snapshot`) are mounted before any code reads from them
- [ ] Idempotency: the handler can be run only once at the upgrade height; no re-entry guards needed but cross-check
- [ ] Backfill of voting-snapshot index reads from staking state at upgrade height — cross-check that the iteration order is stable

### Ante chain (`app/ante/ante.go`, `app/ante/decorators/`)

- [ ] Decorator order is preserved from the prior chain except where intentionally changed
- [ ] Custom `DeductFeeDecorator` correctly handles: bank-send fee, contract-execute fee, contract-execute under feepay, multi-msg tx with mixed feepay status
- [ ] `MsgFilterDecorator` does not silently drop messages that should error — denied messages produce an explicit error, not a silent skip
- [ ] `ChangeRateDecorator` (validator commission rate change) still enforces the same caps as pre-upgrade
- [ ] `RedundantRelayDecorator` (ibc-go) integrates with v11's middleware stack
- [ ] Fee deduction never produces negative balances; verify edge case where feepay balance is exactly the fee amount
- [ ] Fee market base-fee floor is enforced — a tx below floor errors, not paid-from-feepay

### Wasmbindings

- [ ] Every custom query plugin enforces gas charging; none are free
- [ ] Every custom message plugin checks authorization (admin-set checks for tokenfactory, etc.) before mutating state
- [ ] No reentrancy: a custom message that triggers a wasm execute does not recursively re-enter the same custom plugin in a way that bypasses gas
- [ ] Voting-snapshot queries are read-only; verify they can't be invoked through a write path

### Voting-snapshot module (new)

- [ ] LST exclusion list is queryable but only mutable via gov
- [ ] Snapshot retention policy is bounded — old heights are pruned per a configurable retention window
- [ ] Storage cost on a delegation event is bounded; no per-validator iteration on every delegate/undelegate
- [ ] Slashing events update snapshots correctly
- [ ] Querying a future height returns a typed error, not the latest snapshot
- [ ] Querying a height pre-upgrade returns a typed error, not zero (a contract treating zero as "no power" might silently disenfranchise pre-upgrade voters)

### Stargate allow-list

- [ ] Every path is required by an existing on-chain contract or a documented future use
- [ ] No path leaks data that wasn't intended (e.g., slashing details, unbonding entries) without precedent
- [ ] No path was added that ought to require its own gas accounting

### Custom modules — focused re-review

- [ ] `x/mint` — inflation calc cannot underflow / overflow with v0.54 staking params
- [ ] `x/burn` — burning into the fixed module address still works; no dangling tokens
- [ ] `x/cw-hooks` — a contract that panics in a staking hook does not halt the chain (verify ContractFailureRemovalThreshold actually triggers; verify wrapped panic-recover)
- [ ] `x/feepay` — verify a contract cannot register itself for feepay without depositing balance first

### Cross-cutting

- [ ] All event emissions remain consistent — events are part of the indexed contract; an event-name change is a soft API break
- [ ] gRPC reflection still works; tooling depends on it
- [ ] CLI commands (junod tx, junod query) for every custom module return correctly-typed responses
- [ ] OpenAPI / Scalar docs (`app/endpoints/`) regenerate correctly

## Severity definitions

- **Sev-1** — chain halt, state corruption, value loss, unauthorized fund movement. Fix and re-review before any release tag.
- **Sev-2** — incorrect-but-recoverable behavior, gas-cost discrepancies, event-emission regressions. Must be fixed or explicitly deferred with Jake's sign-off and a tracked follow-up issue.
- **Sev-3** — code quality, comment accuracy, naming. Fix if cheap, defer if not.
- **Info** — observations, future-work notes. No blocking effect.

## Sign-off format

When review closes, this file is updated with:

```
## Sign-off

- Reviewer A (handle / role): signed YYYY-MM-DD, findings: [Sev-1: 0, Sev-2: 1 (deferred), Sev-3: 4 (fixed)]
- Reviewer B: ...
- Internal (Juno AI): findings recorded inline above

Sev-2 deferrals:
- <issue link>: <why deferred, when scheduled>

Sev-1 / open issues at sign-off: 0
```

No release tag without zero open Sev-1.

## Timing

Review starts **after** Phase 6 (interchaintest pass green) and runs in parallel with Phase 7 (CI reconciliation). Aim for review-complete by Day 18 of the migration. If review surfaces work that pushes Day 18, that's the right answer — don't compress security review to hit a date.
