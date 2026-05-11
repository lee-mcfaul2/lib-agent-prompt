// Registration of an in-house service. One file per service under
// schemas/service-references/.
type ServiceReference struct {
	Description                                                                               *string `json:"description,omitempty"`
	// Service name as referenced by allowed_responses.service in prompt files.                       
	Name                                                                                      string  `json:"name"`
	// OCI digest of the service's mcp-schema.json artifact. Bundle builder pulls + verifies +        
	// embeds at this digest.                                                                         
	SourceDigest                                                                              string  `json:"source_digest"`
	// Expected SPIFFE URI of the service. Mesh authz keys on this identity.                          
	Spiffe                                                                                    string  `json:"spiffe"`
}
