#!/usr/bin/env bash
# Generate Go bindings from schemas/*.json using quicktype.
# Requires: quicktype on PATH (npm install -g quicktype).
set -euo pipefail
cd "$(dirname "$0")"

OUT=generated
rm -rf "$OUT"
mkdir -p "$OUT"

shopt -s nullglob globstar
for schema in ../schemas/**/*.json ../schemas/*.json; do
  [[ -f "$schema" ]] || continue
  base=$(basename "$schema" .json | tr '-' '_')
  quicktype --src-lang schema --lang go --package types --top-level "$base" --just-types -o "$OUT/${base}.go" "$schema" || true
  # quicktype --just-types omits the package directive; prepend it.
  if [[ -f "$OUT/${base}.go" ]] && ! head -1 "$OUT/${base}.go" | grep -q '^package '; then
    { echo "package types"; echo; cat "$OUT/${base}.go"; } > "$OUT/${base}.go.tmp"
    mv "$OUT/${base}.go.tmp" "$OUT/${base}.go"
  fi
done

go fmt ./generated/... 2>/dev/null || true

echo "Generated $(ls $OUT/*.go 2>/dev/null | wc -l) Go files"
