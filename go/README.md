# agentprompt — Go consumer library

```go
import (
    "context"
    agentprompt "github.com/lee-mcfaul2/lib-agent-prompt/go"
)

verifier := agentprompt.NewVerifier(
    agentprompt.TrustProduction,
    `^https://github\.com/lee-mcfaul2/lib-agent-prompt/.*`,
    "https://token.actions.githubusercontent.com",
)
loader := agentprompt.NewBundleLoader("registry.example.io", verifier)
bundle, err := loader.Load(ctx, "ai-security/bundles/example@sha256:...")
prompt, err := bundle.Prompt(uuid)
err = bundle.ValidateIterationOutput(uuid, llmOutputBytes)
```

## Trust policies

- `TrustRelaxed` — content hash only (dev only)
- `TrustStandard` — + cosign
- `TrustProduction` — + SLSA + maintainer GPG tag (default for prod)

## Codegen status

The `generated/` directory was removed because quicktype's Go output places `import` declarations mid-file. The library currently works with `map[string]any` access for forward compatibility. A future improvement: post-process quicktype output to consolidate imports, or migrate to `oapi-codegen` for richer typed bindings.
