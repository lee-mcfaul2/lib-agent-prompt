module github.com/lee-mcfaul2/lib-agent-prompt/tools/verifier

go 1.24.4

require (
	github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.8.1
)

replace github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify => ../../pkg/verify

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/text v0.14.0 // indirect
)
