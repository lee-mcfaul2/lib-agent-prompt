#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

echo "=== make schemas-validate ==="
make schemas-validate

echo "=== make demo ==="
make demo

echo "=== Go library tests ==="
go test ./go/...

echo "=== Bundle-builder tests ==="
go test ./tools/bundle-builder/...

echo "=== Verifier tests ==="
go test ./pkg/verify/... ./tools/verifier/...

echo "=== Local smoke OK ==="
