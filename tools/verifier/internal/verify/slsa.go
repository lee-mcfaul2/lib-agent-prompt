package verify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type SLSAOptions struct {
	AttestationPath   string
	ExpectedBuilderID string
	ExpectedSourceURI string
}

type slsaPredicate struct {
	BuildDefinition struct {
		BuildType            string         `json:"buildType"`
		ExternalParameters   map[string]any `json:"externalParameters"`
		ResolvedDependencies []struct {
			URI    string            `json:"uri"`
			Digest map[string]string `json:"digest"`
		} `json:"resolvedDependencies"`
	} `json:"buildDefinition"`
	RunDetails struct {
		Builder struct {
			ID string `json:"id"`
		} `json:"builder"`
	} `json:"runDetails"`
}

func VerifySLSA(opts SLSAOptions) error {
	data, err := os.ReadFile(opts.AttestationPath)
	if err != nil {
		return fmt.Errorf("read attestation: %w", err)
	}

	type dsse struct {
		PayloadType string `json:"payloadType"`
		Payload     string `json:"payload"`
	}
	var env dsse
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("parse DSSE envelope: %w", err)
	}
	if !strings.Contains(env.PayloadType, "in-toto") {
		return fmt.Errorf("unexpected payloadType: %s", env.PayloadType)
	}
	payload, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	var statement struct {
		Predicate slsaPredicate `json:"predicate"`
	}
	if err := json.Unmarshal(payload, &statement); err != nil {
		return fmt.Errorf("parse statement: %w", err)
	}

	if opts.ExpectedBuilderID != "" && statement.Predicate.RunDetails.Builder.ID != opts.ExpectedBuilderID {
		return fmt.Errorf("builder mismatch: got %s want %s", statement.Predicate.RunDetails.Builder.ID, opts.ExpectedBuilderID)
	}

	if opts.ExpectedSourceURI != "" {
		found := false
		for _, dep := range statement.Predicate.BuildDefinition.ResolvedDependencies {
			if dep.URI == opts.ExpectedSourceURI {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("source URI %s not in resolvedDependencies", opts.ExpectedSourceURI)
		}
	}

	return nil
}
