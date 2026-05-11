package cmd

import (
	"fmt"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/verifier/internal/verify"
	"github.com/spf13/cobra"
)

var (
	cosignBundle        string
	cosignIdentityRegex string
	cosignOIDCIssuer    string
)

var verifyCmd = &cobra.Command{
	Use:   "verify [bundle.tar.gz]",
	Short: "Verify a packed bundle: content hash + structure + optional cosign signature",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, hash, err := verify.LoadAndHash(args[0])
		if err != nil {
			return err
		}
		if err := verify.VerifyStructure(b); err != nil {
			return err
		}
		if cosignBundle != "" {
			if err := verify.VerifyCosign(verify.CosignOptions{
				BundlePath:               cosignBundle,
				CertificateIdentityRegex: cosignIdentityRegex,
				OIDCIssuer:               cosignOIDCIssuer,
			}, args[0]); err != nil {
				return err
			}
			fmt.Println("cosign signature: OK")
		}
		fmt.Printf("OK: %s\n  digest: %s\n  prompts: %d\n  services: %d\n",
			args[0], hash, len(b.Prompts), len(b.Services))
		return nil
	},
}

func init() {
	verifyCmd.Flags().StringVar(&cosignBundle, "cosign-bundle", "", "Path to a Sigstore bundle (.cosign.bundle.json)")
	verifyCmd.Flags().StringVar(&cosignIdentityRegex, "cosign-identity-regex", "", "Expected certificate identity (regex)")
	verifyCmd.Flags().StringVar(&cosignOIDCIssuer, "cosign-oidc-issuer", "https://token.actions.githubusercontent.com", "Expected OIDC issuer")
}
