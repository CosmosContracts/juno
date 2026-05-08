# 07 — Rollout

After the migration branch is green and the security review is signed, the binary is candidate-ready. This file is the path from candidate to mainnet.

## Stage 1 — Release candidate (Day 0 from rollout)

- Tag `v30.0.0-rc1` against the migration branch
- Build deterministic Docker image; push to `ghcr.io/cosmoscontracts/juno:v30.0.0-rc1`
- Draft release notes — see template below
- Publish a private message in the Juno dev pool: "RC1 cut, asking for testnet validators"

## Stage 2 — Testnet deploy (Days 1–10)

Per `memory/key-persistence.md` and `memory/juno-rpc.md`, the current testnet is uni-7. Two paths:

**Path A — upgrade uni-7 in place.** Submit a chain-upgrade proposal on uni-7, gather quorum from the testnet validator set, execute the upgrade, observe.

**Path B — spin up a new testnet from genesis.** Useful if uni-7 has accumulated state we don't want, or if the upgrade-from-v29 path itself is what's risky and a clean genesis is faster to validate.

**Lean: Path A.** The point of a testnet is to test the upgrade procedure under as close to mainnet conditions as we can get. Genesis-fresh is a different test.

What "testnet success" means:

- Halt height reached, validators upgrade, chain produces blocks again
- Existing testnet contracts (JunoClaw uni-7 deployment, DAO DAO testnet, anything else live) keep running
- A newly-deployed contract can use the new wasmvm v3 features (BN254 if exercised) without issue
- Voting-snapshot module returns sane data for at least one staked-JUNO DAO DAO test
- IBC-go v11 paths still relay against a counterparty (Juno↔ Cosmos Hub testnet, or a sister Cosmos chain on v0.54)
- Indexer / explorer tooling (if any track uni-7) keeps up

Run for **at least 10 days** before mainnet. Shorter windows have caught us before.

## Stage 3 — Governance proposal draft (Day 8, parallel with testnet observation)

Draft the on-chain `MsgSoftwareUpgrade`:

- **Plan name:** `v30`
- **Plan height:** decided in coordination with `#validators-private` on Discord; aim for ~14 days from proposal start to give the voting period plus a safety margin
- **Plan info:** JSON describing release artifact URLs, expected sha256, supported architectures (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
- **Description:** prose summary linking to release notes

Posted to the public forum first (commonwealth.im or whatever Juno's current forum lives at), then on chain after a discussion window.

## Stage 4 — Public communications (Day 8–10)

- Forum post by Jake (or by Juno AI under mandate, see `memory/juno-twitter-account.md`) describing the upgrade and what's in it
- @JunoNetwork official channel announcement
- @juno_ai (Juno AI's account) — if it's posting by then, it covers the upgrade as a milestone in oracle-mode; see `STYLE.md`
- Validators-private Discord: pre-brief
- Cosmos validator Discord channels (`#validators-mainnet` on the cosmos engineering server): cross-post

Communications must include:
- Halt height
- Upgrade artifact (download URL, sha256)
- Required Go version for source builds (1.25.x)
- Time estimate of node downtime during upgrade
- Rollback procedure if the upgrade fails

## Stage 5 — Mainnet halt-height execution (Day 14)

Per `RELEASES.md`, the procedure:

1. Validators told via `#validators-private` of incoming halt
2. Halt-height proposal passes
3. Chain halts at the configured height
4. Validators stop their nodes, swap binaries, restart
5. Chain resumes; first post-upgrade block runs the v30 upgrade handler
6. Spreadsheet check on `#validators-private` to verify >67% have upgraded

Chain is expected to lose 5–30 minutes of block production at the halt height as validators coordinate. This is normal.

## Stage 6 — Post-upgrade observation (Day 14–21)

For one week after upgrade:

- Watch validator participation rate; flag if drops below 95% of pre-upgrade
- Monitor mempool for any pattern of failing txs that didn't fail pre-upgrade (signal: a custom message decoder broke)
- Watch `#validators-private` for any bug reports
- Watch indexer feeds — if `indexer.daodao.zone` (the deployed `indexer-proxy`) starts returning errors, that's a likely first symptom of a proto-path break that didn't surface in test
- Watch the Reece Williams or Faddat dashboards if they exist

If we see a Sev-1 issue in the first 24 hours, the right move is **not** to roll back — it's to patch and ship `v30.0.1` as an emergency patch per `RELEASES.md` emergency-upgrade procedure. Cosmos chains do not have a rollback once consensus is past the upgrade height.

## Stage 7 — Tag mainnet release (Day 21+)

- Tag `v30.0.0` (without `-rc`) once observation period closes clean
- Update `mainnet/juno-1/` in the meta-repo with the upgrade height entry (per the `0xxx`/`1xxx`/`2xxx` numbering convention from CLAUDE.md)
- Update `RELEASES.md` if procedure learnings emerged
- Close the security-review thread

## Release-notes template

```
# Juno v30 — 2026.1 release family

## What's in it

- cosmos-sdk v0.54.3 ("2026.1 release family")
- wasmd v0.70.0 + wasmvm v3.0.4 (consensus-breaking; brings BN254 precompile, IBCv2 async ack handling)
- ibc-go v11.0.0 (IBCv2 / Eureka, ICS27-GMP, attestation light client)
- cometbft v0.39.3
- New x/feemarket module (Skip's AIMD EIP-1559)
- New x/voting-snapshot module (historical staking power for DAO DAO consumption)
- Removed legacy modules: globalfee, crisis, params, nft
- Module path bumped to github.com/CosmosContracts/juno/v30

## Breaking changes for validators

- Go 1.25.x required to build from source
- Binary location unchanged; standard cosmovisor swap
- Halt height: <to be filled in>

## Breaking changes for contract devs

- wasmvm v3 is consensus-breaking, but cosmwasm-std v1 and v2 contracts continue to run unchanged. No contract redeploy required.
- New custom queries available to contracts: VotingPowerAt, TotalVotingPowerAt
- Stargate query allow-list unchanged

## Build instructions

go install github.com/CosmosContracts/juno/v30/cmd/junod@v30.0.0

## Verification

sha256: <filled in at tag time>
git commit: <filled in at tag time>
deterministic build: <docker tag>
```

## Coordination touchpoints

- **Validators**: `#validators-private` on Juno Discord
- **Indexer team**: whoever runs `indexer-proxy` and `indexer.daodao.zone`
- **DAO DAO team**: notify of voting-snapshot bindings; coordinate the `dao-voting-juno-staked` module update timing
- **JunoClaw team** (per `memory/junoclaw-coordination.md`): notify of wasmvm v3 ahead of time so any TEE/ZK code paths that interact with the VM are validated
- **Cosmos validator Discord**: cross-post for awareness; some validators run multiple chains

## Failure modes

If testnet shows a Sev-1 (chain doesn't restart, state is corrupt, contracts don't execute), pull the rc tag, fix, recut as `rc2`, re-test. There is no scheduling pressure that justifies shipping a broken upgrade to mainnet.

If mainnet upgrade fails partway (some validators upgrade, some don't, chain halts at <67%), the Juno dev pool runs the emergency-upgrade procedure: patch in private, brief in `#validators-private`, fast-track binary, restart. This has been done before and the procedure is in `RELEASES.md`.
