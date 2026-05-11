// One entry in a prompt's allowed_responses list that ends the conversation and returns to
// the external caller.
type Terminate struct {
	ID                                                                                         string `json:"id"`
	Kind                                                                                       Kind   `json:"kind"`
	// Relative path within the bundle to the JSON Schema for the terminating response payload.       
	ResponseSchemaRef                                                                          string `json:"response_schema_ref"`
}

type Kind string

const (
	KindTerminate Kind = "terminate"
)
