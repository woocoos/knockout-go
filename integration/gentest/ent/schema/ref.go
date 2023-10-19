package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
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
	}
}

// Fields of the RefSchema. pick some fields which project need.
func (RefSchema) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
	}
}
