package search

import (
	"strconv"
	"strings"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

// Fn turns one search value into a predicate bound to a column. Return
// (nil, nil) to contribute nothing — e.g. an int column given an unparseable
// value drops that value rather than failing the whole search (mirrors
// dataset's `Condition?` null). A non-nil error aborts the whole search.
//
// For column searches, c is the resolved Col; for Named (virtual) searches, c
// is the zero Col and the Fn builds a raw predicate via c.Expr.
type Fn func(c Col, value string) (*metaquery.Filter, error)

// Scope controls whether a column participates in unqualified ("global") terms.
type Scope int

const (
	// ScopeDefault: text columns are global, all others targeted-only.
	ScopeDefault  Scope = iota
	ScopeGlobal         // always searched in unqualified terms
	ScopeTargeted       // only matched via field:value
)

// Field overrides how one result column (keyed by Name in Config.Fields) is
// searched. The zero Field keeps the type-driven defaults, so callers only set
// what differs.
type Field struct {
	Search  Fn       // override the type-default predicate; nil keeps the default
	Scope   Scope    // global vs targeted-only participation
	Aliases []string // extra names usable as the field: target
	Disable bool     // exclude this column from search entirely
}

// Named is a virtual search target not bound to a result column (dataset's
// `search(name) { ... }` — e.g. "daysAgo", "onlyMine"). Search is required.
type Named struct {
	Search Fn   // required; receives the zero Col
	Global bool // participate in unqualified terms (default targeted-only)
}

// Config customizes search compilation. The zero Config yields sensible
// type-driven behavior: text columns are contains-searchable and global; int,
// float and bool columns are exact-match and targeted-only; other kinds are not
// searched. Target and alias resolution is case-insensitive.
type Config struct {
	Fields map[string]Field // by column Name
	Named  map[string]Named // by virtual target name
}

// defaultFn returns the type-driven predicate for a column kind, or nil when
// the kind isn't searchable by default (time/bytes/any — see plan phases).
func defaultFn(col metaquery.Column) Fn {
	switch metaquery.ValueKind(col.GoType) {
	case "text":
		// Plain text → contains; specialized text-like types (enums, uuid,
		// inet, ...) → exact match. Substring search makes no sense for those,
		// and for enums it would wrongly match overlapping labels.
		if isExactText(col.DBType) {
			return EnumFn()
		}
		return func(c Col, v string) (*metaquery.Filter, error) { return c.Contains(v), nil }
	case "int":
		// Operator-aware: bare "5" => =5, ">5"/"<=9"/"5..9" understood.
		return CompareFn(coerceInt)
	case "float":
		return CompareFn(coerceFloat)
	case "bool":
		return func(c Col, v string) (*metaquery.Filter, error) {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, nil
			}
			return c.Eq(b), nil
		}
	default:
		return nil
	}
}

// defaultGlobal reports whether a column participates in unqualified terms by
// default. Only plain text columns do — exact-match types (enums, etc.) are
// targeted-only, since an unqualified free-text term exact-matching an enum is
// rarely intended.
func defaultGlobal(col metaquery.Column) bool {
	return metaquery.ValueKind(col.GoType) == "text" && !isExactText(col.DBType)
}

// plainTextDBTypes are the DB types treated as free-text (contains-searchable).
// Anything else with a string Go type is treated as an exact-match type.
var plainTextDBTypes = map[string]bool{
	"": true, "text": true, "varchar": true, "character varying": true,
	"char": true, "character": true, "bpchar": true, "name": true, "citext": true,
}

func isExactText(dbType string) bool {
	return !plainTextDBTypes[strings.ToLower(dbType)]
}
