// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/woocoos/knockout-go/integration/nocache/ent/nocache"
	"github.com/woocoos/knockout-go/pkg/pagination"
)

// Common entgql types.
type (
	Cursor         = entgql.Cursor[int]
	PageInfo       = entgql.PageInfo[int]
	OrderDirection = entgql.OrderDirection
)

func orderFunc(o OrderDirection, field string) func(*sql.Selector) {
	if o == entgql.OrderDirectionDesc {
		return Desc(field)
	}
	return Asc(field)
}

const errInvalidPagination = "INVALID_PAGINATION"

func validateFirstLast(first, last *int) (err *gqlerror.Error) {
	switch {
	case first != nil && last != nil:
		err = &gqlerror.Error{
			Message: "Passing both `first` and `last` to paginate a connection is not supported.",
		}
	case first != nil && *first < 0:
		err = &gqlerror.Error{
			Message: "`first` on a connection cannot be less than zero.",
		}
		errcode.Set(err, errInvalidPagination)
	case last != nil && *last < 0:
		err = &gqlerror.Error{
			Message: "`last` on a connection cannot be less than zero.",
		}
		errcode.Set(err, errInvalidPagination)
	}
	return err
}

func collectedField(ctx context.Context, path ...string) *graphql.CollectedField {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return nil
	}
	field := fc.Field
	oc := graphql.GetOperationContext(ctx)
walk:
	for _, name := range path {
		for _, f := range graphql.CollectFields(oc, field.Selections, nil) {
			if f.Alias == name {
				field = f
				continue walk
			}
		}
		return nil
	}
	return &field
}

func hasCollectedField(ctx context.Context, path ...string) bool {
	if graphql.GetFieldContext(ctx) == nil {
		return true
	}
	return collectedField(ctx, path...) != nil
}

const (
	edgesField      = "edges"
	nodeField       = "node"
	pageInfoField   = "pageInfo"
	totalCountField = "totalCount"
)

func paginateLimit(first, last *int) int {
	var limit int
	if first != nil {
		limit = *first + 1
	} else if last != nil {
		limit = *last + 1
	}
	return limit
}

// NoCacheEdge is the edge representation of NoCache.
type NoCacheEdge struct {
	Node   *NoCache `json:"node"`
	Cursor Cursor   `json:"cursor"`
}

