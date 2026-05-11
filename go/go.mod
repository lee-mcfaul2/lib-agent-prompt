module github.com/lee-mcfaul2/lib-agent-prompt/go

go 1.24.4

require github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify v0.0.0-00010101000000-000000000000

require (
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	oras.land/oras-go/v2 v2.5.0 // indirect
)

replace github.com/lee-mcfaul2/lib-agent-prompt/pkg/verify => ../pkg/verify
