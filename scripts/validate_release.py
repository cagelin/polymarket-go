"""Record reproducible offline acceptance evidence without account credentials."""

import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
import platform
from pathlib import Path
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[1]
TARGETS = [(system, arch) for system in ("linux", "darwin", "windows")
           for arch in ("amd64", "arm64")]


def source_hashes():
    names = subprocess.check_output(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"], cwd=ROOT
    ).decode().split("\0")
    return {name: hashlib.sha256((ROOT / name).read_bytes()).hexdigest()
            for name in sorted(set(names)) if name and (ROOT / name).is_file()}


def run(command, output, env):
    print("Running:", " ".join(map(str, command)), flush=True)
    with output.open("w") as log:
        subprocess.run(command, cwd=ROOT, env=env, stdout=log,
                       stderr=subprocess.STDOUT, check=True)


def fixture_inventory():
    upstream = json.loads((ROOT / "tests/parity/upstream.json").read_text())
    entries = []
    for path in sorted((ROOT / "testdata").glob("*.json")):
        raw = path.read_bytes()
        content = json.loads(raw)
        if isinstance(content, dict):
            if content.get("upstream_version") != upstream["version"] or content.get("upstream_commit") != upstream["commit"]:
                raise ValueError(f"{path.name}: upstream pin mismatch")
            groups = {key: len(value) for key, value in content.items() if isinstance(value, list)}
        else:
            groups = {"cases": len(content)}
        entries.append({"file": path.relative_to(ROOT).as_posix(),
                        "sha256": hashlib.sha256(raw).hexdigest(), "groups": groups})
    return entries


def coverage_inventory():
    """Resolve member-to-test links; mapping is not a claim of live acceptance."""
    coverage = json.loads((ROOT / "tests/acceptance/coverage.json").read_text())
    fixture_tests = coverage["fixture_tests"]
    actual = {path.relative_to(ROOT).as_posix() for path in (ROOT / "testdata").glob("*-parity.json")}
    if actual != set(fixture_tests):
        raise ValueError("Fixture coverage differs: " + str(sorted(actual ^ set(fixture_tests))))
    direct = {}
    for filename in fixture_tests:
        content = json.loads((ROOT / filename).read_text())
        rows = content if isinstance(content, list) else content.get("cases", [])
        for row in rows:
            if "operation" in row:
                direct.setdefault(row["operation"], set()).add(filename)
    entries = []
    for client in ("public", "secure"):
        surface = json.loads((ROOT / f"testdata/{client}-surface.json").read_text())
        for member in surface["members"]:
            name = member["python"]
            if member.get("excluded"):
                if name not in ("setup_gasless_wallet", "is_gasless_ready"):
                    raise ValueError(f"Unreviewed exclusion: {name}")
                entries.append({"client": client, **member, "status": "excluded_deprecated_noop"})
                continue
            sources = set(direct.get(name, []))
            tests = set()
            for group in coverage["groups"]:
                if name in group["members"]:
                    sources.update(group["fixtures"])
                    tests.update(group["tests"])
            for filename in sources:
                tests.update(fixture_tests[filename])
            if not tests:
                raise ValueError(f"Unmapped member: {client}.{name}")
            for ref in tests:
                filename, function = ref.split("#")
                if f"func {function}(" not in (ROOT / filename).read_text():
                    raise ValueError(f"Missing test: {ref}")
            entries.append({"client": client, **member,
                            "go": "NewSecureClient" if name == "create" else member["go"],
                            "fixtures": sorted(sources), "tests": sorted(tests),
                            "status": "mapped", "live": "not_claimed"})
    return entries


def fixtures(output, env, report):
    before = source_hashes()
    run([sys.executable, "scripts/regenerate.py"], output / "regenerate.log", env)
    after = source_hashes()
    changed = sorted(name for name in before.keys() | after.keys() if before.get(name) != after.get(name))
    inventory = fixture_inventory()
    coverage = coverage_inventory()
    report["fixtures"] = {"reproducible": not changed, "changed_files": changed,
                          "inventory": inventory,
                          "case_count": sum(count for item in inventory for group, count in item["groups"].items() if group != "members")}
    (output / "coverage.json").write_text(json.dumps(coverage, indent=2) + "\n")
    if changed:
        raise ValueError("Generated output changed: " + ", ".join(changed))


