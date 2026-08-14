#!/usr/bin/env python3
"""Validate evidence from a v30.0.0 -> v31 state rehearsal."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any, NoReturn

APP_HASH = re.compile(r"^[0-9a-f]{64}$")
GIT_SHA = re.compile(r"^[0-9a-f]{40}$")
COIN = re.compile(r"([0-9]+)([A-Za-z][A-Za-z0-9/:._-]{2,127})")
POSITIVE_INTEGER = re.compile(r"^[1-9][0-9]*$")


def fail(message: str) -> NoReturn:
    raise ValueError(message)


def object_field(mapping: dict[str, Any], key: str, path: str) -> dict[str, Any]:
    if key not in mapping:
        fail(f"missing required field: {path}.{key}")
    value = mapping[key]
    if not isinstance(value, dict):
        fail(f"{path}.{key} must be an object")
    return value


def required(mapping: dict[str, Any], path: str, *keys: str) -> None:
    for key in keys:
        if key not in mapping:
            fail(f"missing required field: {path}.{key}")


def nonempty_string(value: Any, name: str) -> str:
    if not isinstance(value, str) or not value.strip():
        fail(f"{name} must be a non-empty string")
    return value


def height(value: Any, name: str) -> int:
    # bool is an int subclass, but cannot be accepted as numeric evidence.
    if isinstance(value, bool) or not isinstance(value, int) or value <= 0:
        fail(f"{name} must be a positive integer")
    return value


def app_hash(value: Any, name: str) -> str:
    if not isinstance(value, str) or not APP_HASH.fullmatch(value):
        fail(f"{name} must be a 64-character lowercase hex app hash")
    return value


def parse_coins(value: Any, name: str) -> dict[str, int]:
    """Parse canonical comma-separated Cosmos coins using arbitrary-size ints."""
    if not isinstance(value, str) or not value:
        fail(f"{name} must be a non-empty coin string with denoms")

    coins: dict[str, int] = {}
    for item in value.split(","):
        match = COIN.fullmatch(item)
        if match is None:
            fail(f"{name} contains an invalid coin: {item!r}")
        amount_text, denom = match.groups()
        amount = int(amount_text)
        if amount <= 0:
            fail(f"{name} coin amounts must be positive")
        if denom in coins:
            fail(f"{name} contains duplicate coin denom: {denom}")
        coins[denom] = amount
    return coins


def positive_integer_string(value: Any, name: str) -> int:
    # JSON strings preserve exact values emitted by Cosmos SDK integer queries.
    if not isinstance(value, str) or not POSITIVE_INTEGER.fullmatch(value):
        fail(f"{name} must be a positive integer string")
    return int(value)


def validate(record: dict[str, Any]) -> None:
    for key in ("source", "target", "runner", "commands", "export_import", "state_sync", "upgrade", "modules", "result"):
        object_field(record, key, "record")

    source = record["source"]
    target = record["target"]
    runner = record["runner"]
    commands = record["commands"]
    export_import = record["export_import"]
    state_sync = record["state_sync"]
    upgrade = record["upgrade"]
    modules = record["modules"]
    result = record["result"]

    required(source, "source", "version", "git_commit", "image_digest", "chain_id", "state_provenance")
    required(target, "target", "version", "git_commit", "image_digest")
    if source["version"] != "v30.0.0":
        fail("source.version must be v30.0.0")
    if target["version"] != "v31":
        fail("target.version must be v31")
    nonempty_string(source["chain_id"], "source.chain_id")
    nonempty_string(source["state_provenance"], "source.state_provenance")
    for name, value in (("source.git_commit", source["git_commit"]), ("target.git_commit", target["git_commit"])):
        if not isinstance(value, str) or not GIT_SHA.fullmatch(value):
            fail(f"{name} must be a full lowercase git SHA")
    for name, value in (("source.image_digest", source["image_digest"]), ("target.image_digest", target["image_digest"])):
        if not isinstance(value, str) or not value.startswith("sha256:") or not APP_HASH.fullmatch(value[7:]):
            fail(f"{name} must be sha256:<64 lowercase hex>")

    required(runner, "runner", "identity", "environment")
    nonempty_string(runner["identity"], "runner.identity")
    nonempty_string(runner["environment"], "runner.environment")
    required(commands, "commands", "export_import", "state_sync", "upgrade")
    for gate in ("export_import", "state_sync", "upgrade"):
        gate_commands = commands[gate]
        if not isinstance(gate_commands, list) or not gate_commands:
            fail(f"commands.{gate} must be a non-empty command list")
        for index, command in enumerate(gate_commands):
            nonempty_string(command, f"commands.{gate}[{index}]")

    required(
        export_import,
        "export_import",
        "export_height",
        "pre_export_app_hash_height",
        "pre_export_app_hash",
        "post_import_app_hash_height",
        "post_import_app_hash",
        "module_version_map_preserved",
    )
    export_height = height(export_import["export_height"], "export_import.export_height")
    pre_export_height = height(export_import["pre_export_app_hash_height"], "export_import.pre_export_app_hash_height")
    post_import_height = height(export_import["post_import_app_hash_height"], "export_import.post_import_app_hash_height")
    if pre_export_height != export_height:
        fail("export_import.pre_export_app_hash_height must equal export_height")
    app_hash(export_import["pre_export_app_hash"], "export_import.pre_export_app_hash")
    app_hash(export_import["post_import_app_hash"], "export_import.post_import_app_hash")
    if export_import["module_version_map_preserved"] is not True:
        fail("module version map was not preserved")

    required(
        state_sync,
        "state_sync",
        "snapshot_height",
        "trust_height",
        "verified_height",
        "provider_app_hash_height",
        "provider_app_hash",
        "synced_app_hash_height",
        "synced_app_hash",
    )
    snapshot_height = height(state_sync["snapshot_height"], "state_sync.snapshot_height")
    trust_height = height(state_sync["trust_height"], "state_sync.trust_height")
    verified_height = height(state_sync["verified_height"], "state_sync.verified_height")
    provider_hash_height = height(state_sync["provider_app_hash_height"], "state_sync.provider_app_hash_height")
    synced_hash_height = height(state_sync["synced_app_hash_height"], "state_sync.synced_app_hash_height")
    if not trust_height < snapshot_height <= verified_height:
        fail("invalid state-sync height ordering")
    if provider_hash_height != verified_height or synced_hash_height != verified_height:
        fail("state-sync app hashes must be recorded at the same verified_height")
    provider_hash = app_hash(state_sync["provider_app_hash"], "state_sync.provider_app_hash")
    synced_hash = app_hash(state_sync["synced_app_hash"], "state_sync.synced_app_hash")
    if provider_hash != synced_hash:
        fail("state-sync app hashes differ")

    required(upgrade, "upgrade", "upgrade_height", "verified_height")
    upgrade_height = height(upgrade["upgrade_height"], "upgrade.upgrade_height")
    upgrade_verified_height = height(upgrade["verified_height"], "upgrade.verified_height")
    if upgrade_verified_height <= upgrade_height:
        fail("upgrade.verified_height must be after upgrade.upgrade_height")

    feepay = object_field(modules, "feepay", "modules")
    voting = object_field(modules, "voting_snapshot", "modules")
    required(feepay, "modules.feepay", "pre_restart", "post_restart", "wallet_usages_preserved")
    fee_ledgers: list[dict[str, int]] = []
    for phase, expected_height in (("pre_restart", export_height), ("post_restart", post_import_height)):
        evidence = object_field(feepay, phase, "modules.feepay")
        required(evidence, f"modules.feepay.{phase}", "height", "ledger_total", "module_backing")
        if height(evidence["height"], f"modules.feepay.{phase}.height") != expected_height:
            fail(f"modules.feepay.{phase}.height does not match its exact restart height")
        ledger = parse_coins(evidence["ledger_total"], f"modules.feepay.{phase}.ledger_total")
        backing = parse_coins(evidence["module_backing"], f"modules.feepay.{phase}.module_backing")
        if any(backing.get(denom, 0) < amount for denom, amount in ledger.items()):
            fail(f"FeePay ledger is not fully backed at {phase}")
        fee_ledgers.append(ledger)
    if fee_ledgers[0] != fee_ledgers[1]:
        fail("FeePay ledger changed across restart")
    if feepay["wallet_usages_preserved"] is not True:
        fail("FeePay wallet usages were not preserved")

    required(voting, "modules.voting_snapshot", "pre_restart", "post_restart", "queryable")
    voting_totals: list[int] = []
    for phase, expected_height in (("pre_restart", export_height), ("post_restart", post_import_height)):
        evidence = object_field(voting, phase, "modules.voting_snapshot")
        required(evidence, f"modules.voting_snapshot.{phase}", "height", "total")
        if height(evidence["height"], f"modules.voting_snapshot.{phase}.height") != expected_height:
            fail(f"modules.voting_snapshot.{phase}.height does not match its exact restart height")
        voting_totals.append(positive_integer_string(evidence["total"], f"modules.voting_snapshot.{phase}.total"))
    if voting_totals[0] != voting_totals[1]:
        fail("voting-snapshot current total changed across restart")
    if voting["queryable"] is not True:
        fail("voting-snapshot is not queryable")

    required(result, "result", "export_import_passed", "state_sync_passed", "upgrade_passed")
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
