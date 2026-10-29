"""The acceptance report must reject missing evidence and mismatched references."""

import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch


ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("validate_release", ROOT / "scripts/validate_release.py")
VALIDATION = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VALIDATION)


class EvidenceTests(unittest.TestCase):
    def test_all_current_members_have_existing_test_links(self):
        entries = VALIDATION.coverage_inventory()
        self.assertEqual(sum(row["status"] == "mapped" for row in entries), 180)
        self.assertEqual(sum(row["status"] == "excluded_deprecated_noop" for row in entries), 2)

    def test_fixture_pin_mismatch_is_rejected(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            (root / "tests/parity").mkdir(parents=True)
            (root / "testdata").mkdir()
            (root / "tests/parity/upstream.json").write_text(json.dumps({"version": "1", "commit": "fixed"}))
            (root / "testdata/example.json").write_text(json.dumps({"upstream_version": "1", "upstream_commit": "other", "cases": []}))
            with patch.object(VALIDATION, "ROOT", root):
                with self.assertRaisesRegex(ValueError, "upstream pin mismatch"):
                    VALIDATION.fixture_inventory()

    def test_new_unmapped_member_is_rejected(self):
        original = Path.read_text

        def read(path, *args, **kwargs):
            text = original(path, *args, **kwargs)
            if path.name == "secure-surface.json":
                data = json.loads(text)
                data["members"].append({"python": "new_unmapped_method", "go": "NewUnmappedMethod"})
                return json.dumps(data)
            return text

        with patch.object(Path, "read_text", read):
            with self.assertRaisesRegex(ValueError, "Unmapped member"):
                VALIDATION.coverage_inventory()

    def test_broken_test_link_is_rejected(self):
        original = Path.read_text

        def read(path, *args, **kwargs):
            text = original(path, *args, **kwargs)
            if path.name == "coverage.json":
                data = copy.deepcopy(json.loads(text))
                data["groups"][0]["tests"] = ["client_test.go#TestDoesNotExist"]
                return json.dumps(data)
            return text

        with patch.object(Path, "read_text", read):
            with self.assertRaisesRegex(ValueError, "Missing test"):
                VALIDATION.coverage_inventory()


if __name__ == "__main__":
    unittest.main()
