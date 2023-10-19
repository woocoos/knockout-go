package schemax

import (
	"context"
	"entgo.io/contrib/entgql"
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"errors"
	"fmt"
	"github.com/tsingsun/woocoo/pkg/security"
	"time"
)

type AuditMixin struct {
	mixin.Schema
}

func (e AuditMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("created_by").Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput), entproto.Field(2)),
		field.Time("created_at").Immutable().Default(time.Now).Immutable().
			Annotations(entgql.OrderField("createdAt"), entgql.Skip(entgql.SkipMutationCreateInput),
				entproto.Field(3)),
		field.Int("updated_by").Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entproto.Field(4)),
		field.Time("updated_at").Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entproto.Field(5)),
	}
}

func (AuditMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		AuditHook,
	}
}

func AuditHook(next ent.Mutator) ent.Mutator {
	type AuditLogger interface {
		SetCreatedAt(time.Time)
		CreatedAt() (value time.Time, exists bool)
		SetCreatedBy(int)
		CreatedBy() (id int, exists bool)
		SetUpdatedAt(time.Time)
		UpdatedAt() (value time.Time, exists bool)
		SetUpdatedBy(int)
		UpdatedBy() (id int, exists bool)
	}
	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		ml, ok := m.(AuditLogger)
		if !ok {
			return nil, fmt.Errorf("unexpected audit-log call from mutation type %T", m)
		}
		switch op := m.Op(); {
		case op.Is(ent.OpCreate):
			ml.SetCreatedAt(time.Now())
			if _, exists := ml.CreatedBy(); !exists {
				uid, err := getUserID(ctx)
				if err != nil {
					return nil, err
				}
				ml.SetCreatedBy(uid)
			}
		case op.Is(ent.OpUpdateOne | ent.OpUpdate):
			ml.SetUpdatedAt(time.Now())
			if _, exists := ml.UpdatedBy(); !exists {
				uid, err := getUserID(ctx)
				if err != nil {
					return nil, err
				}
				ml.SetUpdatedBy(uid)
			}
		}
		return next.Mutate(ctx, m)
	})
}

func getUserID(ctx context.Context) (uid int, err error) {
	gi, ok := security.GenericIdentityFromContext(ctx)
	if !ok {
		return 0, errors.New("identity no found")
	}
	uid = gi.NameInt()
	if uid == 0 {
		return 0, fmt.Errorf("unexpected identity %s", gi.Name())
	}
	return
}
