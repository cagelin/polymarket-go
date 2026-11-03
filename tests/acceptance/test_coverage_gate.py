"""The coverage gate must propagate failures and clean temporary reports."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]


class CoverageGateTests(unittest.TestCase):
    def test_threshold_tool_failures_and_cleanup(self):
        scenarios = [
            ("below", {"GATE_SUMMARY": "total: (statements) 79.9%"}, 1),
            ("threshold", {"GATE_SUMMARY": "total: (statements) 80.0%"}, 0),
            ("above", {"GATE_SUMMARY": "total: (statements) 80.4%"}, 0),
            ("test_failure", {"GATE_TEST_EXIT": "7"}, 7),
            ("missing_total", {"GATE_SUMMARY": "missing total"}, 1),
            ("cover_failure", {"GATE_COVER_EXIT": "8"}, 8),
        ]
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            stub = root / "go"
            stub.write_text("""#!/bin/sh
case "$1" in
 test) exit "${GATE_TEST_EXIT:-0}" ;;
 tool) printf '%s\\n' "${GATE_SUMMARY:-total: (statements) 80.4%}"; exit "${GATE_COVER_EXIT:-0}" ;;
 *) exit 9 ;;
esac
""")
            stub.chmod(0o755)
            for name, overrides, expected in scenarios:
                with self.subTest(name=name):
                    env = {key: value for key, value in os.environ.items()
                           if not key.startswith("GATE_")}
                    env.update(PATH=str(root) + os.pathsep + env["PATH"], TMPDIR=folder)
                    env.update(overrides)
                    result = subprocess.run(
                        ["sh", str(ROOT / "scripts/check_coverage.sh")],
                        cwd=ROOT, env=env, capture_output=True, text=True, check=False,
                    )
                    self.assertEqual(result.returncode, expected, result.stdout + result.stderr)
                    self.assertEqual(list(root.glob("polymarket-go-coverage.*")), [])


if __name__ == "__main__":
    unittest.main()
