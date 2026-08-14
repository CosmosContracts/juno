# v31 state rehearsal

This is the release-candidate gate for restarting representative Juno state on
v31. A unit test, a node merely reaching the tip, or an unrecorded successful
command is not sufficient.

## Mandate

Prove all three paths using immutable binaries and a mainnet-derived state
volume or representative extracted subset:

1. `v30.0.0` export followed by a clean v31 genesis import.
2. Snapshot-based state sync into a clean v31 node.
3. The exact in-place `v30.0.0 -> v31` upgrade path.

Never operate on the canonical validator home. Work on a read-only snapshot or
a cloned volume. Keep keys and private-validator state out of the rehearsal.

## Inputs

Record before execution:

- source tag, full commit, and `sha256:` image digest;
- target full commit and `sha256:` image digest;
- chain ID and sanitized source-state provenance;
- export, snapshot, trust, upgrade, and verification heights;
- commands and runner identity needed to reproduce the run.

Resolve tags to commits and images to registry digests before starting. The
source version must be exactly `v30.0.0`; do not substitute `latest` or a moving
branch.

## Local deterministic gate

```sh
go test ./app -run '^TestExportImportPreservesV31State$' -count=1
go test ./x/feepay/types ./x/feepay/keeper -count=1
python3 -m unittest scripts/rehearsal/test_validate_evidence.py
```

The app test performs a real application export into a fresh `InitChain` and
asserts that the FeePay ledger, module backing, wallet-use counters,
voting-snapshot current total, and module version map survive.

## Docker state-sync gate

From `interchaintest/` with the exact candidate image configured:

```sh
make ictest-node
```

`TestStateSync` waits for a real application snapshot, creates a clean node,
state-syncs it through two providers, and verifies at one exact height:

- provider and restored-node app hashes match;
- voting-snapshot parameters match;
- current total voting power matches;
- the node catches the provider tip.

Preserve the test line beginning `state-sync verified:` in the evidence record.

## Mainnet-derived rehearsal

1. Clone the prepared state volume and remove all keyring, node key, and
   private-validator files from the clone.
2. Start the clone with the digest-pinned `v30.0.0` binary and verify its app
   hash against the source node at the same height.
3. Query and save the module version map, FeePay contracts/usages and module
   account backing, and current voting-snapshot totals.
4. Stop v30 and export at the recorded height. Import that export into a clean
   v31 home and start it at the documented initial height.
5. Separately run the repository upgrade interchaintest using exact
   `v30.0.0 -> v31` images.
6. Query the same invariants after import/upgrade. Do not mark the run passed if
   FeePay is under-backed, wallet usages reset, voting power changes, module
   versions regress, or any node fails to catch up.
7. Save the result as JSON and validate it:

```sh
python3 scripts/rehearsal/validate_evidence.py /path/to/evidence.json
```

The evidence JSON must contain the fields enforced by the validator. App hashes
are compared only at identical heights. Amount strings must include their denom.
Attach sanitized logs, the JSON record, and checksums to the release candidate;
never attach homes, databases, keys, or private validator state.
