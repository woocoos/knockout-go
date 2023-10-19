package middleware

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/tsingsun/woocoo/contrib/gql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/woocoos/knockout-go/pkg/pagination"
)

// SimplePagination is a middleware that parses the query string for the simple (similar limit,offset) pagination
// use it like:
//
//	gqlsrv.AroundResponses(gqlx.SimplePagination())
func SimplePagination() graphql.ResponseMiddleware {
	return func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		if op := graphql.GetOperationContext(ctx).Operation; op != nil && op.Operation != ast.Query {
			return next(ctx)
		}
		gctx, _ := gql.FromIncomingContext(ctx)
		if gctx != nil {
			sp, err := pagination.NewSimplePagination(gctx.Query("p"), gctx.Query("c"))
			if err != nil {
				return graphql.ErrorResponse(ctx, "pagination error:%v", err)
			}
			if sp != nil {
				ctx = pagination.WithSimplePagination(ctx, sp)
			}
		}
		return next(ctx)
	}
}
