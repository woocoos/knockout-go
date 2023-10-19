package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/woocoos/knockout-go/ent/schemax"
	gen "github.com/woocoos/knockout-go/integration/helloapp/ent"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/intercept"
)

type World struct {
	ent.Schema
}

// Annotations of the World.
func (World) Annotations() []schema.Annotation {
	return []schema.Annotation{
		schemax.Resources([]string{"name"}),
		schemax.TenantField("tenant_id"),
	}
}

func (World) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemax.IntID{},
		schemax.NewTenantMixin[intercept.Query, *gen.Client]("", intercept.NewQuery),
		schemax.NewSoftDeleteMixin[intercept.Query, *gen.Client](intercept.NewQuery),
	}
}

func (World) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("power_by").Optional().Default("0"),
	}
}
