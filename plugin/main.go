// Standalone plugin binary. sqlc spawns this via the `process`
// plugin protocol; the actual codegen lives in
// `metaquery/codegen` so embedders can wire the same function into
// their own binaries without going through PATH lookup.
package main

import (
	"github.com/sqlc-dev/plugin-sdk-go/codegen"

	mqc "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/codegen"
)

func main() {
	codegen.Run(mqc.Generate)
}
