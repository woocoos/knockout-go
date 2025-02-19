package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// NoCache is in other database schema but in the same database instance.
// it should not be migrated(create or update)
type NoCache struct {
	ent.Schema
}

func (NoCache) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
	}
}

// Fields of the RefSchema. pick some fields which project need.
func (NoCache) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Int("user_id"),
	}
}
