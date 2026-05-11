// A single agent capability: nested MCP Prompt + authorized services + allowed response
// shapes + cost caps.
type Prompt struct {
	AllowedResponses                                                                      []AllowedResponse `json:"allowed_responses"`
	CostCaps                                                                              CostCaps          `json:"cost_caps"`
	ID                                                                                    string            `json:"id"`
	Prompt                                                                                MCPPromptVendored `json:"prompt"`
	SchemaVersion                                                                         string            `json:"schema_version"`
	// Authorization scope: SPIFFE identities agent-gateway may invoke under this prompt's                  
	// request_uuid.                                                                                        
	Services                                                                              []Service         `json:"services"`
}

// One entry in a prompt's allowed_responses list that represents a call to a registered
// service.
//
// One entry in a prompt's allowed_responses list that ends the conversation and returns to
// the external caller.
type AllowedResponse struct {
	// Stable identifier within this prompt's allowed_responses (e.g.,                                 
	// 'call-postgres-get-users').                                                                     
	ID                                                                                         string  `json:"id"`
	Kind                                                                                       Kind    `json:"kind"`
	// Relative path within the bundle to the JSON Schema for the request payload (e.g.,               
	// 'service-schemas/postgresql-service.json#/tools/get_users/inputSchema').                        
	RequestSchemaRef                                                                           *string `json:"request_schema_ref,omitempty"`
	// Name of a service declared in this prompt's `services` list.                                    
	Service                                                                                    *string `json:"service,omitempty"`
	// Relative path within the bundle to the JSON Schema for the terminating response payload.        
	ResponseSchemaRef                                                                          *string `json:"response_schema_ref,omitempty"`
}

type CostCaps struct {
	MaxCostUsd     float64 `json:"max_cost_usd"`
	MaxIterations  int64   `json:"max_iterations"`
	MaxWallclockMS int64   `json:"max_wallclock_ms"`
}

// Vendored copy of the Model Context Protocol Prompt type. Synced from MCP spec; update
// only as part of a major version bump.
type MCPPromptVendored struct {
	// Input parameters the prompt accepts.                           
	Arguments                                              []Argument `json:"arguments,omitempty"`
	// Human-readable description of what this prompt does.           
	Description                                            string     `json:"description"`
	// Templated message sequence the LLM evaluates.                  
	Messages                                               []Message  `json:"messages,omitempty"`
	// Unique identifier for the prompt within the bundle.            
	Name                                                   string     `json:"name"`
}

type Argument struct {
	Default     interface{} `json:"default"`
	Description *string     `json:"description,omitempty"`
	Name        string      `json:"name"`
	Required    *bool       `json:"required,omitempty"`
}

type Message struct {
	Content Content `json:"content"`
	Role    Role    `json:"role"`
}

type Content struct {
	Text     *string `json:"text,omitempty"`
	Type     Type    `json:"type"`
	Data     *string `json:"data,omitempty"`
	MIMEType *string `json:"mimeType,omitempty"`
}

type Service struct {
	Name         string `json:"name"`
	SchemaDigest string `json:"schema_digest"`
	Spiffe       string `json:"spiffe"`
}

type Kind string

const (
	ServiceCall Kind = "service_call"
	Terminate   Kind = "terminate"
)

type Type string

const (
	Image Type = "image"
	Text  Type = "text"
)

type Role string

const (
	Assistant Role = "assistant"
	System    Role = "system"
	User      Role = "user"
)
