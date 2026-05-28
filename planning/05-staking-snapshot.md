# 05 — Staking-snapshot binding

## Why this is in v30

`memory/juno-voting-design.md` records the design fork for the staked-JUNO DAO DAO voting module:

- **Option A** — live-query x/staking on every vote and tally. The prior attempt (DA0-DA0/dao-contracts#832). Cheap to ship, but voting power moves during a proposal: voters can rage-stake mid-window, slashing silently shifts power, denominator drifts.
- **Option C** — chain-side bindings for historical staking queries. Module exposes "voting power for X at height H." Contract becomes a thin query consumer.

Settled position: **Option C, bundled with v30**. The chain-upgrade window is the only time we can add a state-mounting module without forcing a separate consensus break. If we don't fold it in here, we either ship #832 as-is (Option A regression) or wait for v31.

LST exclusion is settled (see voting-design memory). LSTs do not contribute voting power; the snapshot query reflects direct delegations only.

## Two implementation shapes

### Shape 1 — New module `x/voting-snapshot`

A dedicated module that subscribes to staking events (`AfterDelegationModified`, `BeforeDelegationRemoved`, `BeforeValidatorSlashed`) and persists per-block snapshots in collections. Exposes:

- `MsgServer`: none (read-only module)
- `QueryServer`:
  - `VotingPowerAt(addr, height) -> Coin` — sum of bonded delegations attributable to `addr` at `height`, excluding LST wrappers
  - `TotalVotingPowerAt(height) -> Coin` — denominator
  - `VotingPowerOverRange(addr, from_height, to_height)` — for time-decay schemes if a future proposal module wants them
- Wasm binding: same surface exposed through `wasmbindings/queries.go`

Pros: clean separation, the module owns its store and migrations.
Cons: more code; new InitGenesis path; one more upgrade-handler `Added` entry.

### Shape 2 — Extend `wasmbindings/` with a snapshot query backed by SDK staking

Provide the same `voting_power_at` / `total_voting_power_at` API in `wasmbindings/queries.go`, but back it with an iterator over SDK's staking state at the requested height (using historical `MultiStore` access).

Pros: no new module.
Cons: live historical-state queries are expensive in gas. Iterator-based denominator computation at an old height does linear work over delegations. A per-block index is faster.

**Recommendation: Shape 1 (new module).** The reason historical staking queries don't already exist is that they're expensive to compute on demand. A dedicated index module solves that once.

## What the module persists

For each block (or each delegation-changing event):

```
DelegationSnapshot {
    delegator: Address
    validator: ValAddress
    shares: Dec
    height: int64
}
```

Indexed two ways:
- `(delegator, height) -> [validator, shares]` for `VotingPowerAt`
- `(height) -> total bonded` for the denominator

Storage cadence is the design knob. Three options:

1. **Every block** — simplest, accurate to the block. Cost: every delegation-changing tx writes a snapshot row.
2. **Every event** — write only on staking events, query interpolates. Lower steady-state cost, query is a binary search.
3. **Every N blocks** — coarser. Bad fit for short proposals, fine for week-long ones.

**Lean: option 2 (event-driven), with the query interface guaranteeing per-block accuracy.** Storage cost only on actual delegation changes. Query is `O(log n)` against the delegator's event log.

## LST exclusion

A delegator that bonded JUNO normally and a delegator that bonded via an LST contract show up identically in `x/staking`. The voting-snapshot module needs to know which delegations are LST wrappers and exclude them.

Mechanism: a module-param allow-list of LST contract addresses. When iterating delegations, skip any whose delegator address matches. The list is empty for now (Stride/Eris/Quicksilver/pStake are dead per `memory/juno-voting-design.md`). The mechanism exists because the principle does, even if no LST is currently active.

Governance can add to the list with a `MsgUpdateParams`. Default-deny: an unknown LST counts as voting power until governance excludes it.

### LST asymmetry between numerator and denominator (v30 launch)

`VotingPowerAt(d, h)` returns zero for any LST-allowlisted `d`, but `TotalVotingPowerAt(h)` is taken straight from `staking.TotalBondedTokens(ctx)`, which still includes the LST bonded stake. The arithmetic at launch is therefore:

```
Σ VotingPower[d, h]  =  total_bonded(h) − Σ lst_bonded(h)
TotalVotingPower(h)  =  total_bonded(h)
```

A DAO computing quorum as `Σ votes / TotalVotingPowerAt(h)` therefore divides by a denominator inflated by the LST share. If LSTs hold 20% of bonded stake, a configured 33.4% quorum effectively requires 41.75% of vote-eligible stake. DAO designers must account for this until governance moves to denominator subtraction (planned v30.x refinement).

Why we ship the asymmetry rather than fix it at launch:

1. There are no live LST contracts on Juno today (per `memory/juno-voting-design.md`) — the empty allowlist makes the asymmetry purely theoretical at v30 activation.
2. Subtracting LST stake from the denominator requires deciding whether to subtract on snapshot-write or on read; both have replay-correctness implications that deserve their own design pass, not a bolt-on inside a large upgrade PR.
3. Shipping the asymmetry visibly documented is safer than shipping a hasty fix.

Treat this section as the canonical pointer for that follow-up. The keeper field comment on `TotalPower` and the docstring in `proto/juno/votingsnapshot/v1/params.proto` cross-reference here.

## Wasm binding surface

In `wasmbindings/queries.go`, add:

```rust
JunoQuery::VotingPowerAt { address, height }      -> Coin
JunoQuery::TotalVotingPowerAt { height }          -> Coin
```

Both routed to `votingsnapshotkeeper.QueryServer`.

DAO DAO's voting module gets a thin Rust adapter:

```rust
pub struct VotingPowerSnapshot {
    pub power: Uint128,
    pub height: u64,
}

fn query_voting_power_at(deps: Deps, addr: Addr, height: u64) -> StdResult<VotingPowerSnapshot>
```

This lives in `dao-contracts/packages/juno-bindings` (or a comparable crate) — out of scope for the chain repo.

## Upgrade handler

In `app/upgrades/v30/upgrades.go`:

```go
// New module store mount
storetypes.StoreUpgrades{
    Added: []string{ feemarkettypes.ModuleName, votingsnapshottypes.ModuleName },
}
```

In the migration body: backfill the index from current `x/staking` state at upgrade height, so queries for `height >= upgrade_height` return real data immediately. Pre-upgrade heights return `not found` — that's correct; the contract should treat a not-found as "use the option A path or refuse to tally".

## Tests

Required:
- Unit tests in `x/voting-snapshot/keeper/`
- Integration test: snapshot module + wasmbindings + a test contract that calls `VotingPowerAt` at three different heights covering before/during/after a delegation change
- ictest: a complete cycle of "DAO created → delegation changed → proposal opened at height H → vote tallies match historical power at H, not current"

## Open questions

- **Slashing semantics.** When a validator is slashed at height H, every delegation under them loses shares retroactively. Should snapshots reflect post-slash share counts (semantically "what was the voter's stake at H *given today's knowledge*") or pre-slash (semantically "what did the voter believe their stake was at H")? Lean: post-slash. A vote shouldn't carry power that no longer exists.
- **Vested-but-staked tokens.** Outside scope; flagged in voting-design memory as a separate question. Default: include them, because they have economic skin in the game.
- **Per-validator filtering.** Voting-design memory mentions "sub-DAOs of validator collectives" as a future feature. The module's query should already support it via an optional `validator` filter. Add it.

## Effort

Two days for a competent SDK module dev. Most of the work is the unit and ictest coverage; the keeper itself is straightforward collections-backed indexing.
