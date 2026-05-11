// The LLM's per-iteration output envelope. Exactly one of service_calls (non-empty) or
// terminate is populated; the other is null.
type IterationOutput struct {
	ServiceCalls []interface{} `json:"service_calls"`
	Terminate    *Terminate    `json:"terminate"`
}

type Terminate struct {
	// Must match an allowed_responses[].id where kind==terminate.                            
	ID                                                                 string                 `json:"id"`
	// Validated at runtime against the referenced response_schema_ref.                       
	Payload                                                            map[string]interface{} `json:"payload"`
}
