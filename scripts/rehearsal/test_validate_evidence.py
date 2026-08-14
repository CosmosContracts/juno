#!/usr/bin/env python3

import copy
import importlib.util
import unittest
from pathlib import Path

MODULE = Path(__file__).with_name("validate_evidence.py")
SPEC = importlib.util.spec_from_file_location("validate_evidence", MODULE)
assert SPEC is not None
validator = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(validator)


def valid_record():
    return {
        "source": {
            "version": "v30.0.0",
            "git_commit": "1" * 40,
            "image_digest": "sha256:" + "2" * 64,
            "chain_id": "juno-1",
        },
        "target": {
            "version": "v31",
            "git_commit": "3" * 40,
            "image_digest": "sha256:" + "4" * 64,
        },
        "export_import": {
            "export_height": 100,
            "pre_export_app_hash": "AA",
            "post_import_app_hash": "AB",
            "module_version_map_preserved": True,
        },
        "state_sync": {
            "snapshot_height": 120,
            "trust_height": 110,
            "verified_height": 130,
            "provider_app_hash": "BB",
            "synced_app_hash": "BB",
        },
        "modules": {
            "feepay": {
                "ledger_total": "1000000ujuno",
                "module_backing": "1000000ujuno",
                "wallet_usages_preserved": True,
            },
            "voting_snapshot": {
                "pre_restart_total": "42",
                "post_restart_total": "42",
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
    def test_accepts_complete_consistent_record(self):
        validator.validate(valid_record())

    def test_rejects_fee_pay_backing_mismatch(self):
        record = copy.deepcopy(valid_record())
        record["modules"]["feepay"]["module_backing"] = "999999ujuno"
        with self.assertRaisesRegex(ValueError, "not fully backed"):
            validator.validate(record)

    def test_rejects_state_sync_hash_mismatch(self):
        record = copy.deepcopy(valid_record())
        record["state_sync"]["synced_app_hash"] = "CC"
        with self.assertRaisesRegex(ValueError, "state-sync app hashes differ"):
            validator.validate(record)

    def test_rejects_incomplete_gate(self):
        record = copy.deepcopy(valid_record())
        record["result"]["upgrade_passed"] = False
        with self.assertRaisesRegex(ValueError, "mandatory rehearsal gates"):
            validator.validate(record)


if __name__ == "__main__":
    unittest.main()
