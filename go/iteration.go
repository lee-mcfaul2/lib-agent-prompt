package agentprompt

import (
	"encoding/json"
	"fmt"
)

type IterationOutput struct {
	ServiceCalls []struct {
		ID   string         `json:"id"`
		Args map[string]any `json:"args"`
	} `json:"service_calls"`
	Terminate *struct {
		ID      string         `json:"id"`
		Payload map[string]any `json:"payload"`
	} `json:"terminate"`
}

// ValidateIterationOutput checks the LLM's emit against the prompt's allowed_responses.
// Returns an error if the payload doesn't conform; nil on success.
func (b *Bundle) ValidateIterationOutput(promptID string, payload []byte) error {
	prompt, err := b.Prompt(promptID)
	if err != nil {
		return err
	}
	var io IterationOutput
	if err := json.Unmarshal(payload, &io); err != nil {
		return fmt.Errorf("parse iteration output: %w", err)
	}

	hasCalls := len(io.ServiceCalls) > 0
	hasTerminate := io.Terminate != nil
	if hasCalls == hasTerminate {
		return fmt.Errorf("iteration output must populate exactly one of service_calls (non-empty) or terminate")
	}

	allowedCalls := map[string]bool{}
	allowedTerminates := map[string]bool{}
	for _, r := range prompt["allowed_responses"].([]any) {
		rm := r.(map[string]any)
		switch rm["kind"] {
		case "service_call":
			allowedCalls[rm["id"].(string)] = true
		case "terminate":
			allowedTerminates[rm["id"].(string)] = true
		}
	}

	if hasCalls {
		for _, c := range io.ServiceCalls {
			if !allowedCalls[c.ID] {
				return fmt.Errorf("service_call id %s not in prompt's allowed_responses", c.ID)
			}
		}
	}
	if hasTerminate {
		if !allowedTerminates[io.Terminate.ID] {
			return fmt.Errorf("terminate id %s not in prompt's allowed_responses", io.Terminate.ID)
		}
	}

	return nil
}
