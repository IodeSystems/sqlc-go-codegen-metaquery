package search

import "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"

// Apply parses query, compiles it against b's output columns + cfg, and ANDs
// the resulting predicate into b's WHERE clause. It returns the normalized
// (auto-escaped) search string for DataSetResponse.searchRendered.
//
// A zero Config gives type-driven defaults (text columns contains-searchable &
// global; int/float/bool exact-match & targeted). Use the same call for the
// partition field — apply partition first, then search; both AND together.
//
// Compilation errors are returned directly; column/op validation errors from
// the underlying builder surface later at b.Build().
func Apply(b *metaquery.Builder, query string, cfg Config) (rendered string, err error) {
	parsed, err := Parse(query)
	if err != nil {
		return "", err
	}
	filter, err := compile(parsed.Terms, b.OutputColumns(), cfg, b.Dialect())
	if err != nil {
		return parsed.Search, err
	}
	if filter != nil {
		b.ApplyFilter(*filter)
	}
	return parsed.Search, nil
}
