package schemax

import (
	"context"
	"entgo.io/contrib/entgql"
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"fmt"
	"time"
)

// SoftDeleteMixin implements the soft delete pattern for schemas.
type SoftDeleteMixin[T Query, Q Mutator] struct {
	mixin.Schema
	QueryFunc func(ent.Query) (T, error)
}

func NewSoftDeleteMixin[T Query, Q Mutator](qf func(ent.Query) (T, error)) SoftDeleteMixin[T, Q] {
	return SoftDeleteMixin[T, Q]{
		QueryFunc: qf,
	}
}

// Fields of the SoftDeleteMixin.
func (SoftDeleteMixin[T, Q]) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput), entproto.Skip()),
	}
}

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// Interceptors of the SoftDeleteMixin.
func (d SoftDeleteMixin[T, Q]) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			// Skip soft-delete, means include soft-deleted entities.
			if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
				return nil
			}
			df, err := d.QueryFunc(q)
			if err != nil {
				return err
			}
			d.P(df)
			return nil
		}),
	}
}

type softDeleter[Q Mutator] interface {
	Query
	Client() Q
	SetOp(ent.Op)
	SetDeletedAt(time.Time)
}

// Hooks of the SoftDeleteMixin.
func (d SoftDeleteMixin[T, Q]) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if !m.Op().Is(ent.OpDelete | ent.OpDeleteOne) {
					return next.Mutate(ctx, m)
				}
				// Skip soft-delete, means delete the entity permanently.
				if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
					return next.Mutate(ctx, m)
				}
				mx, ok := m.(softDeleter[Q])
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				d.P(mx)
				mx.SetOp(ent.OpUpdate)
				mx.SetDeletedAt(time.Now())
				return mx.Client().Mutate(ctx, m)
			})
		},
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (d SoftDeleteMixin[T, Q]) P(w Query) {
	w.WhereP(
		sql.FieldIsNull("deleted_at"),
	)
}
