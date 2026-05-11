#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

CONTAINER=${CONTAINER:-docker}
if ! command -v "$CONTAINER" >/dev/null; then
  echo "$CONTAINER not available, skipping" >&2
  exit 0
fi

REGISTRY=localhost:5000

echo "=== Starting zot ==="
"$CONTAINER" run -d --rm --name agentprompt-zot \
  -p 5000:5000 \
  -v "$REPO_ROOT/tests/e2e/zot-config.json:/etc/zot/config.json:ro" \
  ghcr.io/project-zot/zot-linux-amd64:latest >/dev/null
trap '"$CONTAINER" stop agentprompt-zot >/dev/null 2>&1 || true' EXIT

# Wait for registry
for _ in $(seq 1 30); do
  if curl -fs "http://$REGISTRY/v2/" >/dev/null; then
    break
  fi
  sleep 1
done

echo "=== Build and pack ==="
make bundle-example

echo "=== Push ==="
go run ./tools/bundle-builder push ./out/bundle-0.1.0-example.tar.gz "$REGISTRY/ai-security/bundles/example:latest"

echo "=== Fetch + verify via Go loader ==="
AGENT_PROMPT_OCI_TEST_REGISTRY="$REGISTRY" \
  go test -run TestLoad_OCI_Smoke -v ./go/...

echo "=== With-zot smoke OK ==="
