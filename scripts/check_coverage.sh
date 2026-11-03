#!/bin/sh
# Keep local validation and CI on the same whole-repository coverage gate.
set -eu
report=$(mktemp "${TMPDIR:-/tmp}/polymarket-go-coverage.XXXXXX")
trap 'rm -f "$report"' 0
CGO_ENABLED=0 go test -coverpkg=./... -coverprofile="$report" ./...
coverage_summary=$(go tool cover -func="$report")
printf '%s\n' "$coverage_summary" | awk '/^total:/ { print; found = 1; if ($3 + 0 < 80) exit 1 } END { if (!found) exit 1 }'
