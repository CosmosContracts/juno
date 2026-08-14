#!/usr/bin/env python3
"""Validate evidence from a v30.0.0 -> v31 state rehearsal."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

SHA256 = re.compile(r"^[0-9a-f]{64}$")
GIT_SHA = re.compile(r"^[0-9a-f]{40}$")


def fail(message: str) -> None:
    raise ValueError(message)


def required(mapping: dict, *keys: str) -> None:
    for key in keys:
        if key not in mapping:
            fail(f"missing required field: {key}")


def validate(record: dict) -> None:
    required(record, "source", "target", "export_import", "state_sync", "modules", "result")
    source = record["source"]
    target = record["target"]
    export_import = record["export_import"]
    state_sync = record["state_sync"]
    modules = record["modules"]
    result = record["result"]

    required(source, "version", "git_commit", "image_digest", "chain_id")
    required(target, "version", "git_commit", "image_digest")
    if source["version"] != "v30.0.0":
        fail("source.version must be v30.0.0")
    if target["version"] != "v31":
        fail("target.version must be v31")
    for name, value in (("source.git_commit", source["git_commit"]), ("target.git_commit", target["git_commit"])):
        if not GIT_SHA.fullmatch(value):
            fail(f"{name} must be a full lowercase git SHA")
    for name, value in (("source.image_digest", source["image_digest"]), ("target.image_digest", target["image_digest"])):
        if not isinstance(value, str) or not value.startswith("sha256:") or not SHA256.fullmatch(value[7:]):
            fail(f"{name} must be sha256:<64 lowercase hex>")

    required(export_import, "export_height", "pre_export_app_hash", "post_import_app_hash", "module_version_map_preserved")
    if export_import["export_height"] <= 0:
        fail("export_height must be positive")
    if not export_import["pre_export_app_hash"] or not export_import["post_import_app_hash"]:
        fail("export/import app hashes must both be recorded")
    if export_import["module_version_map_preserved"] is not True:
        fail("module version map was not preserved")

    required(state_sync, "snapshot_height", "trust_height", "verified_height", "provider_app_hash", "synced_app_hash")
    if not 0 < state_sync["trust_height"] < state_sync["snapshot_height"] <= state_sync["verified_height"]:
        fail("invalid state-sync height ordering")
    if state_sync["provider_app_hash"] != state_sync["synced_app_hash"]:
        fail("state-sync app hashes differ")

    required(modules, "feepay", "voting_snapshot")
    required(modules["feepay"], "ledger_total", "module_backing", "wallet_usages_preserved")
    if modules["feepay"]["ledger_total"] != modules["feepay"]["module_backing"]:
        fail("FeePay ledger is not fully backed")
    if modules["feepay"]["wallet_usages_preserved"] is not True:
        fail("FeePay wallet usages were not preserved")
    required(modules["voting_snapshot"], "pre_restart_total", "post_restart_total", "queryable")
    if modules["voting_snapshot"]["pre_restart_total"] != modules["voting_snapshot"]["post_restart_total"]:
        fail("voting-snapshot current total changed across restart")
    if modules["voting_snapshot"]["queryable"] is not True:
        fail("voting-snapshot is not queryable")

    required(result, "export_import_passed", "state_sync_passed", "upgrade_passed")
    if not all(result[key] is True for key in ("export_import_passed", "state_sync_passed", "upgrade_passed")):
        fail("one or more mandatory rehearsal gates did not pass")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("record", type=Path)
    args = parser.parse_args()
    try:
        record = json.loads(args.record.read_text())
        if not isinstance(record, dict):
            fail("record root must be an object")
        validate(record)
    except (OSError, json.JSONDecodeError, ValueError) as exc:
        print(f"INVALID: {exc}", file=sys.stderr)
        return 1
    print("VALID: v30.0.0 -> v31 rehearsal evidence")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
