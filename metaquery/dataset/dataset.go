// Package dataset turns a single DataSetRequest (search + partition + ordering
// + pagination) into a fully-shaped metaquery builder, so any metaquery select
// becomes a searchable/sortable/paginated "DataSet" in one call.
//
// It is adapter-agnostic: Shape mutates a *metaquery.Builder (no DB), and Run
// takes a scan function so callers wire in their adapter (mqpgx.Scan /
// mqsqlite.Scan) without this package depending on pgx or database/sql.
//
//	b := db.WrapListUsers(arg)            // generated builder over a sqlc query
//	res, err := dataset.Run(ctx, b, req, cfg,
//	    func(ctx context.Context, b *metaquery.Builder) (*metaquery.TypedResult[UserView], error) {
//	        return mqpgx.Scan[UserView](ctx, pool, b)
//	    })
//
// Search/partition are compiled by metaquery/search against the builder's
// output columns; cfg embeds search.Config, so the zero value already searches
// text columns and exact-matches typed columns by target.
package dataset

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/search"
)

const (
	defaultPageSize = 25
	maxPageSize     = 200
)

// Order is one sort directive (mirrors redline's DataSetOrder).
type Order struct {
	Field string `json:"field"`
	Order string `json:"order" enum:"ASC,DESC" doc:"Sort direction."`
}

// Request mirrors redline's DataSetRequest. Page is 0-indexed.
type Request struct {
	Page       int     `json:"page,omitempty"`
	PageSize   int     `json:"pageSize,omitempty"`
	Ordering   []Order `json:"ordering,omitempty"`
	Search     string  `json:"search,omitempty"`
	Partition  string  `json:"partition,omitempty"`
	ShowCounts bool    `json:"showCounts,omitempty"`
}

// Count is the response count envelope (redline DataSetResponse.count).
type Count struct {
	InQuery     int64 `json:"inQuery"`
	InPartition int64 `json:"inPartition"`
}

// Rendered carries the normalized (auto-escaped) search/partition strings, for
// DataSetResponse.searchRendered so the UI can show what actually ran.
type Rendered struct {
	Search    string `json:"search,omitempty"`
	Partition string `json:"partition,omitempty"`
}

// Response mirrors redline's DataSetResponse[T]. It is intentionally lean so it
// can be serialized straight to clients: it does NOT embed metaquery.Meta,
// which carries compiled filter exprs and bound argument values (see Builder
// .Meta() / the scan TypedResult if you need column/filter introspection).
type Response[T any] struct {
	Data     []T      `json:"data"`
	Count    *Count   `json:"count,omitempty"`
	Rendered Rendered `json:"rendered"`
}

// Config drives shaping. The embedded search.Config customizes search/partition
// (Fields, Named); the zero value is sensible. Orderable, when non-empty,
// restricts which fields clients may sort by (OrderBy already whitelists
// against output columns, so this is an additional policy layer).
//
// Searchable / Targetable mirror redline's per-field search config. When either
// is non-empty they form a HARD allowlist: columns in Searchable are global
// free-text + targetable; columns in Targetable are targeted-only (field:value,
// like redline's search(global=false)); every other column is NOT searchable by
// any path. This matches redline, where a field with no search config is not
// searchable. When both are empty, the type-driven default applies (all text
// columns global). Explicit per-column entries in search.Config.Fields always
// win over the allowlist. Names are matched case-insensitively.
type Config struct {
	search.Config
	Searchable      []string // global free-text + targetable; empty+Targetable empty = type default
	Targetable      []string // targeted-only (field:value), not matched by bare terms
	Orderable       []string // allowed sort fields; empty = any output column
	DefaultPageSize int      // 0 -> 25
	MaxPageSize     int      // 0 -> 200
}

// effectiveSearch folds the Searchable/Targetable allowlists into a
// search.Config, without mutating the caller's Config. A non-empty allowlist is
// hard: unlisted columns are disabled (neither global nor targetable). Columns
// with an explicit Fields entry are left untouched (explicit wins).
func (c Config) effectiveSearch(cols []metaquery.Column) search.Config {
	if len(c.Searchable) == 0 && len(c.Targetable) == 0 {
		return c.Config
	}
	global := lowerSet(c.Searchable)
	targeted := lowerSet(c.Targetable)
	fields := make(map[string]search.Field, len(cols))
	maps.Copy(fields, c.Config.Fields)
	for _, col := range cols {
		if _, explicit := fields[col.Name]; explicit {
			continue
		}
		switch key := strings.ToLower(col.Name); {
		case global[key]:
			fields[col.Name] = search.Field{Scope: search.ScopeGlobal}
		case targeted[key]:
			fields[col.Name] = search.Field{Scope: search.ScopeTargeted}
		default:
			fields[col.Name] = search.Field{Disable: true}
		}
	}
	return search.Config{Fields: fields, Named: c.Config.Named}
}

func lowerSet(xs []string) map[string]bool {
	if len(xs) == 0 {
		return nil
	}
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[strings.ToLower(x)] = true
	}
	return m
}

