package validator

import (
	"path/filepath"
	"testing"
)

func TestValidateGoodTree(t *testing.T) {
	root := filepath.Join("..", "loader", "testdata", "good-bundle")
	if err := ValidateSchemaTree(root); err != nil {
		t.Fatalf("ValidateSchemaTree: %v", err)
	}
}

func TestValidateOrphanRequestRejected(t *testing.T) {
	root := filepath.Join("..", "loader", "testdata", "orphan-request")
	if err := ValidateSchemaTree(root); err == nil {
		t.Fatal("expected error for orphan request")
	}
}
