# lib-agent-prompt

Schema library and signed-bundle producer for the AI Agent Security Platform.

This repo defines the cross-component schema library (typed request/tool-call/response envelopes and flat per-tool schemas) and produces signed OCI bundles consumed by `agent-sandbox` at startup and validated inline by `agent-gateway`.

**v1.0.0 is a breaking change from 0.x.** See [`docs/migration-0.x-to-1.0.md`](docs/migration-0.x-to-1.0.md) for the full delta.

See `docs/superpowers/specs/2026-05-11-lib-agent-prompt-design.md` in the umbrella workspace for the full design.

## Layout

- `schemas/` — canonical JSON Schema source (single source of truth)
  - `user-prompt.json`, `final-response.json`, `tool-result.json` — typed envelopes
  - `services/<mcp>/<tool>.{request,response,meta}.json` — flat per-tool schemas
  - `shared/` — reusable sub-schemas (PII types, trace context, error envelope, UUID)
- `go/`, `python/`, `rust/`, `java/`, `dotnet/`, `javascript/`, `typescript/` — per-language consumer libraries
- `pkg/verify/` — Go library for supply-chain and structural verification (used by `agent-sandbox`)
- `tools/bundle-builder/` — Go CLI that assembles signed OCI bundles
- `tests/` — schema tests, conformance, reproducibility, e2e
- `docs/` — authoring guides, verification, migration

## Quickstart

```
make schemas-validate   # validate the schema library
make bundle-example     # build the example bundle locally
make verify-example     # verify the locally built bundle
```

## License

Apache 2.0. See `LICENSE`.
