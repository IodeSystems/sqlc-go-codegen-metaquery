.PHONY: build test release antlr

BIN_NAME = sqlc-go-codegen-metaquery
# Overridable for release testing against a different repo.
RELEASE_REPO ?= IodeSystems/sqlc-go-codegen-metaquery

build:
	go build ./...

test: bin/$(BIN_NAME).wasm
	go test ./...

all: bin/$(BIN_NAME) bin/$(BIN_NAME).wasm

bin/$(BIN_NAME): bin go.mod go.sum $(wildcard **/*.go)
	cd plugin && go build -o ../bin/$(BIN_NAME) ./main.go

bin/$(BIN_NAME).wasm: bin/$(BIN_NAME)
	cd plugin && GOOS=wasip1 GOARCH=wasm go build -o ../bin/$(BIN_NAME).wasm main.go

bin:
	mkdir -p bin

# Regenerate the search Go parser from the shared grammar (ANTLR 4.13.2).
# Run after editing metaquery/search/grammar/*.g4. Requires Java + ANTLR.
antlr:
	./scripts/antlr.sh

# Publish a tagged release. Usage:
#   make release VERSION=v0.2.0
#
# Builds the WASM, creates/updates the GitHub release with the asset,
# and rewrites the release block in README.md. Does NOT auto-commit the
# README update — review + commit yourself.
release:
	@test -n "$(VERSION)" || { echo "VERSION is required, e.g. make release VERSION=v0.2.0" >&2; exit 1; }
	REPO=$(RELEASE_REPO) ./scripts/release.sh $(VERSION)
