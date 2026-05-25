// Package codegen exposes the metaquery code generator as a
// library — the same function sqlc invokes when it spawns the
// `plugin` binary, but importable from any Go program.
//
// The intended use case: a downstream project's dev tool wires the
// generator into its own binary so `sqlc generate` can be driven
// without installing a separate plugin executable on the host.
// `plugin/main.go` itself is now a thin wrapper around this
// package — keeping the installable plugin binary as the supported
// path while letting embedders skip the install dance entirely.
//
// Example:
//
//	import (
//	    "github.com/sqlc-dev/plugin-sdk-go/codegen"
//	    mqc "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/codegen"
//	)
//
//	func main() { codegen.Run(mqc.Generate) }
package codegen

import (
	"context"

	golang "github.com/iodesystems/sqlc-go-codegen-metaquery/internal"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// Generate is the codegen entry point. Signature matches
// `codegen.Handler` from plugin-sdk-go so it can be passed directly
// to `codegen.Run`.
func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	return golang.Generate(ctx, req)
}
