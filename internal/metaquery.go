package golang

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/internal/opts"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// renderMetaQueries produces per-query metadata declarations for a single
// source file. Emission is gated by the emit_metaquery option, overridable
// per query with a `-- metaquery: <level>` comment.
//
// Ladder (additive): off < meta < wrap < cols.
//
// options is threaded in via a closure (see gen.go) so StructName uses the
// same initialisms/rename rules as the rest of the codegen.
func renderMetaQueries(queries []Query, source, engine string, options *opts.Options) string {
	defaultLevel := levelValue(options.EmitMetaquery)
	dialect := dialectForEngine(engine)
	var sb strings.Builder
	for _, q := range queries {
		if q.SourceName != source {
			continue
		}
		level := queryEmitLevel(q, defaultLevel)
		if level == emitOff {
			continue
		}
		renderMetaQuery(&sb, q, dialect)
		if level >= emitWrap && wrappableCmd(q.Cmd) {
			if hasTopLevelOrderBy(q.SQL) {
				// sqlc swallows process-plugin stderr, so the warning lives in
				// the generated code where it is always visible and reviewable.
				// log.Printf is a belt for `sqlc generate --verbose`.
				renderOrderByWarning(&sb, q.MethodName)
				log.Printf("metaquery: query %q ends in a top-level ORDER BY. "+
					"Wrapping re-applies ordering at runtime; the doubled, nested sort "+
					"can defeat index use (see benchmark/README.md). Drop ORDER BY from "+
					"the query and order via .ApplyOrder(...) instead.", q.MethodName)
			}
			renderMetaWrapper(&sb, q)
		}
		if level >= emitCols {
			renderMetaCols(&sb, q, options)
		}
		// Emit aggregation helpers from -- metaquery:agg annotations.
		aggs := parseAggAnnotations(q)
		for _, agg := range aggs {
			renderAggHelper(&sb, q, agg, options)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Emission levels form a numeric ladder so level >= N checks stay readable.
const (
	emitOff  = 0
	emitMeta = 1
	emitWrap = 2
	emitCols = 3
)

func levelValue(s string) int {
	switch s {
	case "off":
		return emitOff
	case "meta":
		return emitMeta
	case "wrap":
		return emitWrap
	case "cols", "":
		return emitCols
	}
	return emitCols
}

// queryEmitLevel scans q.Comments for `-- metaquery: <level>` and returns the
// resolved level (per-query override, else the provided default).
func queryEmitLevel(q Query, fallback int) int {
	for _, c := range q.Comments {
		s := strings.TrimSpace(c)
		s = strings.TrimPrefix(s, "--")
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "metaquery:") {
			continue
		}
		val := strings.TrimSpace(strings.TrimPrefix(s, "metaquery:"))
		return levelValue(val)
	}
	return fallback
}

// wrappableCmd reports whether a query Cmd makes sense to wrap in a Builder.
// :copyfrom has no placeholders; :batch* builds a separate batch object.
func wrappableCmd(cmd string) bool {
	switch cmd {
	case ":one", ":many", ":exec", ":execrows", ":execlastid", ":execresult":
		return true
	}
	return false
}

// hasTopLevelOrderBy reports whether sql contains an ORDER BY clause at
// parenthesis depth 0 — i.e. the outermost statement's own ordering. Because
// metaquery wraps the whole query as a CTE and re-applies ordering at runtime,
// a top-level ORDER BY produces a doubled nested sort (see benchmark/README.md).
// String literals, quoted identifiers, and comments are skipped so an
// "order by" inside them or inside a subquery doesn't trip the check.
func hasTopLevelOrderBy(sql string) bool {
	s := sql
	depth := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'', '"': // string literal / quoted identifier — skip to its close
			q := c
			i++
			for i < len(s) {
				if s[i] == q {
					if i+1 < len(s) && s[i+1] == q { // doubled quote = escaped
						i++
					} else {
						break
					}
				}
				i++
			}
		case '-': // line comment
			if i+1 < len(s) && s[i+1] == '-' {
				for i < len(s) && s[i] != '\n' {
					i++
				}
			}
		case '/': // block comment
			if i+1 < len(s) && s[i+1] == '*' {
				i += 2
				for i+1 < len(s) && !(s[i] == '*' && s[i+1] == '/') {
					i++
				}
				i++
			}
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 && (c == 'o' || c == 'O') && matchOrderBy(s, i) {
				return true
			}
		}
	}
	return false
}

