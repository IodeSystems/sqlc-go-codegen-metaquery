.PHONY: up down wait plugin generate build seed list list-typed counts demo clean

DSN ?= postgres://demo:demo@localhost:5544/demo?sslmode=disable
SQLC ?= $(shell command -v sqlc 2>/dev/null || echo $${HOME}/go/bin/sqlc)
PLUGIN := ../../bin/sqlc-go-codegen-metaquery
DEMO   := ./bin/demo

up:
	docker compose up -d

down:
	docker compose down -v

wait:
	@printf "waiting for pg..."
	@until docker compose exec -T db pg_isready -U demo -d demo >/dev/null 2>&1; do \
		printf "."; sleep 1; \
	done; echo " ready"

plugin:
	$(MAKE) -C ../.. bin/sqlc-go-codegen-metaquery

generate: plugin
	$(SQLC) generate

build:
	mkdir -p bin
	go build -o $(DEMO) .

seed: build
	$(DEMO) seed --db '$(DSN)'

list: build
	$(DEMO) list --db '$(DSN)' --size 10

list-typed: build
	$(DEMO) list --typed --db '$(DSN)' --size 10

counts: build
	$(DEMO) counts --db '$(DSN)'

# Full walk-through: start db, generate code, seed, and show each subcommand.
demo: up wait generate build seed
	@echo '---- list (untyped, page 1 of 2) ----'
	$(DEMO) list --db '$(DSN)' --size 3 \
		--where '[{"column":"name","op":"ILIKE","value":"%a%"}]' \
		--order '[{"column":"name","dir":"ASC"}]'
	@echo '---- list --typed ----'
	$(DEMO) list --typed --db '$(DSN)' --size 3
	@echo '---- counts (group by author_id) ----'
	$(DEMO) counts --db '$(DSN)' \
		--order '[{"column":"author_id","dir":"ASC"}]'

clean:
	rm -rf bin
