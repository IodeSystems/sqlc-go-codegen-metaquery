package endtoend

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mqc "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// authorsRequest builds a plugin.GenerateRequest that mirrors the
// testdata/authors project: one table (authors), four queries.
// The same request is reused by F2 to test per-query and global emission levels.
func authorsRequest() *plugin.GenerateRequest {
	idType := &plugin.Identifier{Name: "bigserial"}
	textType := &plugin.Identifier{Name: "text"}
	authorsIdent := &plugin.Identifier{Name: "authors"}

	authorsTable := &plugin.Table{
		Rel: authorsIdent,
		Columns: []*plugin.Column{
			{Name: "id", Type: idType, NotNull: true, Table: authorsIdent},
			{Name: "name", Type: textType, NotNull: true, Table: authorsIdent},
			{Name: "bio", Type: textType, Table: authorsIdent},
		},
	}

	colID := &plugin.Column{Name: "id", Type: idType, NotNull: true, Table: authorsIdent}
	colName := &plugin.Column{Name: "name", Type: textType, NotNull: true, Table: authorsIdent}
	colBio := &plugin.Column{Name: "bio", Type: textType, Table: authorsIdent}

	return &plugin.GenerateRequest{
		Settings: &plugin.Settings{
			Engine: "postgresql",
		},
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{
					Name:   "public",
					Tables: []*plugin.Table{authorsTable},
				},
			},
		},
		Queries: []*plugin.Query{
			{
				Name:            "CreateAuthor",
				Cmd:             ":one",
				Filename:        "query.sql",
				Text:            "INSERT INTO authors (\n  name, bio\n) VALUES (\n  $1, $2\n)\nRETURNING id, name, bio",
				Params:          []*plugin.Parameter{{Number: 1, Column: colName}, {Number: 2, Column: colBio}},
				Columns:         []*plugin.Column{colID, colName, colBio},
				InsertIntoTable: authorsIdent,
			},
			{
				Name:     "DeleteAuthor",
				Cmd:      ":exec",
				Filename: "query.sql",
				Text:     "DELETE FROM authors\nWHERE id = $1",
				Params:   []*plugin.Parameter{{Number: 1, Column: colID}},
			},
			{
				Name:     "GetAuthor",
				Cmd:      ":one",
				Filename: "query.sql",
				Text:     "SELECT id, name, bio FROM authors\nWHERE id = $1 LIMIT 1",
				Params:   []*plugin.Parameter{{Number: 1, Column: colID}},
				Columns:  []*plugin.Column{colID, colName, colBio},
			},
			{
				Name:     "ListAuthors",
				Cmd:      ":many",
				Filename: "query.sql",
				Text:     "SELECT id, name, bio FROM authors\nORDER BY name",
				Columns:  []*plugin.Column{colID, colName, colBio},
			},
		},
		SqlcVersion: "v1.23.0",
	}
}

// pluginOpts marshals a map to the JSON bytes that opts.Parse expects in
// GenerateRequest.PluginOptions.
func pluginOpts(m map[string]any) []byte {
	b, _ := json.Marshal(m)
	return b
}

// stockFileNames are the files the stock (non-metaquery) codegen produces.
var stockFileNames = []string{"db.go", "models.go", "query.sql.go"}

// findFile returns the file contents for the given name from a GenerateResponse,
// or "" if not found.
func findFile(resp *plugin.GenerateResponse, name string) (string, bool) {
	for _, f := range resp.Files {
		if f.Name == name {
			return string(f.Contents), true
		}
	}
	return "", false
}

