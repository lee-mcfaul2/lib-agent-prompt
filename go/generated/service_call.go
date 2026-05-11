// One entry in a prompt's allowed_responses list that represents a call to a registered
// service.
type ServiceCall struct {
	// Stable identifier within this prompt's allowed_responses (e.g.,                         
	// 'call-postgres-get-users').                                                             
	ID                                                                                  string `json:"id"`
	Kind                                                                                Kind   `json:"kind"`
	// Relative path within the bundle to the JSON Schema for the request payload (e.g.,       
	// 'service-schemas/postgresql-service.json#/tools/get_users/inputSchema').                
	RequestSchemaRef                                                                    string `json:"request_schema_ref"`
	// Name of a service declared in this prompt's `services` list.                            
	Service                                                                             string `json:"service"`
}

type Kind string

const (
	KindServiceCall Kind = "service_call"
)
