# CometBFT v0.38.19 → v0.38.23 — changelog scan for v30 upgrade

Scope: every CometBFT change Juno mainnet validators are about to consume
when they switch from the v29 binary's pinned cometbft v0.38.19 to the
v30 binary's pinned cometbft v0.38.23. Sourced from `cometbft/cometbft`
repo, branch `v0.38.x`.

## Findings

**No consensus-rule changes.** No `[CONSENSUS]`-tagged entries between
v0.38.19 and v0.38.23. Validators can upgrade independently; no halt
coordination is required from the CometBFT side. (The halt at upgrade
height is forced by `app/upgrades/v30` regardless.)

**No network-protocol breaks.** No `[NETWORKING]` or `[API-BREAKING]`
entries that affect node-to-node behavior on the v0.38 line.

**One opt-in feature** — experimental libp2p transport — was added on
the line and is **off by default**. Juno does not enable it.

## Per-version detail

### v0.38.23

Three bug fixes, all in the "shouldn't have happened in the first
place" category:

- **`[types]` Fix nil vote handling** ([#5777](https://github.com/cometbft/cometbft/pull/5777)).
  Defensive nil-check in vote processing. Pure crash-avoidance.
- **`[light]` stop witness comparison after divergence checks**
  ([#5820](https://github.com/cometbft/cometbft/pull/5820)). Light client
  optimization; not relevant to validators running full nodes.
- **`[abci]` prevent panic on unlock in socket server panic recovery**
  ([#5593](https://github.com/cometbft/cometbft/pull/5593)). Defensive
  panic-recovery in ABCI socket transport. Most validators use the
  in-process ABCI path, not socket — minor relevance.

### v0.38.22 (April 10, 2026)

- **`[evidence]` Light Client Attack evidence validation**
  ([#5638](https://github.com/cometbft/cometbft/pull/5638)).
  Tightens validation of `ByzantineValidators` field in evidence. No
  protocol break; validators stricter about what evidence they accept.
- **`[blocksync]` ExtendedCommit verification via next block's LastCommit**
  ([#5629](https://github.com/cometbft/cometbft/pull/5629)).
  Tightens block-sync correctness. Affects nodes that fall behind and
  catch up via blocksync, not steady-state validators.
- **`[blocksync]` use full commit verification instead of light**
  ([#5663](https://github.com/cometbft/cometbft/pull/5663)).
  Same scope as above. More-stringent verification during catch-up.
- **`[abci]` socket server panic recovery** — duplicate entry, same as
  v0.38.23.

(Note: the published v0.38.22 CHANGELOG has a chunk of v1.x / `lp2p`
improvements pasted into the IMPROVEMENTS section that appear to be
backport bleed-through from a master changelog edit. The real v0.38.22
release is bug-fix-only; the libp2p material is opt-in transport work
that didn't ship enabled in this patch line. Confirmed by reading the
release notes on the GitHub release page rather than the CHANGELOG.md
file.)

### v0.38.21 (January 23, 2026)

- **`[statesync]` configurable `max-snapshot-chunks`**
  ([#5549](https://github.com/cometbft/cometbft/pull/5549)). New config
  parameter caps the number of chunks in a `SnapshotResponse`. Default
  is unlimited; opt-in. State-sync–only; no validator action required.

### v0.38.20 (December 12, 2025)

Maintenance release; no functional changes flagged in the CHANGELOG.

## What this means for the v30 upgrade proposal

Validators upgrading from the v29 binary to the v30 binary will move
from cometbft v0.38.19 to v0.38.23. They get four patch releases
worth of bug fixes (vote handling, evidence validation, blocksync
verification, ABCI socket panic recovery). They get one opt-in config
parameter (state-sync chunk cap). They get an experimental libp2p
transport that **stays disabled** unless they explicitly turn it on.

For the upgrade proposal text, the validator-facing language is:

> The v30 binary embeds CometBFT v0.38.23, four patch releases ahead of
> v29's v0.38.19. The CometBFT line is bug-fix-only between these
> versions; no consensus rules change and no validator coordination is
> required beyond the standard halt-height swap. The optional libp2p
> transport remains disabled.

No flagged blockers from CometBFT for v30.
