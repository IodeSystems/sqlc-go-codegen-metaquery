.PHONY: build test

BIN_NAME = sqlc-go-codegen-metaquery

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