// NoCacheConnection is the connection containing edges to NoCache.
type NoCacheConnection struct {
	Edges      []*NoCacheEdge `json:"edges"`
	PageInfo   PageInfo       `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

func (c *NoCacheConnection) build(nodes []*NoCache, pager *nocachePager, after *Cursor, first *int, before *Cursor, last *int) {
	c.PageInfo.HasNextPage = before != nil
	c.PageInfo.HasPreviousPage = after != nil
	if first != nil && *first+1 == len(nodes) {
		c.PageInfo.HasNextPage = true
		nodes = nodes[:len(nodes)-1]
	} else if last != nil && *last+1 == len(nodes) {
		c.PageInfo.HasPreviousPage = true
		nodes = nodes[:len(nodes)-1]
	}
	var nodeAt func(int) *NoCache
	if last != nil {
		n := len(nodes) - 1
		nodeAt = func(i int) *NoCache {
			return nodes[n-i]
		}
	} else {
		nodeAt = func(i int) *NoCache {
			return nodes[i]
		}
	}
	c.Edges = make([]*NoCacheEdge, len(nodes))
	for i := range nodes {
		node := nodeAt(i)
		c.Edges[i] = &NoCacheEdge{
			Node:   node,
			Cursor: pager.toCursor(node),
		}
	}
	if l := len(c.Edges); l > 0 {
		c.PageInfo.StartCursor = &c.Edges[0].Cursor
		c.PageInfo.EndCursor = &c.Edges[l-1].Cursor
	}
	if c.TotalCount == 0 {
		c.TotalCount = len(nodes)
	}
}

// NoCachePaginateOption enables pagination customization.
type NoCachePaginateOption func(*nocachePager) error

// WithNoCacheOrder configures pagination ordering.
func WithNoCacheOrder(order *NoCacheOrder) NoCachePaginateOption {
	if order == nil {
		order = DefaultNoCacheOrder
	}
	o := *order
	return func(pager *nocachePager) error {
		if err := o.Direction.Validate(); err != nil {
			return err
		}
		if o.Field == nil {
			o.Field = DefaultNoCacheOrder.Field
		}
		pager.order = &o
		return nil
	}
}

// WithNoCacheFilter configures pagination filter.
func WithNoCacheFilter(filter func(*NoCacheQuery) (*NoCacheQuery, error)) NoCachePaginateOption {
	return func(pager *nocachePager) error {
		if filter == nil {
			return errors.New("NoCacheQuery filter cannot be nil")
		}
		pager.filter = filter
		return nil
	}
}

type nocachePager struct {
	reverse bool
	order   *NoCacheOrder
	filter  func(*NoCacheQuery) (*NoCacheQuery, error)
}

func newNoCachePager(opts []NoCachePaginateOption, reverse bool) (*nocachePager, error) {
	pager := &nocachePager{reverse: reverse}
	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}
	if pager.order == nil {
		pager.order = DefaultNoCacheOrder
	}
	return pager, nil
}

func (p *nocachePager) applyFilter(query *NoCacheQuery) (*NoCacheQuery, error) {
	if p.filter != nil {
		return p.filter(query)
	}
	return query, nil
}

func (p *nocachePager) toCursor(nc *NoCache) Cursor {
	return p.order.Field.toCursor(nc)
}

func (p *nocachePager) applyCursors(query *NoCacheQuery, after, before *Cursor) (*NoCacheQuery, error) {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	for _, predicate := range entgql.CursorsPredicate(after, before, DefaultNoCacheOrder.Field.column, p.order.Field.column, direction) {
		query = query.Where(predicate)
	}
	return query, nil
}

func (p *nocachePager) applyOrder(query *NoCacheQuery) *NoCacheQuery {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	query = query.Order(p.order.Field.toTerm(direction.OrderTermOption()))
	if p.order.Field != DefaultNoCacheOrder.Field {
		query = query.Order(DefaultNoCacheOrder.Field.toTerm(direction.OrderTermOption()))
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return query
}

func (p *nocachePager) orderExpr(query *NoCacheQuery) sql.Querier {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return sql.ExprFunc(func(b *sql.Builder) {
		b.Ident(p.order.Field.column).Pad().WriteString(string(direction))
		if p.order.Field != DefaultNoCacheOrder.Field {
			b.Comma().Ident(DefaultNoCacheOrder.Field.column).Pad().WriteString(string(direction))
		}
	})
}

// Paginate executes the query and returns a relay based cursor connection to NoCache.
func (nc *NoCacheQuery) Paginate(
	ctx context.Context, after *Cursor, first *int,
	before *Cursor, last *int, opts ...NoCachePaginateOption,
) (*NoCacheConnection, error) {
	if err := validateFirstLast(first, last); err != nil {
		return nil, err
	}
	pager, err := newNoCachePager(opts, last != nil)
	if err != nil {
		return nil, err
	}
	if nc, err = pager.applyFilter(nc); err != nil {
		return nil, err
	}
	conn := &NoCacheConnection{Edges: []*NoCacheEdge{}}
	ignoredEdges := !hasCollectedField(ctx, edgesField)
	if hasCollectedField(ctx, totalCountField) || hasCollectedField(ctx, pageInfoField) {
		hasPagination := after != nil || first != nil || before != nil || last != nil
		if hasPagination || ignoredEdges {
			c := nc.Clone()
			c.ctx.Fields = nil
			if conn.TotalCount, err = c.Count(ctx); err != nil {
				return nil, err
			}
			conn.PageInfo.HasNextPage = first != nil && conn.TotalCount > 0
			conn.PageInfo.HasPreviousPage = last != nil && conn.TotalCount > 0
		}
	}
	if ignoredEdges || (first != nil && *first == 0) || (last != nil && *last == 0) {
		return conn, nil
	}
	if nc, err = pager.applyCursors(nc, after, before); err != nil {
		return nil, err
	}
	limit := paginateLimit(first, last)
	if limit != 0 {
		nc.Limit(limit)
	}
	if sp, ok := pagination.SimplePaginationFromContext(ctx); ok {
		nc.Offset(sp.Offset(first, last))
	}
	if field := collectedField(ctx, edgesField, nodeField); field != nil {
		if err := nc.collectField(ctx, limit == 1, graphql.GetOperationContext(ctx), *field, []string{edgesField, nodeField}); err != nil {
			return nil, err
		}
	}
	nc = pager.applyOrder(nc)
	nodes, err := nc.All(ctx)
	if err != nil {
		return nil, err
	}
	conn.build(nodes, pager, after, first, before, last)
	return conn, nil
}

// NoCacheOrderField defines the ordering field of NoCache.
type NoCacheOrderField struct {
	// Value extracts the ordering value from the given NoCache.
	Value    func(*NoCache) (ent.Value, error)
	column   string // field or computed.
	toTerm   func(...sql.OrderTermOption) nocache.OrderOption
	toCursor func(*NoCache) Cursor
}

// NoCacheOrder defines the ordering of NoCache.
type NoCacheOrder struct {
	Direction OrderDirection     `json:"direction"`
	Field     *NoCacheOrderField `json:"field"`
}

// DefaultNoCacheOrder is the default ordering of NoCache.
var DefaultNoCacheOrder = &NoCacheOrder{
	Direction: entgql.OrderDirectionAsc,
	Field: &NoCacheOrderField{
		Value: func(nc *NoCache) (ent.Value, error) {
			return nc.ID, nil
		},
		column: nocache.FieldID,
		toTerm: nocache.ByID,
		toCursor: func(nc *NoCache) Cursor {
			return Cursor{ID: nc.ID}
		},
	},
}

// ToEdge converts NoCache into NoCacheEdge.
func (nc *NoCache) ToEdge(order *NoCacheOrder) *NoCacheEdge {
	if order == nil {
		order = DefaultNoCacheOrder
	}
	return &NoCacheEdge{
		Node:   nc,
		Cursor: order.Field.toCursor(nc),
	}
}
