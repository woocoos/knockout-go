package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/woocoos/knockout-go/ent/schemax"
	gen "github.com/woocoos/knockout-go/integration/helloapp/ent"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/intercept"
)

type Hello struct {
	ent.Schema
}

// Annotations of the World.
func (Hello) Annotations() []schema.Annotation {
	return []schema.Annotation{
		schemax.Resources([]string{"name"}),
		schemax.TenantField("org_id"),
	}
}

func (Hello) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemax.IntID{},
		schemax.NewTenantMixin[intercept.Query, *gen.Client]("", intercept.NewQuery,
			schemax.WithTenantMixinStorageKey[intercept.Query, *gen.Client]("org_id")),
	}
}

func (Hello) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Int(schemax.FieldTenantID).StorageKey("org_id").Immutable(),
	}
}
