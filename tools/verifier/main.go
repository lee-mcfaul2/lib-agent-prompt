package main

import (
	"fmt"
	"os"

	"github.com/lee-mcfaul2/lib-agent-prompt/tools/verifier/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "verify failed:", err)
		os.Exit(2)
	}
}
