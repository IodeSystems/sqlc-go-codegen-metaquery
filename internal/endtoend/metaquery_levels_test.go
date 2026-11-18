package endtoend

import (
	"context"
	"strings"
	"testing"

	mqc "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/codegen"
)

// TestGenerate_ThreeSymbolsAtDefault proves F2 scenario 1:
// Given a query at the default emission level cols
// When codegen runs
// Then the generated file contains Meta<Name>
// And it contains Wrap<Name>
// And it contains <Name>Cols
func TestGenerate_ThreeSymbolsAtDefault(t *testing.T) {
	req := authorsRequest()
	req.PluginOptions = pluginOpts(map[string]any{
		"package":     "querytest",
		"sql_package": "pgx/v5",
	})
	resp, err := mqc.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate (default) failed: %v", err)
	}

	mqFile, ok := findFile(resp, "query.sql.metaquery.go")
	if !ok {
		names := make([]string, 0, len(resp.Files))
		for _, f := range resp.Files {
			names = append(names, f.Name)
		}
		t.Fatalf("expected query.sql.metaquery.go; got files: %v", names)
	}

	// All four queries get Meta and Wrap.
	for _, name := range []string{"CreateAuthor", "DeleteAuthor", "GetAuthor", "ListAuthors"} {
		if !strings.Contains(mqFile, "Meta"+name) {
			t.Errorf("default level: missing Meta%s", name)
		}
		if !strings.Contains(mqFile, "Wrap"+name) {
			t.Errorf("default level: missing Wrap%s", name)
		}
	}
	// Queries with return columns also get Cols. DeleteAuthor is :exec (no
	// return columns) so it correctly has no Cols struct.
	for _, name := range []string{"CreateAuthor", "GetAuthor", "ListAuthors"} {
		if !strings.Contains(mqFile, name+"Cols") {
			t.Errorf("default level: missing %sCols", name)
		}
	}
}

// TestGenerate_PerQueryOff proves F2 scenario 2:
// Given a query annotated with the per-query override comment `-- metaquery: off`
// When codegen runs
// Then the plugin emits no metaquery symbols for that query
// And it leaves the other queries unchanged
func TestGenerate_PerQueryOff(t *testing.T) {
	req := authorsRequest()
	// Annotate GetAuthor with `-- metaquery: off`.
	req.Queries[2].Comments = []string{"metaquery: off"}

	req.PluginOptions = pluginOpts(map[string]any{
		"package":     "querytest",
		"sql_package": "pgx/v5",
	})
	resp, err := mqc.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate (per-query off) failed: %v", err)
	}

	mqFile, ok := findFile(resp, "query.sql.metaquery.go")
	if !ok {
		names := make([]string, 0, len(resp.Files))
		for _, f := range resp.Files {
			names = append(names, f.Name)
		}
		t.Fatalf("expected query.sql.metaquery.go; got files: %v", names)
	}

	// GetAuthor must have NO metaquery symbols.
	for _, sym := range []string{"MetaGetAuthor", "WrapGetAuthor", "GetAuthorCols"} {
		if strings.Contains(mqFile, sym) {
			t.Errorf("per-query off: GetAuthor should have no %s, but it is present", sym)
		}
	}

	// Sibling queries must still have Meta and Wrap.
	for _, name := range []string{"CreateAuthor", "DeleteAuthor", "ListAuthors"} {
		for _, prefix := range []string{"Meta", "Wrap"} {
			sym := prefix + name
			if !strings.Contains(mqFile, sym) {
				t.Errorf("per-query off: sibling %s should have %s", name, sym)
			}
		}
	}
	// Siblings with return columns must have Cols.
	for _, name := range []string{"CreateAuthor", "ListAuthors"} {
		if !strings.Contains(mqFile, name+"Cols") {
			t.Errorf("per-query off: sibling %s should have %sCols", name, name)
		}
	}
}

// TestGenerate_GlobalMeta proves F2 scenario 3:
// Given the emit_metaquery option set to meta
// When codegen runs
// Then the generated file contains only the query metadata
// And it contains no builder wrapper or typed columns
func TestGenerate_GlobalMeta(t *testing.T) {
	req := authorsRequest()
	req.PluginOptions = pluginOpts(map[string]any{
		"package":        "querytest",
		"sql_package":    "pgx/v5",
		"emit_metaquery": "meta",
	})
	resp, err := mqc.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate (meta) failed: %v", err)
	}

	mqFile, ok := findFile(resp, "query.sql.metaquery.go")
	if !ok {
		names := make([]string, 0, len(resp.Files))
		for _, f := range resp.Files {
			names = append(names, f.Name)
		}
		t.Fatalf("expected query.sql.metaquery.go; got files: %v", names)
	}

	for _, name := range []string{"CreateAuthor", "DeleteAuthor", "GetAuthor", "ListAuthors"} {
		// Meta symbols must be present.
		if !strings.Contains(mqFile, "Meta"+name) {
			t.Errorf("meta level: missing Meta%s", name)
		}
		// Wrap and Cols must be absent.
		if strings.Contains(mqFile, "Wrap"+name) {
			t.Errorf("meta level: should not contain Wrap%s", name)
		}
		if strings.Contains(mqFile, name+"Cols") {
			t.Errorf("meta level: should not contain %sCols", name)
		}
	}
}
