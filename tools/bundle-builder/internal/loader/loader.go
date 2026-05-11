package loader

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// LoadJSONFile reads and parses a JSON file into a generic map.
func LoadJSONFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}

// LoadAllJSON walks a directory and loads every *.json file, keyed by relative path.
func LoadAllJSON(root string) (map[string]map[string]any, error) {
	out := make(map[string]map[string]any)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		doc, err := LoadJSONFile(path)
		if err != nil {
			return err
		}
		out[rel] = doc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
