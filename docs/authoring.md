# Authoring a service schema

`lib-agent-prompt` v1.0 carries one source of truth for every MCP tool's
request and response shape. Adding a new tool means adding three JSON files
under `schemas/services/<mcp>/`.

## Files per tool

| Filename | Purpose |
|---|---|
| `<tool>.request.json` | JSON Schema (draft 2020-12) for the tool's input args. **Required.** |
| `<tool>.response.json` | JSON Schema for the tool's response payload. **Required.** |
| `<tool>.meta.json` | `{requires_permissions, write}` authz hint. **Optional** — absent = "any authenticated user", `write: false`. |

## Request schema discipline

- Declare `"$schema": "https://json-schema.org/draft/2020-12/schema"` at the root.
- Declare `"additionalProperties": false`.
- For **read tools**: include explicit `selection` (which fields to return),
  `filter` (typed predicates), pagination (`limit`, `cursor`), ordering (`sort_by`,
  `order`). Do not accept free-form text-to-SQL inputs.
- For **write tools**: list every required field at the top level. Mark the tool
  `"write": true` in `meta.json`. Add the verb to `requires_permissions`
  (`customers:write`, `transactions:tombstone`, etc.).

## Response schema discipline

- Declare `"additionalProperties": false` so the gateway rejects MCPs that leak
  unintended fields.
- For paginated reads, include `next_cursor` in the schema. The schema is the
  documentation.

## After adding files

1. Run `bundle-builder validate --source-dir .` from the repo root. The validator
   checks request/response pairing and that schemas declare draft-2020-12.
2. Run the unit tests: `go test ./...` from `go/` to confirm the consumer
   library still loads the bundle cleanly.
3. Open a PR. CI rebuilds the bundle, signs it, and the new tool is available
   at the next pinned release.