// assertStockFilesIdentical checks that the stock files in two responses are
// byte-identical, proving the metaquery plugin does not modify stock output.
func assertStockFilesIdentical(t *testing.T, a, b *plugin.GenerateResponse, labelA, labelB string) {
	t.Helper()
	for _, name := range stockFileNames {
		ga, oka := findFile(a, name)
		gb, okb := findFile(b, name)
		if !oka {
			t.Errorf("%s: missing stock file %s", labelA, name)
			continue
		}
		if !okb {
			t.Errorf("%s: missing stock file %s", labelB, name)
			continue
		}
		if ga != gb {
			// Show first divergence for easier debugging.
			for i := 0; i < len(ga) && i < len(gb); i++ {
				if ga[i] != gb[i] {
					t.Errorf("stock file %s differs between %s and %s at byte %d:\n  %s: %q\n  %s: %q",
						name, labelA, labelB, i,
						labelA, ga[max(0, i-20):min(len(ga), i+20)],
						labelB, gb[max(0, i-20):min(len(gb), i+20)])
					break
				}
			}
			if len(ga) != len(gb) {
				t.Errorf("stock file %s length mismatch: %s=%d %s=%d", name, labelA, len(ga), labelB, len(gb))
			}
		}
	}
}

// loadTestdataStockFiles reads the pinned stock tree at testdata/authors/go.
func loadTestdataStockFiles(t *testing.T) map[string]string {
	t.Helper()
	dir := filepath.Join("testdata", "authors", "go")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading testdata stock tree %s: %v", dir, err)
	}
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		m[e.Name()] = string(b)
	}
	return m
}

// assertStockMatchesTestdata checks that the stock files in resp are
// byte-identical to the pinned testdata/authors/go tree.
func assertStockMatchesTestdata(t *testing.T, resp *plugin.GenerateResponse, stock map[string]string) {
	t.Helper()
	for _, name := range stockFileNames {
		want, ok := stock[name]
		if !ok {
			t.Errorf("testdata stock tree missing %s", name)
			continue
		}
		got, ok := findFile(resp, name)
		if !ok {
			t.Errorf("response missing stock file %s", name)
			continue
		}
		if got != want {
			// Find first divergence for the error message.
			n := len(want)
			if len(got) < n {
				n = len(got)
			}
			for i := 0; i < n; i++ {
				if got[i] != want[i] {
					lo, hi := max(0, i-30), min(len(got), i+30)
					t.Errorf("stock file %s differs from testdata at byte %d:\n  testdata: %q\n  generated: %q\n  (lengths: testdata=%d generated=%d)",
						name, i, want[max(0, i-30):min(len(want), i+30)], got[lo:hi], len(want), len(got))
					break
				}
			}
			if len(got) != len(want) {
				t.Errorf("stock file %s length: testdata=%d generated=%d", name, len(want), len(got))
			}
		}
	}
}

// metaquerySymbols are the symbols that appear in the metaquery file at the
// default (cols) emission level.
var metaquerySymbols = []string{
	"MetaCreateAuthor",
	"MetaDeleteAuthor",
	"MetaGetAuthor",
	"MetaListAuthors",
	"WrapCreateAuthor",
	"WrapDeleteAuthor",
	"WrapGetAuthor",
	"WrapListAuthors",
	"CreateAuthorCols",
	"GetAuthorCols",
	"ListAuthorsCols",
}

// TestGenerate_RegenTestdataStock regenerates the pinned stock tree at
// testdata/authors/go by running Generate with emit_metaquery: off.
// Run manually: go test -run RegenTestdataStock ./internal/endtoend/ -v
func TestGenerate_RegenTestdataStock(t *testing.T) {
	req := authorsRequest()
	req.PluginOptions = pluginOpts(map[string]any{
		"package":        "querytest",
		"sql_package":    "pgx/v5",
		"emit_metaquery": "off",
	})
	resp, err := mqc.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := filepath.Join("testdata", "authors", "go")
	for _, f := range resp.Files {
		if strings.HasSuffix(f.Name, ".metaquery.go") {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, f.Name), f.Contents, 0o644); err != nil {
			t.Fatalf("writing %s: %v", f.Name, err)
		}
		t.Logf("wrote %s (%d bytes)", f.Name, len(f.Contents))
	}
}

