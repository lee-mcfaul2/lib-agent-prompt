package agentprompt

import (
	"context"
	"os"
	"testing"
)

func TestLoad_OCI_Smoke(t *testing.T) {
	registry := os.Getenv("AGENT_PROMPT_OCI_TEST_REGISTRY")
	if registry == "" {
		t.Skip("set AGENT_PROMPT_OCI_TEST_REGISTRY to run OCI fetch test")
	}
	v := NewVerifier(TrustRelaxed, "", "")
	loader := NewBundleLoader(registry, v)
	_, err := loader.Load(context.Background(), "ai-security/bundles/example:latest")
	if err != nil {
		t.Fatal(err)
	}
}
