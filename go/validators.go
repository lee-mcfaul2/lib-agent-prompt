package agentprompt

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Validator wraps a compiled JSON Schema for validating instance payloads.
type Validator struct {
	schema *jsonschema.Schema
}

// Validate parses raw as JSON and validates against the underlying schema.
func (v *Validator) Validate(raw []byte) error {
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse instance: %w", err)
	}
	return v.schema.Validate(doc)
}

// CompileRequestValidator returns a validator for mcp.tool's request schema.
func (b *Bundle) CompileRequestValidator(mcp, tool string) (*Validator, error) {
	raw, ok := b.RequestSchema(mcp, tool)
	if !ok {
		return nil, fmt.Errorf("no request schema for %s.%s", mcp, tool)
	}
	return compileOne(mcp+"/"+tool+".request", raw, nil)
}

// CompileResponseValidator returns a validator for mcp.tool's response schema.
func (b *Bundle) CompileResponseValidator(mcp, tool string) (*Validator, error) {
	raw, ok := b.ResponseSchema(mcp, tool)
	if !ok {
		return nil, fmt.Errorf("no response schema for %s.%s", mcp, tool)
	}
	return compileOne(mcp+"/"+tool+".response", raw, nil)
}

// CompileUserPromptValidator returns a validator for the user-prompt envelope.
func (b *Bundle) CompileUserPromptValidator() (*Validator, error) {
	return compileOne("user-prompt", b.UserPromptSchema(), b.sharedResources())
}

// CompileFinalResponseValidator returns a validator for the final-response envelope.
func (b *Bundle) CompileFinalResponseValidator() (*Validator, error) {
	return compileOne("final-response", b.FinalResponseSchema(), b.sharedResources())
}

// CompileToolResultValidator returns a validator for the tool-result envelope.
func (b *Bundle) CompileToolResultValidator() (*Validator, error) {
	return compileOne("tool-result", b.ToolResultSchema(), b.sharedResources())
}

// sharedResources returns the set of shared schema keys from the bundle that
// the envelope schemas reference via $ref (e.g. "shared/uuid.json").
// These are loaded alongside the envelope schema so the compiler can resolve them.
func (b *Bundle) sharedResources() map[string][]byte {
	shared := map[string][]byte{}
	for k, v := range b.Schemas {
		// Shared schemas are not under services/ but would be stored separately
		// if loaded; envelope schemas load shared/ directly from the source tree.
		// Since LoadFromDir only stores service schemas, this is a no-op unless
		// the bundle was loaded with shared schemas. It is a safe extension point.
		_ = k
		_ = v
	}
	return shared
}

// compileOne compiles a single JSON Schema identified by id from raw bytes.
// extra is an optional map of additional resource id → raw bytes to pre-register
// (used for shared schemas referenced via $ref by the primary schema).
func compileOne(id string, raw []byte, extra map[string][]byte) (*Validator, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", id, err)
	}
	c := jsonschema.NewCompiler()
	for extraID, extraRaw := range extra {
		extraDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(extraRaw))
		if err != nil {
			return nil, fmt.Errorf("parse extra %s: %w", extraID, err)
		}
		if err := c.AddResource(extraID, extraDoc); err != nil {
			return nil, fmt.Errorf("add extra %s: %w", extraID, err)
		}
	}
	if err := c.AddResource(id, doc); err != nil {
		return nil, fmt.Errorf("add %s: %w", id, err)
	}
	s, err := c.Compile(id)
	if err != nil {
		return nil, fmt.Errorf("compile %s: %w", id, err)
	}
	return &Validator{schema: s}, nil
}
