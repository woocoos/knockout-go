package schemax

import (
	"context"
	"entgo.io/contrib/entgql"
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"errors"
	"fmt"
	"github.com/tsingsun/woocoo/pkg/security"
	"strconv"
	"time"
)

// AuditMixin is a mixin that adds created_at, created_by, updated_at, and updated_by fields to the schema.
type AuditMixin struct {
	mixin.Schema
	// Precision is the precision of the time.Time field.
	Precision int
}

func (e AuditMixin) Fields() []ent.Field {
	ca := field.Time("created_at").Immutable().Default(time.Now).Immutable()
	ua := field.Time("updated_at").Optional()
	if e.Precision > 0 {
		st := map[string]string{
			dialect.Postgres: fmt.Sprintf("TIMESTAMP(%d)", e.Precision),
			dialect.MySQL:    fmt.Sprintf("TIMESTAMP(%d)", e.Precision),
		}
		ca = ca.SchemaType(st)
		ua = ua.SchemaType(st)
	}
	return []ent.Field{
		field.Int("created_by").Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput), entproto.Field(2)),
		ca.Annotations(entgql.OrderField("createdAt"), entgql.Skip(entgql.SkipMutationCreateInput),
			entproto.Field(3)),
		field.Int("updated_by").Optional().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entproto.Field(4)),
		ua.Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
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
	user, ok := security.FromContext(ctx)
	if !ok {
		return 0, errors.New("user no found")
	}
	uid, _ = strconv.Atoi(user.Identity().Name())
	if uid == 0 {
		return 0, fmt.Errorf("unexpected identity %s", user.Identity().Name())
	}
	return
}
