// Package mqpgx is the pgx/v5 adapter for metaquery builders. It executes a
// *metaquery.Builder and returns an untyped Result or a typed TypedResult[T].
package mqpgx

import (
	"context"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/jackc/pgx/v5"
)

// Queryer is the minimal pgx surface the adapter needs. Implemented by
// *pgx.Conn, *pgxpool.Pool, pgx.Tx, etc.
type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Run executes b and returns every row as a positional []any in the same
// order as Meta.Columns. If b.WantsTotal(), a follow-up COUNT(*) is issued
// and Meta.Pagination.Total is populated.
func Run(ctx context.Context, q Queryer, b *metaquery.Builder) (*metaquery.Result, error) {
	sql, args, err := b.Build()
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	data, err := collectValues(rows)
	if err != nil {
		return nil, err
	}

	meta := b.Meta()
	if err := populateTotal(ctx, q, b, &meta); err != nil {
		return nil, err
	}
	return &metaquery.Result{Data: data, Meta: meta}, nil
}

// Scan executes b, scans rows into []T via pgx.RowToStructByName, and
// returns a TypedResult with the same Meta envelope as Run. T is validated
// against b.OutputColumns() before querying.
func Scan[T any](ctx context.Context, q Queryer, b *metaquery.Builder) (*metaquery.TypedResult[T], error) {
	if err := metaquery.Validate[T](b); err != nil {
		return nil, err
	}
	sql, args, err := b.Build()
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	meta := b.Meta()
	if err := populateTotal(ctx, q, b, &meta); err != nil {
		return nil, err
	}
	return &metaquery.TypedResult[T]{Data: data, Meta: meta}, nil
}

// ScanOne runs b expecting exactly one row and returns the scanned T plus
// the Meta envelope.
func ScanOne[T any](ctx context.Context, q Queryer, b *metaquery.Builder) (T, metaquery.Meta, error) {
	var zero T
	if err := metaquery.Validate[T](b); err != nil {
		return zero, metaquery.Meta{}, err
	}
	sql, args, err := b.Build()
	if err != nil {
		return zero, metaquery.Meta{}, err
	}
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return zero, metaquery.Meta{}, err
	}
	defer rows.Close()
	v, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return zero, metaquery.Meta{}, err
	}
	return v, b.Meta(), nil
}

// collectValues reads pgx rows into positional []any slices.
func collectValues(rows pgx.Rows) ([][]any, error) {
	defer rows.Close()
	var out [][]any
	for rows.Next() {
		v, err := rows.Values()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// populateTotal runs b.BuildCount() when the caller opted into WithTotal and
// writes the result into meta.Pagination.Total.
func populateTotal(ctx context.Context, q Queryer, b *metaquery.Builder, meta *metaquery.Meta) error {
	if !b.WantsTotal() {
		return nil
	}
	sql, args, err := b.BuildCount()
	if err != nil {
		return err
	}
	var total int64
	if err := q.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return err
	}
	if meta.Pagination == nil {
		meta.Pagination = &metaquery.Pagination{Total: -1}
	}
	meta.Pagination.Total = total
	return nil
}
