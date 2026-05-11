package agentprompt

import (
	verify "github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify"
)

type Verifier struct {
	Policy              TrustPolicy
	CosignIdentityRegex string
	OIDCIssuer          string
	GPGKeyring          string
	SLSABuilderID       string
	SLSASourceURI       string
}

func NewVerifier(policy TrustPolicy, cosignIdentityRegex, oidcIssuer string) *Verifier {
	return &Verifier{
		Policy:              policy,
		CosignIdentityRegex: cosignIdentityRegex,
		OIDCIssuer:          oidcIssuer,
	}
}

// Verify executes the verification chain matching the configured policy.
// bundlePath is the tarball; cosignBundle and slsaAttestation are paths to side artifacts (may be "" if policy doesn't require).
func (v *Verifier) Verify(bundlePath, cosignBundle, slsaAttestation string) error {
	b, _, err := verify.LoadAndHash(bundlePath)
	if err != nil {
		return err
	}
	if err := verify.VerifyStructure(b); err != nil {
		return err
	}
	if v.Policy >= TrustStandard && cosignBundle != "" {
		if err := verify.VerifyCosign(verify.CosignOptions{
			BundlePath:               cosignBundle,
			CertificateIdentityRegex: v.CosignIdentityRegex,
			OIDCIssuer:               v.OIDCIssuer,
		}, bundlePath); err != nil {
			return err
		}
	}
	if v.Policy >= TrustProduction && slsaAttestation != "" {
		if err := verify.VerifySLSA(verify.SLSAOptions{
			AttestationPath:   slsaAttestation,
			ExpectedBuilderID: v.SLSABuilderID,
			ExpectedSourceURI: v.SLSASourceURI,
		}); err != nil {
			return err
		}
	}
	return nil
}
