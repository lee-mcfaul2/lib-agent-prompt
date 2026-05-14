package loader

import (
	"path/filepath"
	"testing"
)

func TestLoadServices(t *testing.T) {
	root := filepath.Join("testdata", "good-bundle")
	svc, err := LoadServices(root)
	if err != nil {
		t.Fatalf("LoadServices: %v", err)
	}
	if len(svc) != 2 {
		t.Fatalf("services len = %d, want 2", len(svc))
	}
	byMCP := map[string]Service{}
	for _, s := range svc {
		byMCP[s.MCP] = s
	}
	kb, ok := byMCP["kb"]
	if !ok {
		t.Fatal("missing kb")
	}
	if len(kb.Tools) != 2 {
		t.Errorf("kb tools = %d, want 2", len(kb.Tools))
	}
	for _, tool := range kb.Tools {
		if tool.RequestDigest == "" || tool.ResponseDigest == "" {
			t.Errorf("kb.%s: missing digest", tool.Name)
		}
	}

	audit := byMCP["audit_db"]
	if len(audit.Tools) != 1 {
		t.Errorf("audit_db tools = %d, want 1", len(audit.Tools))
	}
	if audit.Tools[0].Name != "search" {
		t.Errorf("audit_db tool name = %q", audit.Tools[0].Name)
	}
}

func TestLoadServicesOrphan(t *testing.T) {
	root := filepath.Join("testdata", "orphan-request")
	if _, err := LoadServices(root); err == nil {
		t.Fatal("expected error for request without paired response")
	}
}

func TestLoadServicesEmpty(t *testing.T) {
	root := filepath.Join("testdata", "empty")
	if _, err := LoadServices(root); err == nil {
		t.Fatal("expected error for empty bundle")
	}
}

func TestDigestDeterministic(t *testing.T) {
	root := filepath.Join("testdata", "good-bundle")
	a, _ := LoadServices(root)
	b, _ := LoadServices(root)
	if a[0].Tools[0].RequestDigest != b[0].Tools[0].RequestDigest {
		t.Errorf("digest not deterministic: %q vs %q", a[0].Tools[0].RequestDigest, b[0].Tools[0].RequestDigest)
	}
}
