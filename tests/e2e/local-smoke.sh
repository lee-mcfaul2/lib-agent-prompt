#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

BUNDLE_OUT="$REPO_ROOT/out/smoke-bundle"

echo "=== make schemas-validate ==="
make schemas-validate

echo "=== Bundle build (new flags) ==="
go run ./tools/bundle-builder build \
  --source-dir . \
  --out-dir "$BUNDLE_OUT" \
  --bundle-version 1.0.0 \
  --schema-library-version 1.0.0 \
  --source-commit "$(git rev-parse --short HEAD)" \
  --builder-id local-smoke \
  --max-iterations 8 \
  --max-wallclock-ms 300000 \
  --max-cost-usd 1.0

echo "=== Assert bundle shape ==="

# Envelope schemas
test -f "$BUNDLE_OUT/schemas/user-prompt.json"   || { echo "FAIL: missing user-prompt.json";   exit 1; }
test -f "$BUNDLE_OUT/schemas/final-response.json" || { echo "FAIL: missing final-response.json"; exit 1; }
test -f "$BUNDLE_OUT/schemas/tool-result.json"   || { echo "FAIL: missing tool-result.json";   exit 1; }

# Service dirs
test -d "$BUNDLE_OUT/schemas/services/kb"       || { echo "FAIL: missing services/kb/";       exit 1; }
test -d "$BUNDLE_OUT/schemas/services/customer" || { echo "FAIL: missing services/customer/"; exit 1; }

# Manifest shape assertions
jq -e '.services | length >= 1' "$BUNDLE_OUT/bundle-manifest.json" > /dev/null \
  || { echo "FAIL: manifest.services empty or missing"; exit 1; }
jq -e '.services[] | select(.mcp == "kb") | .tools | length >= 2' "$BUNDLE_OUT/bundle-manifest.json" > /dev/null \
  || { echo "FAIL: kb tools missing"; exit 1; }
jq -e 'has("prompts") | not' "$BUNDLE_OUT/bundle-manifest.json" > /dev/null \
  || { echo "FAIL: manifest still has prompts[] field"; exit 1; }

echo "=== Validate sample-user-prompt fixture against bundle schema ==="
go run ./tests/e2e/validate-fixture \
  --schema "$BUNDLE_OUT/schemas/user-prompt.json" \
  --instance tests/fixtures/sample-user-prompt.json \
  || { echo "FAIL: sample-user-prompt invalid"; exit 1; }

echo "=== Go library tests ==="
go test ./go/...

echo "=== Bundle-builder tests ==="
go test ./tools/bundle-builder/...

echo "=== Verifier tests ==="
go test ./pkg/verify/... ./tools/verifier/...

echo "=== Local smoke OK ==="
