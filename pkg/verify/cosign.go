package verify

import (
	"bytes"
	"fmt"
	"os/exec"
)

type CosignOptions struct {
	BundlePath               string // path to the *.cosign.bundle.json (sigstore bundle format)
	CertificateIdentityRegex string
	OIDCIssuer               string
}

// VerifyCosign shells out to the cosign CLI to verify a sigstore bundle against an artifact.
// Using the CLI rather than sigstore-go directly insulates this code from frequent API churn
// in the sigstore-go library and matches the approach used by the .NET consumer library.
func VerifyCosign(opts CosignOptions, artifactPath string) error {
	if opts.BundlePath == "" {
		return fmt.Errorf("cosign bundle path required")
	}
	if opts.CertificateIdentityRegex == "" {
		return fmt.Errorf("CertificateIdentityRegex required")
	}
	if opts.OIDCIssuer == "" {
		opts.OIDCIssuer = "https://token.actions.githubusercontent.com"
	}

	cmd := exec.Command("cosign",
		"verify-blob",
		"--bundle", opts.BundlePath,
		"--certificate-identity-regexp", opts.CertificateIdentityRegex,
		"--certificate-oidc-issuer", opts.OIDCIssuer,
		artifactPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cosign verify-blob: %w: %s", err, stderr.String())
	}
	return nil
}