// lowerCanon maps lowercased names to their original (canonical) spelling.
func lowerCanon(xs []string) map[string]string {
	if len(xs) == 0 {
		return nil
	}
	m := make(map[string]string, len(xs))
	for _, x := range xs {
		m[strings.ToLower(x)] = x
	}
	return m
}

func (c Config) defPageSize() int {
	if c.DefaultPageSize > 0 {
		return c.DefaultPageSize
	}
	return defaultPageSize
}

func (c Config) maxPgSize() int {
	if c.MaxPageSize > 0 {
		return c.MaxPageSize
	}
	return maxPageSize
}

// Shape applies req to b: partition first, then search (both AND into WHERE),
// then whitelisted ordering, then clamped pagination (with total when
// ShowCounts). It returns the normalized search/partition strings. No DB.
func Shape(b *metaquery.Builder, req Request, cfg Config) (Rendered, error) {
	var r Rendered
	scfg := cfg.effectiveSearch(b.OutputColumns())

	if s := strings.TrimSpace(req.Partition); s != "" {
		rp, err := search.Apply(b, s, scfg)
		if err != nil {
			return r, fmt.Errorf("dataset: partition: %w", err)
		}
		r.Partition = rp
	}
	if s := strings.TrimSpace(req.Search); s != "" {
		rs, err := search.Apply(b, s, scfg)
		if err != nil {
			return r, fmt.Errorf("dataset: search: %w", err)
		}
		r.Search = rs
	}

	// Case-insensitive Orderable allowlist (like the search allowlists). The map
	// resolves a client field to the developer's canonical column name, so
	// OrderBy (which whitelists case-sensitively against columns) gets the right
	// casing. Empty Orderable = any field, passed through for OrderBy to vet.
	canon := lowerCanon(cfg.Orderable)
	for _, o := range req.Ordering {
		field := o.Field
		if canon != nil {
			c, ok := canon[strings.ToLower(o.Field)]
			if !ok {
				continue
			}
			field = c
		}
		dir := metaquery.Asc
		if strings.EqualFold(o.Order, "DESC") {
			dir = metaquery.Desc
		}
		b.OrderBy(field, dir)
	}

	size := req.PageSize
	if size <= 0 {
		size = cfg.defPageSize()
	}
	if m := cfg.maxPgSize(); size > m {
		size = m
	}
	page := max(req.Page, 0)
	b.ApplyPagination(metaquery.PageRequest{Page: page, Size: size, Total: req.ShowCounts})

	return r, nil
}

// ScanFunc executes a shaped builder and returns typed rows + Meta. Satisfied
// by a closure over mqpgx.Scan[T] / mqsqlite.Scan[T].
type ScanFunc[T any] func(context.Context, *metaquery.Builder) (*metaquery.TypedResult[T], error)

// Run shapes b from req and scans it, returning the DataSet response envelope.
// When req.ShowCounts, InQuery is the total after partition+search.
//
// InPartition (rows after partition only) currently mirrors InQuery; a
// partition-only count is a second pass over a fresh builder — see
// RunWithPartitionCount and the plan's phase 4.
func Run[T any](ctx context.Context, b *metaquery.Builder, req Request, cfg Config, scan ScanFunc[T]) (Response[T], error) {
	rendered, err := Shape(b, req, cfg)
	if err != nil {
		return Response[T]{}, err
	}
	res, err := scan(ctx, b)
	if err != nil {
		return Response[T]{}, err
	}
	out := Response[T]{Data: res.Data, Rendered: rendered}
	if out.Data == nil {
		out.Data = []T{}
	}
	if req.ShowCounts {
		if p := res.Meta.Pagination; p != nil && p.Total >= 0 {
			out.Count = &Count{InQuery: p.Total, InPartition: p.Total}
		}
	}
	return out, nil
}

// CountFunc returns count(*) over a builder (its BuildCount query). Satisfied
// by a closure over the adapter's count path.
type CountFunc func(context.Context, *metaquery.Builder) (int64, error)

// RunWithPartitionCount is Run plus a true InPartition: it builds a second,
// partition-only builder via newBuilder (a fresh WrapXxx()), applies only the
// partition, and counts it. Use when count.inPartition must exclude the search.
func RunWithPartitionCount[T any](
	ctx context.Context,
	newBuilder func() *metaquery.Builder,
	req Request,
	cfg Config,
	scan ScanFunc[T],
	count CountFunc,
) (Response[T], error) {
	out, err := Run(ctx, newBuilder(), req, cfg, scan)
	if err != nil {
		return out, err
	}
	if !req.ShowCounts || out.Count == nil || strings.TrimSpace(req.Partition) == "" {
		return out, nil // no partition → InPartition already equals InQuery
	}
	pb := newBuilder()
	if _, err := search.Apply(pb, req.Partition, cfg.effectiveSearch(pb.OutputColumns())); err != nil {
		return out, fmt.Errorf("dataset: partition count: %w", err)
	}
	n, err := count(ctx, pb)
	if err != nil {
		return out, fmt.Errorf("dataset: partition count: %w", err)
	}
	out.Count.InPartition = n
	return out, nil
}
