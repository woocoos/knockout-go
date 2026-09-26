package schemax

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/tsingsun/woocoo/pkg/security"
)

// AuditMixin 为 schema 添加 created_at, created_by, updated_at, updated_by 四个审计字段.
// 时间戳(created_at, updated_at)由 Hook 强制设置, 不允许外部覆盖;
// user ID(created_by, updated_by)允许外部预填, Hook 仅在未设置时从 context 中获取.
// updated_by 为 Optional, 若其值为空则表明该记录自创建后未被更新过.
type AuditMixin struct {
	mixin.Schema
	// Precision is the precision of the time.Time field.
	Precision int
}

func (e AuditMixin) Fields() []ent.Field {
	ca := field.Time("created_at").Immutable()
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
			now := time.Now()
			ml.SetCreatedAt(now)
			ml.SetUpdatedAt(now)
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
		return 0, errors.New("user not found")
	}
	uid, _ = strconv.Atoi(user.Identity().Name())
	if uid == 0 {
		return 0, fmt.Errorf("unexpected identity %s", user.Identity().Name())
	}
	return
}
