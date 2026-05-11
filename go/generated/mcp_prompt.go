// Vendored copy of the Model Context Protocol Prompt type. Synced from MCP spec; update
// only as part of a major version bump.
type MCPPrompt struct {
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
