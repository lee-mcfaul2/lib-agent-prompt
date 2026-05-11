// Enumeration of PII types plus regex/rule backstops used by agent-gateway's scrubber as
// the second line of detection after schema annotations.
type PiiTypes struct {
	Patterns                                                           map[string]interface{} `json:"patterns"`
	// Backstop ruleset version; bumps with the schema-library version.                       
	Version                                                            string                 `json:"version"`
}
