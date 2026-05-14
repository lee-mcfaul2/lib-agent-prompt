#!/usr/bin/env bash
# Generate Go bindings from the 4 envelope schemas using quicktype.
# Requires: quicktype on PATH (npm install -g quicktype).
#
# Only the top-level envelope schemas are generated here.  Service-specific
# schemas (schemas/services/**) and shared primitives (schemas/shared/*) are
# intentionally excluded — they are consumed by the per-MCP repos directly.
set -euo pipefail
cd "$(dirname "$0")"

OUT=generated
rm -rf "$OUT"
mkdir -p "$OUT"

# Explicit list — mirrors codegen.config.yaml sources.
declare -A SCHEMAS=(
  [user_prompt]="../schemas/user-prompt.json"
  [final_response]="../schemas/final-response.json"
  [tool_result]="../schemas/tool-result.json"
  [bundle_manifest]="../schemas/bundle-manifest.json"
)

for base in user_prompt final_response tool_result bundle_manifest; do
  schema="${SCHEMAS[$base]}"
  quicktype --src-lang schema --lang go --package types --top-level "$base" --just-types -o "$OUT/${base}.go" "$schema" || true
  # quicktype --just-types omits the package directive; prepend it.
  if [[ -f "$OUT/${base}.go" ]] && ! head -1 "$OUT/${base}.go" | grep -q '^package '; then
    { echo "package types"; echo; cat "$OUT/${base}.go"; } > "$OUT/${base}.go.tmp"
    mv "$OUT/${base}.go.tmp" "$OUT/${base}.go"
  fi
done

go fmt ./generated/... 2>/dev/null || true

echo "Generated $(ls $OUT/*.go 2>/dev/null | wc -l) Go files"
