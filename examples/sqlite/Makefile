.PHONY: plugin generate build init seed list list-typed counts demo clean

DSN ?= file:demo.db?_pragma=foreign_keys(1)
SQLC ?= $(shell command -v sqlc 2>/dev/null || echo $${HOME}/go/bin/sqlc)
PLUGIN := ../../bin/sqlc-go-codegen-metaquery
DEMO   := ./bin/demo

plugin:
	$(MAKE) -C ../.. bin/sqlc-go-codegen-metaquery

generate: plugin
	$(SQLC) generate

build:
	mkdir -p bin
	go build -o $(DEMO) .

init: build
	rm -f demo.db
	$(DEMO) init --db '$(DSN)'

seed: init
	$(DEMO) seed --db '$(DSN)'

list: build
	$(DEMO) list --db '$(DSN)' --size 10

list-typed: build
	$(DEMO) list --typed --db '$(DSN)' --size 10

counts: build
	$(DEMO) counts --db '$(DSN)'

# Full walk-through: regenerate code, init schema, seed, and show each subcommand.
demo: generate seed
	@echo '---- list (untyped, with filter+order) ----'
	$(DEMO) list --db '$(DSN)' --size 3 \
		--where '[{"column":"name","op":"LIKE","value":"%a%"}]' \
		--order '[{"column":"name","dir":"ASC"}]'
	@echo '---- list --typed ----'
	$(DEMO) list --typed --db '$(DSN)' --size 3
	@echo '---- counts (group by author_id) ----'
	$(DEMO) counts --db '$(DSN)' \
		--order '[{"column":"author_id","dir":"ASC"}]'

clean:
	rm -rf bin demo.db
