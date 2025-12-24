package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/woocoos/knockout-go/ent/schemax"
)

// ExGIDSchema 为测试不接收node请求
type ExGIDSchema struct {
	ent.Schema
}

func (s ExGIDSchema) Annotations() []schema.Annotation {
	return []schema.Annotation{
		schemax.ExcludeNodeQuery(),
	}
}

// Fields of the RefSchema. pick some fields which project need.
func (ExGIDSchema) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
	}
}
