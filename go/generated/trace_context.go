// Request-scope identifiers propagated through every cross-component call.
type TraceContext struct {
	// The platform-internal request UUID. Used as the AAD scope for pii-tokenizer.        
	RequestUUID                                                                    string  `json:"request_uuid"`
	// W3C Trace Context header value.                                                     
	Traceparent                                                                    *string `json:"traceparent,omitempty"`
	Tracestate                                                                     *string `json:"tracestate,omitempty"`
}
