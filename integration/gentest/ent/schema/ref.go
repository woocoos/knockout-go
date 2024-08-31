package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// RefSchema is in other database schema but in the same database instance.
// it should not be migrated(create or update)
type RefSchema struct {
	ent.Schema
}

func (RefSchema) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ref_table"},
		entgql.RelayConnection(),
	}
}

// Fields of the RefSchema. pick some fields which project need.
func (RefSchema) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Int("user_id"),
	}
}

// Edges of the RefSchema.
func (RefSchema) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Unique().Ref("refs").Unique().Required().Field("user_id"),
	}
}
