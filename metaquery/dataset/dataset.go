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

// Response mirrors redline's DataSetResponse[T].
type Response[T any] struct {
	Data     []T            `json:"data"`
	Count    *Count         `json:"count,omitempty"`
	Meta     metaquery.Meta `json:"meta"`
	Rendered Rendered       `json:"rendered"`
}

// Config drives shaping. The embedded search.Config customizes search/partition
// (Fields, Named); the zero value is sensible. Orderable, when non-empty,
// restricts which fields clients may sort by (OrderBy already whitelists
// against output columns, so this is an additional policy layer).
//
// Searchable mirrors redline's FieldConfig.Searchable: when non-empty, only the
// listed columns participate in unqualified ("global") free-text terms; every
// other column becomes targeted-only (still reachable via field:value, but not
// matched by a bare term). When empty, the type-driven default applies (all
// text columns are global). Explicit per-column entries in search.Config.Fields
// always win over this allowlist.
type Config struct {
	search.Config
	Searchable      []string // global free-text columns; empty = all text columns
	Orderable       []string // allowed sort fields; empty = any output column
	DefaultPageSize int      // 0 -> 25
	MaxPageSize     int      // 0 -> 200
}

// effectiveSearch folds the Searchable allowlist into a search.Config by
// pinning each output column's Scope (global vs targeted), without mutating the
// caller's Config. Columns with an explicit Fields entry are left untouched.
func (c Config) effectiveSearch(cols []metaquery.Column) search.Config {
	if len(c.Searchable) == 0 {
		return c.Config
	}
	allow := make(map[string]bool, len(c.Searchable))
	for _, s := range c.Searchable {
		allow[strings.ToLower(s)] = true
	}
	fields := make(map[string]search.Field, len(cols))
	maps.Copy(fields, c.Config.Fields)
	for _, col := range cols {
		if _, explicit := fields[col.Name]; explicit {
			continue
		}
		f := search.Field{Scope: search.ScopeTargeted}
		if allow[strings.ToLower(col.Name)] {
			f.Scope = search.ScopeGlobal
		}
		fields[col.Name] = f
	}
	return search.Config{Fields: fields, Named: c.Config.Named}
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

	var allow map[string]bool
	if len(cfg.Orderable) > 0 {
		allow = make(map[string]bool, len(cfg.Orderable))
		for _, c := range cfg.Orderable {
			allow[c] = true
		}
	}
	for _, o := range req.Ordering {
		if allow != nil && !allow[o.Field] {
			continue
		}
		dir := metaquery.Asc
		if strings.EqualFold(o.Order, "DESC") {
			dir = metaquery.Desc
		}
		b.OrderBy(o.Field, dir)
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
	out := Response[T]{Data: res.Data, Meta: res.Meta, Rendered: rendered}
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
