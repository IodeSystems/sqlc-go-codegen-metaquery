package main

import (
	"github.com/sqlc-dev/plugin-sdk-go/codegen"

	golang "github.com/iodesystems/sqlc-go-codegen-metaquery/internal"
)

func main() {
	codegen.Run(golang.Generate)
}
