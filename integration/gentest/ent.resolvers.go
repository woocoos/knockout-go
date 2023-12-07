package gentest

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.41

import (
	"context"
	"fmt"

	"entgo.io/contrib/entgql"
	"github.com/gin-gonic/gin"
	"github.com/woocoos/knockout-go/integration/gentest/ent"
	"github.com/woocoos/knockout-go/pkg/pagination"
)

// Node is the resolver for the node field.
func (r *queryResolver) Node(ctx context.Context, id string) (ent.Noder, error) {
	panic(fmt.Errorf("not implemented"))
}

// Nodes is the resolver for the nodes field.
func (r *queryResolver) Nodes(ctx context.Context, ids []string) ([]ent.Noder, error) {
	panic(fmt.Errorf("not implemented"))
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, after *entgql.Cursor[int], first *int, before *entgql.Cursor[int], last *int, orderBy *ent.UserOrder, where *ent.UserWhereInput) (*ent.UserConnection, error) {
	gctx := ctx.Value(gin.ContextKey).(*gin.Context)
	sp, err := pagination.NewSimplePagination(gctx.Query("p"), gctx.Query("c"))
	if err != nil {
		return nil, err
	}
	return r.client.User.Query().Paginate(pagination.WithSimplePagination(ctx, sp), after, first, before, last)
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
