// Command validate-fixture compiles a JSON Schema and validates an instance file.
// Used by local-smoke.sh to assert a fixture conforms to a bundle's schema.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func main() {
	schemaPath := flag.String("schema", "", "path to JSON Schema")
	instancePath := flag.String("instance", "", "path to instance file")
	flag.Parse()
	if *schemaPath == "" || *instancePath == "" {
		fmt.Fprintln(os.Stderr, "both --schema and --instance required")
		os.Exit(2)
	}

	c := jsonschema.NewCompiler()

	// Pre-load shared schemas so relative $ref values resolve.
	// Determine the schemas/ root from the schema path.
	schemasRoot := filepath.Dir(*schemaPath)
	for filepath.Base(schemasRoot) != "schemas" && schemasRoot != "/" {
		schemasRoot = filepath.Dir(schemasRoot)
	}
	sharedDir := filepath.Join(schemasRoot, "shared")
	if entries, err := os.ReadDir(sharedDir); err == nil {
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			full := filepath.Join(sharedDir, e.Name())
			raw, err := os.ReadFile(full)
			if err != nil {
				continue
			}
			doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
			if err != nil {
				continue
			}
			// Register under relative path used in $ref values.
			_ = c.AddResource("shared/"+e.Name(), doc)
			// Also register under the absolute $id URI.
			var schemaDoc map[string]any
			if json.Unmarshal(raw, &schemaDoc) == nil {
				if id, ok := schemaDoc["$id"].(string); ok && id != "" {
					rawDoc, _ := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
					_ = c.AddResource(id, rawDoc)
				}
			}
		}
	}

	schemaRaw, err := os.ReadFile(*schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read schema: %v\n", err)
		os.Exit(2)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaRaw))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse schema: %v\n", err)
		os.Exit(2)
	}
	if err := c.AddResource(*schemaPath, doc); err != nil {
		fmt.Fprintf(os.Stderr, "add schema: %v\n", err)
		os.Exit(2)
	}
	s, err := c.Compile(*schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compile schema: %v\n", err)
		os.Exit(2)
	}

	instRaw, err := os.ReadFile(*instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read instance: %v\n", err)
		os.Exit(2)
	}
	var inst any
	if err := json.Unmarshal(instRaw, &inst); err != nil {
		fmt.Fprintf(os.Stderr, "parse instance: %v\n", err)
		os.Exit(2)
	}
	if err := s.Validate(inst); err != nil {
		fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("ok")
}
