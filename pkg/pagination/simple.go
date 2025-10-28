package pagination

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"entgo.io/ent/dialect/sql"
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

// Offset returns the offset for the given first and last values.
func (sp *SimplePagination) Offset(first, last *int) (offset int) {
	if first != nil {
		offset = (sp.PageIndex - sp.CurrentIndex - 1) * *first
	}
	if last != nil {
		offset = (sp.CurrentIndex - sp.PageIndex - 1) * *last
	}
	return
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

// LimitPerRow returns a query modifier that limits the number of (edges) rows returned
// by the given partition and pagination. This helper function is used mainly by the paginated API to
// override the default Limit behavior for limit returned per node and not limit for all query.
func LimitPerRow(partitionBy string, limit, offset int, orderBy ...sql.Querier) func(s *sql.Selector) {
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

var (
	ErrFirstOrLastMissing = errors.New("first or last is required")
	ErrGreaterThanMaxRow  = errors.New("first or last is greater than maxRow")
)

// NeedLimit returns an error if first or last is not set, if maxRow is set and first or last is greater than maxRow, it returns an error.
func NeedLimit(first, last *int, maxRow int) error {
	if first == nil && last == nil {
		return ErrFirstOrLastMissing
	}
	if maxRow > 0 && (first != nil && *first > maxRow || last != nil && *last > maxRow) {
		return fmt.Errorf("%w: %d", ErrGreaterThanMaxRow, maxRow)
	}
	return nil
}