// matchOrderBy reports whether s has the keyword sequence "order" <ws> "by" at
// position i, bounded by non-identifier characters on both sides.
func matchOrderBy(s string, i int) bool {
	if i > 0 && isIdentByte(s[i-1]) {
		return false
	}
	j := i
	if !hasWordFold(s, j, "order") {
		return false
	}
	j += len("order")
	if j >= len(s) || !isSpaceByte(s[j]) {
		return false
	}
	for j < len(s) && isSpaceByte(s[j]) {
		j++
	}
	if !hasWordFold(s, j, "by") {
		return false
	}
	j += len("by")
	return j >= len(s) || !isIdentByte(s[j])
}

func hasWordFold(s string, i int, word string) bool {
	if i+len(word) > len(s) {
		return false
	}
	return strings.EqualFold(s[i:i+len(word)], word)
}

func isIdentByte(b byte) bool {
	return b == '_' || (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}

// renderOrderByWarning emits a comment above a Wrap func whose source query
// ends in a top-level ORDER BY (the doubled-sort footgun; see Performance in
// the README).
func renderOrderByWarning(sb *strings.Builder, method string) {
	fmt.Fprintf(sb, "// WARNING: %s ends in a top-level ORDER BY. Wrapping re-applies\n", method)
	sb.WriteString("// ordering at runtime, so .ApplyOrder(...) produces a doubled, nested sort\n")
	sb.WriteString("// that can defeat index use. Drop ORDER BY from the query and order via\n")
	sb.WriteString("// .ApplyOrder(...) instead. See benchmark/README.md.\n")
}

// renderMetaWrapper emits `func Wrap<Name>(...) *metaquery.Builder { ... }`
// with a typed parameter list derived from q.Arg.
func renderMetaWrapper(sb *strings.Builder, q Query) {
	params, callArgs := wrapperPieces(q)
	fmt.Fprintf(sb, "// Wrap%s returns a metaquery.Builder over Meta%s, pre-bound with typed arguments.\n", q.MethodName, q.MethodName)
	fmt.Fprintf(sb, "func Wrap%s(%s) *metaquery.Builder {\n", q.MethodName, params)
	if callArgs == "" {
		fmt.Fprintf(sb, "\treturn metaquery.Wrap(&Meta%s)\n", q.MethodName)
	} else {
		fmt.Fprintf(sb, "\treturn metaquery.Wrap(&Meta%s, %s)\n", q.MethodName, callArgs)
	}
	sb.WriteString("}\n")
}

// renderMetaCols emits a `var <Method>Cols = struct { ... }{ ... }` so callers
// get compile-time-safe column names and typed filter builders:
//
//	b.ApplyFilter(db.ListAuthorsCols.Name.ILike("%foo%"))
//	b.ApplyOrder(db.ListAuthorsCols.CreatedAt.Desc())
func renderMetaCols(sb *strings.Builder, q Query, options *opts.Options) {
	cols := metaColumns(q)
	if len(cols) == 0 {
		return
	}
	fmt.Fprintf(sb, "// %sCols gives typed, name-safe access to %s's output columns.\n", q.MethodName, q.MethodName)
	fmt.Fprintf(sb, "var %sCols = struct {\n", q.MethodName)
	fieldNames := make([]string, len(cols))
	seen := map[string]int{}
	for i, c := range cols {
		fn := colFieldName(c, options)
		if n, dup := seen[fn]; dup {
			seen[fn] = n + 1
			fn = fmt.Sprintf("%s_%d", fn, n+1)
		} else {
			seen[fn] = 1
		}
		fieldNames[i] = fn
		fmt.Fprintf(sb, "\t%s metaquery.%s\n", fn, columnKindFor(c.GoType))
	}
	sb.WriteString("}{\n")
	for i, c := range cols {
		fmt.Fprintf(sb, "\t%s: metaquery.%s(%q),\n", fieldNames[i], columnKindCtor(c.GoType), c.OriginalName)
	}
	sb.WriteString("}\n")
}

// columnKindFor maps a sqlc-emitted GoType to a metaquery col-kind type name.
// Unknowns fall back to AnyCol — the escape hatch with untyped ops.
func columnKindFor(goType string) string {
	switch goType {
	case "string", "pgtype.Text", "pgtype.Varchar", "pgtype.Name", "pgtype.BPChar":
		return "TextCol"
	case "int", "int16", "int32", "int64", "uint", "uint16", "uint32", "uint64",
		"pgtype.Int2", "pgtype.Int4", "pgtype.Int8":
		return "IntCol"
	case "float32", "float64", "pgtype.Float4", "pgtype.Float8":
		return "FloatCol"
	case "bool", "pgtype.Bool":
		return "BoolCol"
	case "time.Time", "pgtype.Timestamp", "pgtype.Timestamptz", "pgtype.Date", "pgtype.Time":
		return "TimeCol"
	case "[]byte":
		return "BytesCol"
	default:
		return "AnyCol"
	}
}

func columnKindCtor(goType string) string {
	return "New" + columnKindFor(goType)
}

// colFieldName is the Go field name used for a column inside the generated
// Cols struct. Uses sqlc's existing StructName so the field names align
// with the sqlc-generated model structs (ID, CreatedAt, AuthorID, ...).
func colFieldName(c metaCol, options *opts.Options) string {
	name := c.Name
	if name == "" {
		name = c.OriginalName
	}
	return StructName(name, options)
}

// wrapperPieces computes the Go parameter list for the wrapper function and
// the comma-separated expression list forwarded into metaquery.Wrap.
//
// Shapes:
//   - no args:               params="",                 callArgs=""
//   - single scalar arg:     params="id int64",         callArgs="id"
//   - multi-param (flat):    params="name string, bio pgtype.Text",
//                            callArgs="name, bio"
//   - struct param (emit):   params="arg CreateAuthorParams",
//                            callArgs="arg.Name, arg.Bio"
func wrapperPieces(q Query) (string, string) {
	if q.Arg.isEmpty() {
		return "", ""
	}
	pairs := q.Arg.Pairs()
	paramParts := make([]string, len(pairs))
	for i, p := range pairs {
		paramParts[i] = p.Name + " " + p.Type
	}
	params := strings.Join(paramParts, ", ")

	var callParts []string
	if q.Arg.Struct != nil {
		for _, f := range q.Arg.Struct.Fields {
			callParts = append(callParts, q.Arg.VariableForField(f))
		}
	} else {
		callParts = append(callParts, q.Arg.Name)
	}
	return params, strings.Join(callParts, ", ")
}

// dialectForEngine maps a sqlc engine string to the metaquery Dialect token
// rendered into the generated Meta<Name>. Empty string -> Postgres (zero
// value, the implicit default for prior consumers).
func dialectForEngine(engine string) string {
	switch engine {
	case "sqlite":
		return "metaquery.DialectSQLite"
	default:
		return ""
	}
}

func renderMetaQuery(sb *strings.Builder, q Query, dialect string) {
	fmt.Fprintf(sb, "var Meta%s = metaquery.Query{\n", q.MethodName)
	fmt.Fprintf(sb, "\tName: %q,\n", q.MethodName)
	fmt.Fprintf(sb, "\tCmd: %q,\n", q.Cmd)
	fmt.Fprintf(sb, "\tSource: %q,\n", q.SourceName)
	if dialect != "" {
		fmt.Fprintf(sb, "\tDialect: %s,\n", dialect)
	}
	sb.WriteString("\tSQL: `")
	// backtick-escape: sqlc's SQL text may contain backticks in extension
	// function bodies; replace them with concatenation
	sb.WriteString(strings.ReplaceAll(q.SQL, "`", "`+\"`\"+`"))
	sb.WriteString("`,\n")

	cols := metaColumns(q)
	if len(cols) > 0 {
		sb.WriteString("\tColumns: []metaquery.Column{\n")
		for _, c := range cols {
			renderMetaColumn(sb, c)
		}
		sb.WriteString("\t},\n")
	}

	args := metaArgs(q)
	if len(args) > 0 {
		sb.WriteString("\tArgs: []metaquery.Arg{\n")
		for _, a := range args {
			renderMetaArg(sb, a)
		}
		sb.WriteString("\t},\n")
	}

	if q.Table != nil && (q.Table.Catalog != "" || q.Table.Schema != "" || q.Table.Name != "") {
		sb.WriteString("\tTable: &metaquery.Table{")
		if q.Table.Catalog != "" {
			fmt.Fprintf(sb, "Catalog: %q, ", q.Table.Catalog)
		}
		if q.Table.Schema != "" {
			fmt.Fprintf(sb, "Schema: %q, ", q.Table.Schema)
		}
		fmt.Fprintf(sb, "Name: %q", q.Table.Name)
		sb.WriteString("},\n")
	}

	sb.WriteString("}\n")
}

type metaCol struct {
	Name, OriginalName, GoType, DBType string
	Table, TableAlias                  string
	NotNull, IsArray                   bool
}

type metaArg struct {
	Position                int
	Name, GoType, DBType    string
	NotNull, IsArray, Slice bool
}

func metaColumns(q Query) []metaCol {
	if q.Ret.isEmpty() {
		return nil
	}
	if q.Ret.Struct != nil {
		out := make([]metaCol, 0, len(q.Ret.Struct.Fields))
		for _, f := range q.Ret.Struct.Fields {
			out = append(out, fieldToMetaCol(f))
		}
		return out
	}
	return []metaCol{scalarToMetaCol(q.Ret)}
}

func fieldToMetaCol(f Field) metaCol {
	c := metaCol{
		Name:         f.DBName,
		OriginalName: f.DBName,
		GoType:       f.Type,
	}
	if f.Column != nil {
		if f.Column.OriginalName != "" {
			c.OriginalName = f.Column.OriginalName
		}
		c.DBType = columnDBType(f.Column)
		c.NotNull = f.Column.NotNull
		c.IsArray = f.Column.IsArray
		if f.Column.Table != nil {
			c.Table = f.Column.Table.Name
		}
		c.TableAlias = f.Column.TableAlias
	}
	return c
}

func scalarToMetaCol(v QueryValue) metaCol {
	c := metaCol{
		Name:         v.DBName,
		OriginalName: v.DBName,
		GoType:       v.Typ,
	}
	if v.Column != nil {
		if v.Column.OriginalName != "" {
			c.OriginalName = v.Column.OriginalName
		}
		c.DBType = columnDBType(v.Column)
		c.NotNull = v.Column.NotNull
		c.IsArray = v.Column.IsArray
		if v.Column.Table != nil {
			c.Table = v.Column.Table.Name
		}
		c.TableAlias = v.Column.TableAlias
	}
	return c
}

func metaArgs(q Query) []metaArg {
	if q.Arg.isEmpty() {
		return nil
	}
	if q.Arg.Struct != nil {
		out := make([]metaArg, 0, len(q.Arg.Struct.Fields))
		for i, f := range q.Arg.Struct.Fields {
			a := metaArg{
				Position: i + 1,
				Name:     f.DBName,
				GoType:   f.Type,
			}
			if f.Column != nil {
				a.DBType = columnDBType(f.Column)
				a.NotNull = f.Column.NotNull
				a.IsArray = f.Column.IsArray
				a.Slice = f.Column.IsSqlcSlice
			}
			out = append(out, a)
		}
		return out
	}
	a := metaArg{
		Position: 1,
		Name:     q.Arg.DBName,
		GoType:   q.Arg.Typ,
	}
	if q.Arg.Column != nil {
		a.DBType = columnDBType(q.Arg.Column)
		a.NotNull = q.Arg.Column.NotNull
		a.IsArray = q.Arg.Column.IsArray
		a.Slice = q.Arg.Column.IsSqlcSlice
	}
	return []metaArg{a}
}

func columnDBType(c *plugin.Column) string {
	if c == nil || c.Type == nil {
		return ""
	}
	return c.Type.Name
}

func renderMetaColumn(sb *strings.Builder, c metaCol) {
	sb.WriteString("\t\t{")
	fmt.Fprintf(sb, "Name: %q, OriginalName: %q, GoType: %q", c.Name, c.OriginalName, c.GoType)
	if c.DBType != "" {
		fmt.Fprintf(sb, ", DBType: %q", c.DBType)
	}
	if c.NotNull {
		sb.WriteString(", NotNull: true")
	}
	if c.IsArray {
		sb.WriteString(", IsArray: true")
	}
	if c.Table != "" {
		fmt.Fprintf(sb, ", Table: %q", c.Table)
	}
	if c.TableAlias != "" {
		fmt.Fprintf(sb, ", TableAlias: %q", c.TableAlias)
	}
	sb.WriteString("},\n")
}

func renderMetaArg(sb *strings.Builder, a metaArg) {
	sb.WriteString("\t\t{")
	fmt.Fprintf(sb, "Position: %s", strconv.Itoa(a.Position))
	if a.Name != "" {
		fmt.Fprintf(sb, ", Name: %q", a.Name)
	}
	fmt.Fprintf(sb, ", GoType: %q", a.GoType)
	if a.DBType != "" {
		fmt.Fprintf(sb, ", DBType: %q", a.DBType)
	}
	if a.NotNull {
		sb.WriteString(", NotNull: true")
	}
	if a.IsArray {
		sb.WriteString(", IsArray: true")
	}
	if a.Slice {
		sb.WriteString(", IsSqlcSlice: true")
	}
	sb.WriteString("},\n")
}

// ── Aggregation annotations ──────────────────────────────────────────────────

// aggAnnotation holds a parsed `-- metaquery:agg <Name> <ops...>` directive.
type aggAnnotation struct {
	Name string   // e.g. "CountedGrouped"
	Ops  []aggOp  // ordered operations
}

type aggOp struct {
	Kind  string // "group_by_expr", "group_by", "count", "sum", "max", "min", "avg"
	Alias string // output column alias
	Col   string // source column (for sum/max/min/avg)
	Expr  string // raw SQL expr (for group_by_expr)
	Type  string // Go type (for group_by_expr)
}

// parseAggAnnotations extracts all `-- metaquery:agg` directives from q.Comments.
func parseAggAnnotations(q Query) []aggAnnotation {
	var out []aggAnnotation
	for _, c := range q.Comments {
		s := strings.TrimSpace(c)
		s = strings.TrimPrefix(s, "--")
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "metaquery:agg ") {
			continue
		}
		s = strings.TrimPrefix(s, "metaquery:agg ")
		agg := parseOneAgg(s)
		if agg.Name != "" {
			out = append(out, agg)
		}
	}
	return out
}

