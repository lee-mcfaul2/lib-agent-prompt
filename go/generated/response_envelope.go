// Wrapper agent-gateway applies to service responses before delivering them to the sandbox.
// Frames the payload as data rather than instructions so the LLM does not interpret
// retrieved content as commands.
type ResponseEnvelope struct {
	// Version of the response-envelope schema. Matches the schema-library SemVer.                        
	EnvelopeVersion                                                                string                 `json:"envelope_version"`
	// Display markers the gateway inserts in the prompt context.                                         
	Framing                                                                        *Framing               `json:"framing,omitempty"`
	// Validated at runtime against the service's response schema (from                                   
	// service-schemas/<name>.json).                                                                      
	Payload                                                                        map[string]interface{} `json:"payload"`
	// id from the iteration-output.service_calls[] entry that triggered this call.                       
	RequestID                                                                      string                 `json:"request_id"`
	// Name of the service that produced this payload.                                                    
	Service                                                                        string                 `json:"service"`
}

// Display markers the gateway inserts in the prompt context.
type Framing struct {
	Close *string `json:"close,omitempty"`
	Open  *string `json:"open,omitempty"`
}
