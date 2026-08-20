#!/usr/bin/env python3

import importlib.util
import unittest
from pathlib import Path

MODULE = Path(__file__).with_name("validate_evidence.py")
SPEC = importlib.util.spec_from_file_location("validate_evidence", MODULE)
assert SPEC is not None
validator = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(validator)

APP_HASH_A = "a" * 64
APP_HASH_B = "b" * 64
CANONICAL_HASH_A = "c" * 64


def valid_record():
    return {
        "source": {
            "version": "v30.0.0",
            "git_commit": "1" * 40,
            "image_digest": "sha256:" + "2" * 64,
            "chain_id": "juno-1",
            "state_provenance": "sanitized clone of mainnet snapshot provider-1 at height 100",
        },
        "target": {
            "version": "v31",
            "git_commit": "3" * 40,
            "image_digest": "sha256:" + "4" * 64,
        },
        "runner": {
            "identity": "release-engineer@example.invalid",
            "environment": "ci-runner-17/linux-amd64",
        },
        "commands": {
            "export_import": ["junod export --height 100", "junod init && junod start"],
            "state_sync": ["make ictest-node"],
            "upgrade": ["make ictest-upgrade"],
        },
        "export_import": {
            "export_height": 100,
            "pre_export_app_hash_height": 100,
            "pre_export_app_hash": APP_HASH_A,
            "post_import_app_hash_height": 101,
            "post_import_app_hash": APP_HASH_B,
            "module_version_map": {
                "pre_export": {"count": 3, "sha256": CANONICAL_HASH_A},
                "post_import": {"count": 3, "sha256": CANONICAL_HASH_A},
            },
        },
        "state_sync": {
            "snapshot_height": 120,
            "trust_height": 110,
            "verified_height": 130,
            "provider_app_hash_height": 130,
            "provider_app_hash": APP_HASH_A,
            "synced_app_hash_height": 130,
            "synced_app_hash": APP_HASH_A,
        },
        "upgrade": {
            "upgrade_height": 140,
            "verified_height": 141,
        },
        "modules": {
            "feepay": {
                "pre_restart": {
                    "height": 100,
                    "ledger_total": "1000000000000000000000000000000ujuno,7ibc/ABC",
                    "module_backing": "1000000000000000000000000000001ujuno,9ibc/ABC,5uatom",
                    "wallet_usages": {"count": 2, "sha256": CANONICAL_HASH_A},
                },
                "post_restart": {
                    "height": 101,
                    "ledger_total": "1000000000000000000000000000000ujuno,7ibc/ABC",
                    "module_backing": "1000000000000000000000000000001ujuno,9ibc/ABC,5uatom",
                    "wallet_usages": {"count": 2, "sha256": CANONICAL_HASH_A},
                },
            },
            "voting_snapshot": {
                "pre_restart": {"height": 100, "total": "42"},
                "post_restart": {"height": 101, "total": "42"},
                "queryable": True,
            },
        },
        "result": {
            "export_import_passed": True,
            "state_sync_passed": True,
            "upgrade_passed": True,
        },
    }