// TestGenerate_AdditiveEmission proves F1 scenario 1:
// Given a sqlc project configured with the metaquery plugin (emit_metaquery: cols, the default)
// When sqlc generate runs (i.e. Generate is called)
// Then the stock sqlc output is unchanged (identical to the off run)
// And the plugin produces a new query.sql.metaquery.go file with query symbols
func TestGenerate_AdditiveEmission(t *testing.T) {
	// Run with the plugin off to get the baseline stock output.
	offReq := authorsRequest()
	offReq.PluginOptions = pluginOpts(map[string]any{
		"package":        "querytest",
		"sql_package":    "pgx/v5",
		"emit_metaquery": "off",
	})
	offResp, err := mqc.Generate(context.Background(), offReq)
	if err != nil {
		t.Fatalf("Generate (off) failed: %v", err)
	}

	// Run with the plugin on (default: cols).
	onReq := authorsRequest()
	onReq.PluginOptions = pluginOpts(map[string]any{
		"package":     "querytest",
		"sql_package": "pgx/v5",
	})
	onResp, err := mqc.Generate(context.Background(), onReq)
	if err != nil {
		t.Fatalf("Generate (on) failed: %v", err)
	}

	// Stock files must be identical between off and on runs.
	assertStockFilesIdentical(t, offResp, onResp, "off", "on")

	// Stock files must also be byte-identical to the pinned testdata tree.
	stock := loadTestdataStockFiles(t)
	assertStockMatchesTestdata(t, onResp, stock)

	// The metaquery file must appear in the on run.
	mqFile, ok := findFile(onResp, "query.sql.metaquery.go")
	if !ok {
		names := make([]string, 0, len(onResp.Files))
		for _, f := range onResp.Files {
			names = append(names, f.Name)
		}
		t.Fatalf("expected query.sql.metaquery.go in on response; got files: %v", names)
	}

	// The metaquery file must contain query symbols.
	for _, sym := range metaquerySymbols {
		if !strings.Contains(mqFile, sym) {
			t.Errorf("metaquery file missing symbol %q", sym)
		}
	}
}

// TestGenerate_NoEmissionUnconfigured proves F1 scenario 2 at the Generate
// boundary. The literal "no .metaquery.go when plugin not configured"
// guarantee is a sqlc-level property (the plugin binary is never invoked) and
// cannot be tested here. The correct in-process assertion, per the tech lead's
// correction, is that with emit_metaquery: off the stock files are
// byte-identical to the default run (the plugin only adds, never modifies),
// and the metaquery file (if present) contains no query symbols.
func TestGenerate_NoEmissionUnconfigured(t *testing.T) {
	// Run with the plugin off.
	offReq := authorsRequest()
	offReq.PluginOptions = pluginOpts(map[string]any{
		"package":        "querytest",
		"sql_package":    "pgx/v5",
		"emit_metaquery": "off",
	})
	offResp, err := mqc.Generate(context.Background(), offReq)
	if err != nil {
		t.Fatalf("Generate (off) failed: %v", err)
	}

	// Run with the plugin on (default: cols).
	onReq := authorsRequest()
	onReq.PluginOptions = pluginOpts(map[string]any{
		"package":     "querytest",
		"sql_package": "pgx/v5",
	})
	onResp, err := mqc.Generate(context.Background(), onReq)
	if err != nil {
		t.Fatalf("Generate (on) failed: %v", err)
	}

	// Stock files must be identical — the plugin never modifies them.
	assertStockFilesIdentical(t, offResp, onResp, "off", "on")

	// The metaquery file may exist as a near-empty stub. It must not contain
	// any query symbols.
	mqFile, exists := findFile(offResp, "query.sql.metaquery.go")
	if exists {
		for _, sym := range metaquerySymbols {
			if strings.Contains(mqFile, sym) {
				t.Errorf("emit_metaquery:off should not emit %q; found it in the metaquery file", sym)
			}
		}
	}
}
