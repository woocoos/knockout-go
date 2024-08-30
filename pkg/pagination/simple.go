package pagination

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"strconv"
)

var (
	simplePaginationKey = "_woocoos/knockout/simplePagination"
)

// SimplePagination is a simple pagination implementation.
type SimplePagination struct {
	CurrentIndex int
	PageIndex    int
}

// NewSimplePagination creates a new SimplePagination from the given page and count. If both are empty, it returns nil.
func NewSimplePagination(p, c string) (sp *SimplePagination, err error) {
	if p == "" && c == "" {
		return nil, nil
	}
	sp = &SimplePagination{}
	if p != "" {
		if sp.PageIndex, err = strconv.Atoi(p); err != nil {
			return nil, err
		}
	}
	if c != "" {
		if sp.CurrentIndex, err = strconv.Atoi(c); err != nil {
			return nil, err
		}
	}
	return sp, nil
}

// SimplePaginationFromContext returns the SimplePagination from the given context.
func SimplePaginationFromContext(ctx context.Context) (*SimplePagination, bool) {
	sp, ok := ctx.Value(simplePaginationKey).(*SimplePagination)
	return sp, ok
}

// WithSimplePagination returns a new context with the given SimplePagination.
func WithSimplePagination(ctx context.Context, sp *SimplePagination) context.Context {
	if sp == nil {
		return ctx
	}
	return context.WithValue(ctx, simplePaginationKey, sp)
}

// LimitRows returns a function that limits the rows of the selector based on the given partitionBy, limit, first, last and orderBy.
// It is used for pagination template.
func LimitRows(ctx context.Context, partitionBy string, limit int, first, last *int, orderBy ...sql.Querier) func(s *sql.Selector) {
	offset := 0
	if sp, ok := SimplePaginationFromContext(ctx); ok {
		if first != nil {
			offset = (sp.PageIndex - sp.CurrentIndex - 1) * *first
		}
		if last != nil {
			offset = (sp.CurrentIndex - sp.PageIndex - 1) * *last
		}
	}
	return func(s *sql.Selector) {
		d := sql.Dialect(s.Dialect())
		s.SetDistinct(false)
		with := d.With("src_query").
			As(s.Clone()).
			With("limited_query").
			As(
				d.Select("*").
					AppendSelectExprAs(
						sql.RowNumber().PartitionBy(partitionBy).OrderExpr(orderBy...),
						"row_number",
					).
					From(d.Table("src_query")),
			)
		t := d.Table("limited_query").As(s.TableName())
		if offset != 0 {
			*s = *d.Select(s.UnqualifiedColumns()...).
				From(t).
				Where(sql.GT(t.C("row_number"), offset)).Limit(limit).
				Prefix(with)
		} else {
			*s = *d.Select(s.UnqualifiedColumns()...).
				From(t).
				Where(sql.LTE(t.C("row_number"), limit)).
				Prefix(with)
		}
	}
}
