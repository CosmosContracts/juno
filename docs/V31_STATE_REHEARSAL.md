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

Record before execution in the evidence JSON:

- source tag, full commit, and `sha256:` image digest;
- target full commit and `sha256:` image digest;
- chain ID and a non-empty, sanitized description of the source-state
  provenance (provider/snapshot identity and source height, but no secrets);
- export, snapshot, trust, upgrade, and verification heights;
- every command used for the export/import, state-sync, and upgrade gates;
- the runner's identity and execution environment.

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

## Evidence schema

The validator is the executable schema. Use this shape (values are illustrative):

```json
{
  "source": {
    "version": "v30.0.0",
    "git_commit": "<40 lowercase hex>",
    "image_digest": "sha256:<64 lowercase hex>",
    "chain_id": "juno-1",
    "state_provenance": "sanitized snapshot/provider and source-height description"
  },
  "target": {
    "version": "v31",
    "git_commit": "<40 lowercase hex>",
    "image_digest": "sha256:<64 lowercase hex>"
  },
  "runner": {"identity": "operator or CI identity", "environment": "runner/OS/architecture"},
  "commands": {
    "export_import": ["<exact command>", "<exact command>"],
    "state_sync": ["<exact command>"],
    "upgrade": ["<exact command>"]
  },
  "export_import": {
    "export_height": 100,
    "pre_export_app_hash_height": 100,
    "pre_export_app_hash": "<64 lowercase hex>",
    "post_import_app_hash_height": 101,
    "post_import_app_hash": "<64 lowercase hex>",
    "module_version_map_preserved": true
  },
  "state_sync": {
    "snapshot_height": 120,
    "trust_height": 110,
    "verified_height": 130,
    "provider_app_hash_height": 130,
    "provider_app_hash": "<64 lowercase hex>",
    "synced_app_hash_height": 130,
    "synced_app_hash": "<64 lowercase hex>"
  },
  "upgrade": {"upgrade_height": 140, "verified_height": 141},
  "modules": {
    "feepay": {
      "pre_restart": {"height": 100, "ledger_total": "1000000ujuno", "module_backing": "1000001ujuno"},
      "post_restart": {"height": 101, "ledger_total": "1000000ujuno", "module_backing": "1000001ujuno"},
      "wallet_usages_preserved": true
    },
    "voting_snapshot": {
      "pre_restart": {"height": 100, "total": "42"},
      "post_restart": {"height": 101, "total": "42"},
      "queryable": true
    }
  },
  "result": {"export_import_passed": true, "state_sync_passed": true, "upgrade_passed": true}
}
```

All heights are positive JSON integers (not strings or booleans). App hashes
are exactly 64 lowercase hexadecimal characters, and provider/restored hashes
are compared only at the shared `verified_height`. FeePay totals are structured
as pre/post evidence at exact heights. Coin strings are comma-separated positive
arbitrary-precision integer amounts with denoms; backing may contain surplus
amounts or additional denoms, but must cover every ledger denom and amount.
Pre/post ledgers must be equal. Voting totals are positive integer strings,
must be non-zero, and must be equal across the restart.

Attach sanitized logs, the JSON record, and checksums to the release candidate;
never attach homes, databases, keys, or private validator state.