// parseOneAgg parses: <Name> op(args) op(args) ...
// Example: CountedGrouped group_by_expr(group_value, "context_json->>?", string) count(entry_count) max(max_value, value)
func parseOneAgg(s string) aggAnnotation {
	// First token is the name.
	name, rest, ok := strings.Cut(s, " ")
	if !ok {
		return aggAnnotation{Name: s}
	}
	agg := aggAnnotation{Name: name}

	// Parse op(args) tokens. Parens may contain quoted strings with commas.
	for rest != "" {
		rest = strings.TrimSpace(rest)
		if rest == "" {
			break
		}
		paren := strings.IndexByte(rest, '(')
		if paren < 0 {
			break
		}
		kind := rest[:paren]
		close := findMatchingParen(rest, paren)
		if close < 0 {
			break
		}
		inner := rest[paren+1 : close]
		rest = rest[close+1:]

		op := aggOp{Kind: kind}
		args := splitArgs(inner)
		switch kind {
		case "group_by_expr":
			// group_by_expr(alias, expr, goType)
			if len(args) >= 3 {
				op.Alias = args[0]
				op.Expr = args[1]
				op.Type = args[2]
			}
		case "group_by":
			// group_by(col)
			if len(args) >= 1 {
				op.Col = args[0]
				op.Alias = args[0]
			}
		case "count":
			// count(alias)
			if len(args) >= 1 {
				op.Alias = args[0]
			}
		case "sum", "max", "min", "avg":
			// op(alias, col)
			if len(args) >= 2 {
				op.Alias = args[0]
				op.Col = args[1]
			}
		}
		agg.Ops = append(agg.Ops, op)
	}
	return agg
}

