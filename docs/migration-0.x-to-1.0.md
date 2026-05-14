# Migrating from lib-agent-prompt 0.x to 1.0

v1.0 is a deliberate breaking change. The state-machine prompt model is gone;
every MCP traffic shape is now an explicit per-tool schema and the request/response
envelopes are typed.

## What's dropped

| 0.x | Replacement |
|---|---|
| `schemas/iteration-output.json` | Gone. Agent sandbox uses native LLM function-calling. |
| `schemas/prompt.json` | Gone. No per-UUID prompt files. |
| `schemas/allowed-response/*` | Gone. LLM is unconstrained at the iteration level. |
| `schemas/response-envelope.json` | Gone. Gateway's `tool-result` envelope replaces it. |
| `schemas/service-reference.json` + `service-references/` | Gone. Replaced by per-tool schemas. |
| `prompts/*` | Gone. Replaced by one `schemas/user-prompt.json`. |
| `manifest.prompts[]` | Gone. |

## What's added

| New | Purpose |
|---|---|
| `schemas/user-prompt.json` | The single inbound request envelope. |
| `schemas/final-response.json` | The sandbox's terminate envelope shape. |
| `schemas/tool-result.json` | The gateway-to-sandbox tool-call result envelope. |
| `schemas/services/<mcp>/<tool>.{request,response,meta}.json` | Flat per-tool schemas. |

## Manifest changes

`bundle-manifest.json` now has `services: [{mcp, tools: [{name, request_digest, response_digest, write?, requires_permissions?}]}]` and no `prompts[]` array. See `schemas/bundle-manifest.json` for the full meta-schema.

## Downstream rebuild order

1. **agent-sql-mcp** — bump bundled digest reference. Its 8 per-tool schemas now live canonically in `lib-agent-prompt/schemas/services/customer/`.
2. **agent-sandbox** — bump bundled digest reference; refresh fixtures.
3. **agent-gateway** — separate per-repo plan (drop iteration-output parser, switch to native function-calling, add `/v1/bundle_digest`).
4. **secure-agent-demo** — Helm values for the new image/chart digests.

Consumers hold an explicit OCI digest, so the 0.x artifact stays available indefinitely. No flag day.
