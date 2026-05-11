// Standard error response shape used by every component.
type ErrorEnvelope struct {
	// Stable identifier for the error class (e.g., SCHEMA_VALIDATION_FAILED, SIGNATURE_INVALID).        
	ErrorType                                                                                    string  `json:"error_type"`
	// Human-readable description. Must NOT contain plaintext PII.                                       
	Message                                                                                      string  `json:"message"`
	// Whether the caller should retry the same request.                                                 
	Retriable                                                                                    bool    `json:"retriable"`
	TraceID                                                                                      *string `json:"trace_id,omitempty"`
}