func findMatchingParen(s string, open int) int {
	depth := 0
	inQuote := false
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
		case '(':
			if !inQuote {
				depth++
			}
		case ')':
			if !inQuote {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return -1
}

// splitArgs splits on commas, respecting quoted strings. Trims whitespace and quotes.
func splitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, ch := range s {
		switch {
		case ch == '"':
			inQuote = !inQuote
		case ch == ',' && !inQuote:
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 {
		out = append(out, strings.TrimSpace(cur.String()))
	}
	return out
}

// renderAggHelper emits a typed struct + wrapper function for an aggregation annotation.
func renderAggHelper(sb *strings.Builder, q Query, agg aggAnnotation, options *opts.Options) {
	structName := q.MethodName + agg.Name + "Row"

	// Determine struct fields from ops.
	type field struct {
		Name, GoType, DBTag string
	}
	var fields []field
	var hasExprParam bool

	for _, op := range agg.Ops {
		switch op.Kind {
		case "group_by_expr":
			fields = append(fields, field{
				Name:   StructName(op.Alias, options),
				GoType: goTypeForAggType(op.Type),
				DBTag:  op.Alias,
			})
			hasExprParam = true
		case "group_by":
			fields = append(fields, field{
				Name:   StructName(op.Alias, options),
				GoType: goTypeForColumn(q, op.Col),
				DBTag:  op.Alias,
			})
		case "count":
			fields = append(fields, field{
				Name:   StructName(op.Alias, options),
				GoType: "int64",
				DBTag:  op.Alias,
			})
		case "sum":
			fields = append(fields, field{
				Name:   StructName(op.Alias, options),
				GoType: sumTypeFor(goTypeForColumn(q, op.Col)),
				DBTag:  op.Alias,
			})
		case "max", "min":
			fields = append(fields, field{
				Name:   StructName(op.Alias, options),
				GoType: goTypeForColumn(q, op.Col),
				DBTag:  op.Alias,
			})
		case "avg":
			fields = append(fields, field{
				Name:   StructName(op.Alias, options),
				GoType: "float64",
				DBTag:  op.Alias,
			})
		}
	}

	// Emit struct.
	fmt.Fprintf(sb, "// %s is the scan target for %s aggregation on %s.\n", structName, agg.Name, q.MethodName)
	fmt.Fprintf(sb, "type %s struct {\n", structName)
	for _, f := range fields {
		fmt.Fprintf(sb, "\t%s %s `db:\"%s\" json:\"%s\"`\n", f.Name, f.GoType, f.DBTag, f.DBTag)
	}
	sb.WriteString("}\n\n")

	// Emit wrapper function.
	// Params: same as base Wrap + optional expr params (e.g. groupKey string).
	basePairs := q.Arg.Pairs()
	var paramParts []string
	for _, p := range basePairs {
		paramParts = append(paramParts, p.Name+" "+p.Type)
	}
	if hasExprParam {
		paramParts = append(paramParts, "groupKey string")
	}
	params := strings.Join(paramParts, ", ")

	wrapName := "Wrap" + q.MethodName + agg.Name
	fmt.Fprintf(sb, "// %s returns a Builder pre-configured with %s aggregation over %s.\n", wrapName, agg.Name, q.MethodName)
	fmt.Fprintf(sb, "func %s(%s) *metaquery.Builder {\n", wrapName, params)

	// Call the base Wrap.
	var baseCallArgs []string
	if q.Arg.Struct != nil {
		for _, f := range q.Arg.Struct.Fields {
			baseCallArgs = append(baseCallArgs, q.Arg.VariableForField(f))
		}
	} else if !q.Arg.isEmpty() {
		baseCallArgs = append(baseCallArgs, q.Arg.Name)
	}
	if len(basePairs) > 1 || (len(basePairs) == 1 && q.Arg.Struct != nil) {
		// Struct arg: pass the struct
		fmt.Fprintf(sb, "\tb := Wrap%s(%s)\n", q.MethodName, basePairs[0].Name)
	} else if len(basePairs) == 1 {
		fmt.Fprintf(sb, "\tb := Wrap%s(%s)\n", q.MethodName, basePairs[0].Name)
	} else {
		fmt.Fprintf(sb, "\tb := Wrap%s()\n", q.MethodName)
	}

	// Chain ops.
	for _, op := range agg.Ops {
		switch op.Kind {
		case "group_by_expr":
			fmt.Fprintf(sb, "\tb.GroupByExpr(%q, %q, %q, groupKey)\n", op.Alias, op.Expr, op.Type)
		case "group_by":
			fmt.Fprintf(sb, "\tb.GroupBy(%q)\n", op.Col)
		case "count":
			fmt.Fprintf(sb, "\tb.Count(%q)\n", op.Alias)
		case "sum":
			fmt.Fprintf(sb, "\tb.Sum(%q, %q)\n", op.Alias, op.Col)
		case "max":
			fmt.Fprintf(sb, "\tb.Max(%q, %q)\n", op.Alias, op.Col)
		case "min":
			fmt.Fprintf(sb, "\tb.Min(%q, %q)\n", op.Alias, op.Col)
		case "avg":
			fmt.Fprintf(sb, "\tb.Avg(%q, %q)\n", op.Alias, op.Col)
		}
	}
	sb.WriteString("\treturn b\n}\n")
}

func goTypeForAggType(t string) string {
	switch t {
	case "string":
		return "string"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "float64":
		return "float64"
	case "bool":
		return "bool"
	default:
		return "string"
	}
}

func goTypeForColumn(q Query, col string) string {
	if q.Ret.Struct != nil {
		for _, f := range q.Ret.Struct.Fields {
			if f.DBName == col {
				return f.Type
			}
		}
	}
	return "int32" // fallback
}

func sumTypeFor(goType string) string {
	switch goType {
	case "int32", "int16":
		return "int64"
	case "int64":
		return "int64"
	case "float32", "float64":
		return "float64"
	default:
		return "int64"
	}
}
