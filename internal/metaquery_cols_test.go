package golang

import (
	"strings"
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/internal/opts"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// A query that aliases two columns off the same underlying name — `SELECT p.name AS program_name,
// r.name AS region_name` — which is where the accessors used to break.
func aliasedQuery() Query {
	col := func(orig string) *plugin.Column {
		return &plugin.Column{Name: orig, OriginalName: orig, NotNull: true,
			Type: &plugin.Identifier{Name: "text"}}
	}
	return Query{
		Cmd:        ":many",
		MethodName: "ListThings",
		Ret: QueryValue{
			Struct: &Struct{
				Name: "ListThingsRow",
				Fields: []Field{
					{Name: "ProgramName", DBName: "program_name", Type: "string", Column: col("name")},
					{Name: "RegionName", DBName: "region_name", Type: "string", Column: col("name")},
					{Name: "ThingID", DBName: "thing_id", Type: "int64",
						Column: &plugin.Column{Name: "thing_id", OriginalName: "thing_id", NotNull: true,
							Type: &plugin.Identifier{Name: "bigint"}}},
				},
			},
		},
	}
}

// TestRenderMetaCols_UsesOutputAliasNotOriginalName pins the name the accessors emit.
//
// It must be the query's OUTPUT name. The wrapper wraps the query as a subquery, so only output
// names are in scope, and Builder.lookupColumn matches Name OR OriginalName and returns the FIRST
// hit. Emitting OriginalName made the accessors ambiguous as well as wrong: both columns below
// rendered as "name", so RegionName silently resolved to program_name — the wrong column, with no
// error anywhere. A caller sorting by RegionName got program_name.
func TestRenderMetaCols_UsesOutputAliasNotOriginalName(t *testing.T) {
	var sb strings.Builder
	renderMetaCols(&sb, aliasedQuery(), &opts.Options{})
	got := sb.String()

	// Compared before gofmt, so no alignment padding.
	for _, want := range []string{
		`ProgramName: metaquery.NewTextCol("program_name")`,
		`RegionName: metaquery.NewTextCol("region_name")`,
		`ThingId: metaquery.NewIntCol("thing_id")`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("emitted Cols missing %s\n--- got ---\n%s", want, got)
		}
	}
	// The pre-fix output. Two accessors collapsing to the same string is the ambiguity itself.
	if strings.Contains(got, `NewTextCol("name")`) {
		t.Errorf("emitted the ORIGINAL name for an aliased column — accessors would resolve to the "+
			"wrong column\n--- got ---\n%s", got)
	}
}
