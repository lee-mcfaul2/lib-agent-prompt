package verify

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifySLSA_BuilderMismatch(t *testing.T) {
	dir := t.TempDir()
	statement := map[string]any{
		"predicate": map[string]any{
			"runDetails": map[string]any{
				"builder": map[string]any{"id": "https://example.com/bad-builder"},
			},
			"buildDefinition": map[string]any{
				"resolvedDependencies": []map[string]any{
					{"uri": "git+https://github.com/lee-mcfaul2/lib-agent-prompt"},
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(statement)
	envelope := map[string]any{
		"payloadType": "application/vnd.in-toto+json",
		"payload":     base64.StdEncoding.EncodeToString(payloadBytes),
	}
	envBytes, _ := json.Marshal(envelope)
	path := filepath.Join(dir, "att.jsonl")
	os.WriteFile(path, envBytes, 0o644)

	err := VerifySLSA(SLSAOptions{
		AttestationPath:   path,
		ExpectedBuilderID: "https://github.com/lee-mcfaul2/lib-agent-prompt/.github/workflows/build-and-sign.yml@refs/tags/v1.0.0",
	})
	if err == nil {
		t.Fatal("expected builder mismatch error")
	}
}

func TestVerifySLSA_AllExpectedMatch(t *testing.T) {
	dir := t.TempDir()
	statement := map[string]any{
		"predicate": map[string]any{
			"runDetails": map[string]any{
				"builder": map[string]any{"id": "https://github.com/lee-mcfaul2/lib-agent-prompt/.github/workflows/build-and-sign.yml@refs/tags/v1.0.0"},
			},
			"buildDefinition": map[string]any{
				"resolvedDependencies": []map[string]any{
					{"uri": "git+https://github.com/lee-mcfaul2/lib-agent-prompt"},
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(statement)
	envelope := map[string]any{
		"payloadType": "application/vnd.in-toto+json",
		"payload":     base64.StdEncoding.EncodeToString(payloadBytes),
	}
	envBytes, _ := json.Marshal(envelope)
	path := filepath.Join(dir, "att.jsonl")
	os.WriteFile(path, envBytes, 0o644)

	err := VerifySLSA(SLSAOptions{
		AttestationPath:   path,
		ExpectedBuilderID: "https://github.com/lee-mcfaul2/lib-agent-prompt/.github/workflows/build-and-sign.yml@refs/tags/v1.0.0",
		ExpectedSourceURI: "git+https://github.com/lee-mcfaul2/lib-agent-prompt",
	})
	if err != nil {
		t.Fatalf("expected pass, got: %v", err)
	}
}
