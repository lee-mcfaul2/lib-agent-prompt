#!/usr/bin/env bash
set -euo pipefail

# Build a clean bundle and produce corrupted variants for verifier negative tests.

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
OUT="$REPO_ROOT/out/fixture-corrupt"
rm -rf "$OUT"
mkdir -p "$OUT"

# Build the clean bundle using the new flags.
( cd "$REPO_ROOT" && go run ./tools/bundle-builder build \
  --source-dir . \
  --out-dir "$OUT/clean" \
  --bundle-version 0.1.0-fixture \
  --builder-id fixture-builder )

# Pack the clean reference
( cd "$REPO_ROOT" && go run ./tools/bundle-builder pack "$OUT/clean" )

# Corrupt-manifest variant: replace manifest with invalid JSON.
cp -r "$OUT/clean" "$OUT/corrupt-manifest"
echo '{"bogus": true}' > "$OUT/corrupt-manifest/bundle-manifest.json"
( cd "$OUT/corrupt-manifest" && tar -czf "$OUT/corrupt-manifest.tar.gz" -- * )

# Missing-service variant: remove one service schema dir.
cp -r "$OUT/clean" "$OUT/missing-service"
# Remove the first MCP directory found under schemas/services/.
FIRST_MCP=$(ls "$OUT/missing-service/schemas/services/" | head -1)
rm -rf "$OUT/missing-service/schemas/services/$FIRST_MCP"
( cd "$OUT/missing-service" && tar -czf "$OUT/missing-service.tar.gz" -- * )

echo "fixtures ready in $OUT/"