class EvidenceValidationTest(unittest.TestCase):
    def test_accepts_complete_consistent_record_with_surplus_backing(self):
        validator.validate(valid_record())

    def test_accepts_arbitrary_precision_coins(self):
        record = valid_record()
        huge = "9" * 200
        for phase in ("pre_restart", "post_restart"):
            record["modules"]["feepay"][phase]["ledger_total"] = f"{huge}ujuno"
            record["modules"]["feepay"][phase]["module_backing"] = f"1{huge}ujuno"
        validator.validate(record)

    def test_rejects_missing_source_provenance(self):
        record = valid_record()
        del record["source"]["state_provenance"]
        with self.assertRaisesRegex(ValueError, "state_provenance"):
            validator.validate(record)

    def test_rejects_missing_runner_identity(self):
        record = valid_record()
        record["runner"]["identity"] = ""
        with self.assertRaisesRegex(ValueError, "runner.identity"):
            validator.validate(record)

    def test_rejects_missing_reproduction_command(self):
        record = valid_record()
        record["commands"]["upgrade"] = []
        with self.assertRaisesRegex(ValueError, "commands.upgrade"):
            validator.validate(record)

    def test_rejects_boolean_or_nonpositive_height(self):
        for value in (True, 0, "100"):
            with self.subTest(value=value):
                record = valid_record()
                record["export_import"]["export_height"] = value
                with self.assertRaisesRegex(ValueError, "export_import.export_height"):
                    validator.validate(record)

    def test_rejects_hash_without_exact_matching_height(self):
        record = valid_record()
        record["state_sync"]["synced_app_hash_height"] = 129
        with self.assertRaisesRegex(ValueError, "same verified_height"):
            validator.validate(record)

    def test_rejects_post_import_height_at_or_before_export_height(self):
        for post_height in (100, 99):
            with self.subTest(post_height=post_height):
                record = valid_record()
                record["export_import"]["post_import_app_hash_height"] = post_height
                record["modules"]["feepay"]["post_restart"]["height"] = post_height
                record["modules"]["voting_snapshot"]["post_restart"]["height"] = post_height
                with self.assertRaisesRegex(ValueError, "must be after export_height"):
                    validator.validate(record)

    def test_rejects_unsupported_preservation_booleans(self):
        record = valid_record()
        del record["export_import"]["module_version_map"]
        record["export_import"]["module_version_map_preserved"] = True
        with self.assertRaisesRegex(ValueError, "module_version_map"):
            validator.validate(record)

        record = valid_record()
        for phase in ("pre_restart", "post_restart"):
            del record["modules"]["feepay"][phase]["wallet_usages"]
        record["modules"]["feepay"]["wallet_usages_preserved"] = True
        with self.assertRaisesRegex(ValueError, "wallet_usages"):
            validator.validate(record)

    def test_rejects_changed_module_version_map_evidence(self):
        for field, value in (("count", 4), ("sha256", "d" * 64)):
            with self.subTest(field=field):
                record = valid_record()
                record["export_import"]["module_version_map"]["post_import"][field] = value
                with self.assertRaisesRegex(ValueError, "module version map changed"):
                    validator.validate(record)

    def test_rejects_changed_or_malformed_wallet_usage_evidence(self):
        record = valid_record()
        record["modules"]["feepay"]["post_restart"]["wallet_usages"]["sha256"] = "d" * 64
        with self.assertRaisesRegex(ValueError, "wallet usages changed"):
            validator.validate(record)

        for field, value in (("count", True), ("count", -1), ("sha256", "ABC")):
            with self.subTest(field=field, value=value):
                record = valid_record()
                record["modules"]["feepay"]["pre_restart"]["wallet_usages"][field] = value
                with self.assertRaisesRegex(ValueError, "wallet_usages"):
                    validator.validate(record)

    def test_rejects_invalid_app_hashes(self):
        for bad_hash in ("AA", "A" * 64, "a" * 63, "g" * 64, 123):
            with self.subTest(app_hash=bad_hash):
                record = valid_record()
                record["export_import"]["pre_export_app_hash"] = bad_hash
                with self.assertRaisesRegex(ValueError, "64-character lowercase hex"):
                    validator.validate(record)

    def test_rejects_fee_pay_under_backing_per_denom(self):
        record = valid_record()
        record["modules"]["feepay"]["post_restart"]["module_backing"] = (
            "999999999999999999999999999999ujuno,9ibc/ABC"
        )
        with self.assertRaisesRegex(ValueError, "not fully backed"):
            validator.validate(record)

    def test_rejects_malformed_or_missing_denom_coin_evidence(self):
        for amount in ("1000000", "1.5ujuno", "-1ujuno", "1ujuno,2ujuno", {"amount": 1}):
            with self.subTest(amount=amount):
                record = valid_record()
                record["modules"]["feepay"]["pre_restart"]["ledger_total"] = amount
                with self.assertRaisesRegex(ValueError, "coin"):
                    validator.validate(record)

    def test_rejects_changed_fee_pay_ledger(self):
        record = valid_record()
        record["modules"]["feepay"]["post_restart"]["ledger_total"] = "8ibc/ABC,1000000000000000000000000000000ujuno"
        with self.assertRaisesRegex(ValueError, "ledger changed"):
            validator.validate(record)

    def test_rejects_nonnumeric_or_nonpositive_voting_total(self):
        for total in ("forty-two", "0", "-1", 42, True):
            with self.subTest(total=total):
                record = valid_record()
                record["modules"]["voting_snapshot"]["post_restart"]["total"] = total
                with self.assertRaisesRegex(ValueError, "positive integer string"):
                    validator.validate(record)

    def test_rejects_changed_voting_total(self):
        record = valid_record()
        record["modules"]["voting_snapshot"]["post_restart"]["total"] = "43"
        with self.assertRaisesRegex(ValueError, "current total changed"):
            validator.validate(record)

    def test_rejects_state_sync_hash_mismatch(self):
        record = valid_record()
        record["state_sync"]["synced_app_hash"] = "c" * 64
        with self.assertRaisesRegex(ValueError, "state-sync app hashes differ"):
            validator.validate(record)

    def test_rejects_incomplete_gate(self):
        record = valid_record()
        record["result"]["upgrade_passed"] = False
        with self.assertRaisesRegex(ValueError, "mandatory rehearsal gates"):
            validator.validate(record)


if __name__ == "__main__":
    unittest.main()
