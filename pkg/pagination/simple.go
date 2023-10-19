package pagination

import (
	"context"
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
