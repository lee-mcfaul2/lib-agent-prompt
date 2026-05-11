# lib-agent-prompt

Schema library and signed-bundle producer for the AI Agent Security Platform.

This repo defines the cross-component schema library (request, tool-call, response-prompt envelope) and produces signed OCI bundles containing prompt definitions that are loaded by `agent-sandbox` at startup and consumed by `agent-gateway` and in-house services.

See `docs/superpowers/specs/2026-05-11-lib-agent-prompt-design.md` in the umbrella workspace for the full design.

## Layout

- `schemas/` — canonical JSON Schema source (single source of truth)
- `prompts/` — actual prompt JSON files (per-capability)
- `go/`, `python/`, `rust/`, `java/`, `dotnet/`, `javascript/`, `typescript/` — per-language consumer libraries
- `tools/bundle-builder/` — Go CLI that assembles signed OCI bundles
- `tools/verifier/` — reference verifier (Go)
- `tests/` — schema tests, conformance, reproducibility, e2e
- `docs/` — authoring guides, verification, release process

## Quickstart

```
make schemas-validate   # validate the schema library
make bundle-example     # build the example bundle locally
make verify-example     # verify the locally built bundle
```

## License

Apache 2.0. See `LICENSE`.
