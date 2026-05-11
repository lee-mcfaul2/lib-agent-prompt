#!/usr/bin/env bash
set -euo pipefail

# Build a clean bundle and produce three corrupted variants for verifier negative tests.

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
OUT="$REPO_ROOT/out/fixture-corrupt"
rm -rf "$OUT"
mkdir -p "$OUT"

# Build the example bundle
( cd "$REPO_ROOT" && go run ./tools/bundle-builder build \
  --schema-lib ./schemas \
  --prompts ./prompts/example \
  --services ./schemas/service-references \
  --output "$OUT/clean" \
  --version 0.1.0-fixture \
  --allow-placeholder )

cp -r "$OUT/clean" "$OUT/corrupt-manifest"
echo '{"bogus": true}' > "$OUT/corrupt-manifest/bundle-manifest.json"
( cd "$OUT/corrupt-manifest" && tar -czf "$OUT/corrupt-manifest.tar.gz" -- * )

cp -r "$OUT/clean" "$OUT/missing-service"
rm "$OUT/missing-service/service-schemas/postgresql-service.json"
( cd "$OUT/missing-service" && tar -czf "$OUT/missing-service.tar.gz" -- * )

cp -r "$OUT/clean" "$OUT/modified-prompt"
PROMPT_FILE=$(ls "$OUT/modified-prompt/prompts"/*.json | head -1)
python3 -c "import json,sys;d=json.load(open('$PROMPT_FILE'));d['cost_caps']['max_cost_usd']=99999;json.dump(d,open('$PROMPT_FILE','w'))"
( cd "$OUT/modified-prompt" && tar -czf "$OUT/modified-prompt.tar.gz" -- * )

# Pack the clean reference
go run "$REPO_ROOT/tools/bundle-builder" pack "$OUT/clean"

echo "fixtures ready in $OUT/"