def platforms(output, env, report):
    report["platforms"] = []
    packages = subprocess.check_output(
        ["go", "list", "-f", "{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}", "./..."],
        cwd=ROOT, env=env).decode().split()
    for system, arch in TARGETS:
        label = f"{system}-{arch}"
        binaries = output / label
        binaries.mkdir()
        target_env = dict(env, GOOS=system, GOARCH=arch, CGO_ENABLED="0")
        result = {"os": system, "arch": arch, "cgo": False,
                  "build": "not_run", "runtime": "not_run"}
        report["platforms"].append(result)
        try:
            run(["go", "build", "./..."], output / f"{label}-build.log", target_env)
            for package in packages:
                folder = binaries.joinpath(*package.split("/"))
                folder.mkdir(parents=True, exist_ok=True)
                suffix = ".exe" if system == "windows" else ""
                run(["go", "test", "-c", "-o", str(folder / ("suite.test" + suffix)), package],
                    folder / "build.log", target_env)
        except subprocess.CalledProcessError:
            result["build"] = "failed"
            raise
        result["build"] = "passed"
    run(["go", "test", "-json", "-count=1", "./..."], output / "host-tests.jsonl", env)
    events = [json.loads(line) for line in (output / "host-tests.jsonl").read_text().splitlines()
              if line.startswith("{")]
    for result in report["platforms"]:
        if (result["os"], result["arch"]) == (report["go"]["GOHOSTOS"], report["go"]["GOHOSTARCH"]):
            result["runtime"] = "offline_passed"
    report["skipped_tests"] = [{"package": event["Package"], "test": event["Test"]}
                               for event in events if event.get("Action") == "skip" and "Test" in event]


def benchmarks(output, env, report):
    run(["go", "test", "-run", "^$", "-bench", ".", "-benchmem", "-benchtime=300ms", "-count=3",
         ".", "./internal/signing", "./streams", "./models"], output / "benchmarks.txt", env)
    report["benchmarks"] = {"status": "passed", "samples": 3, "benchtime": "300ms",
                            "network": "none", "threshold": None}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("step", choices=["fixtures", "platforms", "benchmarks", "all"])
    args = parser.parse_args()
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S.%fZ")
    output = ROOT / ".local/validation" / stamp
    output.mkdir(parents=True)
    # Never inherit opt-in live switches, wallet credentials, or cross-compile targets.
    env = {key: value for key, value in os.environ.items()
           if not key.startswith("POLYMARKET_") and key not in ("GOOS", "GOARCH", "CGO_ENABLED")}
    env["CGO_ENABLED"] = "0"
    report = {
        "created_at": stamp, "step": args.step, "status": "running",
        "commit": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT).decode().strip(),
        "dirty": bool(subprocess.check_output(["git", "status", "--porcelain"], cwd=ROOT)),
        "source_hashes": source_hashes(),
        "host": platform.platform(),
        "go": json.loads(subprocess.check_output(["go", "env", "-json", "GOVERSION", "GOHOSTOS", "GOHOSTARCH"], env=env)),
        "upstream": {key: value for key, value in json.loads((ROOT / "tests/parity/upstream.json").read_text()).items()
                     if key in ("version", "commit")},
        "funded_validation": "not_run", "session_key_live_validation": "not_run",
    }
    try:
        for name, action in [("fixtures", fixtures), ("platforms", platforms), ("benchmarks", benchmarks)]:
            if args.step in (name, "all"):
                action(output, env, report)
        report["status"] = "passed"
    except (subprocess.CalledProcessError, ValueError, OSError) as exc:
        report["status"] = "failed"
        report["error"] = str(exc)
    finally:
        (output / "report.json").write_text(json.dumps(report, indent=2) + "\n")
        print(f"{report['status']}: {output / 'report.json'}", flush=True)
    return 0 if report["status"] == "passed" else 1


if __name__ == "__main__":
    sys.exit(main())
