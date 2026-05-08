# 01 — Current state of `jakehartnell/v30`

## What's already there

Branch: `jakehartnell/v30`. 82 commits ahead of `origin/main` (last main commit June 18, 2025; last v30 commit Nov 10, 2025). The branch is not stale work — it's a substantial chain upgrade in flight that paused.

### Dependency state (`go.mod`)

```
go 1.25.2
cosmos-sdk            v0.53.4
cosmossdk.io/api      v0.9.2
cosmossdk.io/store    v1.1.2
cosmossdk.io/x/upgrade v0.2.0
cosmossdk.io/x/feegrant v0.2.0
cosmossdk.io/x/evidence v0.2.0
wasmd                 v0.54.2
wasmvm/v2             v2.2.4
ibc-go/v8             v8.7.0
cometbft              v0.38.19
ibc-apps/pfm/v8       v8.2.0
ibc-apps/async-icq/v8 v8.0.1
ibc-apps/ibc-hooks/v8 v8.0.0
```

### Module changes already landed

From `app/upgrades/v30/constants.go`:

- **Deleted** stores: `globalfee`, `crisis`, `params`, `nft` (legacy modules removed in SDK 0.53 line)
- **Added** store: `feemarket` (Skip's AIMD EIP-1559 base-fee module, fully wired in)

From the commit history:

- `feemarket` ported from Skip and integrated into the ante chain
- `cw-hooks` rewritten ("cw hooks v2" — `7dbbf115`)
- `x/stream` reworked twice; current state is "simple block listener" (`07f90331`)
- `tokenfactory` migrated; tests fixed up (`6b256cde`)
- Repo-wide lint pass against stricter config (`dfced760`)
- Modernised `Dockerfile` and compose setup (`55d7cbda`)
- Interchaintest framework upgraded to "official Cosmos version" (`f55c5c55`)
- Module path bumped to `github.com/CosmosContracts/juno/v30`

### Upgrade handler scope (`app/upgrades/v30/upgrades.go`)

Configures fee-market params (alpha, beta, gamma, delta, MinBaseGasPrice 0.075 ujuno, learning-rate bounds, denom from staking params, distribute fees on). Configures `cw-hooks` `ContractFailureRemovalThreshold = 3`. Runs `mm.RunMigrations(...)`.

### Custom modules present

`x/clock`, `x/cw-hooks`, `x/feemarket`, `x/feepay`, `x/feeshare`, `x/tokenfactory`, `x/mint`, `x/burn`, `x/drip`, `x/stream`, `x/wrappers/gov`. All compile under the current `go.mod`.

## The gap to upstream as of 2026-05-08

| Component | v30 branch | Latest stable | Gap |
|---|---|---|---|
| go | 1.25.2 | 1.25.x | up to date |
| cosmos-sdk | v0.53.4 | v0.53.7 (patch) / v0.54.3 (next line) | 3 patches behind / 1 minor behind |
| cosmossdk.io/store | v1.1.2 | store/v2.0.0 | major. v0.54 line moves to store/v2 |
| wasmd | v0.54.2 | v0.61.11 (sdk-v0.53 line) / v0.70.0 (sdk-v0.54 line) | 7 minor / 16 minor behind |
| wasmvm | v2.2.4 | v3.0.4 | major. wasmvm v3 is consensus-breaking. cosmwasm-std v1/v2 contracts continue to run per upstream. |
| ibc-go | v8.7.0 | v8.8.0 (patch) / v10.6.0 (IBCv2) / v11.0.0 (IBCv2 + ICS27-GMP, sdk v0.54, cometbft v0.39) | 1 patch / 2 minor / 3 minor |
| cometbft | v0.38.19 | v0.38.23 (latest patch) / v0.39.3 (next line) | 4 patches / 1 minor |

## CI mismatch (latent breakage)

`.github/workflows/build.yml` and `interchaintest-E2E.yml` both pin `GO_VERSION: 1.23.9`, but `go.mod` requires `go 1.25.2` and `.mise.toml` declares `go = "1.25.2"`. CI on `jakehartnell/v30` won't build. This was not caught because the branch hasn't been pushed in CI-active mode for months.

The interchaintest matrix in the workflow lists 16 suites; the `Makefile` `.PHONY` list only declares 13 ictest targets. Several CI matrix entries (`ictest-statesync`, `ictest-ibchooks` (note hyphen mismatch with `ictest-ibc-hooks`), `ictest-feeshare`, `ictest-unity-deploy`, `ictest-feepay`, `ictest-cwhooks`, `ictest-clock`, `ictest-gov-fix`) reference targets that don't exist in the current `Makefile`. Either the Makefile lost targets during the v30 rework, or the CI workflow lost sync with it. Either way: the CI workflow needs reconciling before "green CI" is a meaningful gate.

## Submodule drift inside `interchaintest/`

`interchaintest/go.mod` pins `wasmd v0.60.0` while the parent module pins `v0.54.2`. The interchaintest module is technically separate, but the version drift signals incomplete work — someone started bumping wasmd in the test harness and didn't finish in the parent.

## TODOs that pre-date this work

- `app/keepers/keepers.go:661` — *"we need to wait until ALL cosmos sdk and juno modules FULLY implement SDK collections"* — collections migration is queued but not in scope here.
- `x/stream/keeper/keeper.go` — same TODO.
- `app/ante/ante.go:28` — *"readd maxBypassMinFeeMsgGasUsage, gone because of Globalfee removal"* — minor regression from the globalfee deletion. Worth fixing as part of the ante audit (`04-modules.md`).

## Practical implication

The branch is roughly two-thirds of the way to a v30 release. The remaining work is:

1. Finish the upstream version bumps (decision point in `02-targets.md`)
2. Reconcile CI with the actual Makefile and Go version
3. Add the staking-snapshot binding (`05`) — the only feature add, justified because the chain-upgrade window is the right place for it
4. Run the test suite, fix breakage, re-run
5. Security review

This is not a from-scratch upgrade. It's finishing a started one.
