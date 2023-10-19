package schemax

import (
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// IntID helps to generate an int type id.
type IntID struct {
	// ID is the unique identifier of the user in the database.
	mixin.Schema
}

func (id IntID) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").SchemaType(id.SchemaType()).
			Annotations(entproto.Field(1)),
	}
}

func (IntID) SchemaType() map[string]string {
	return map[string]string{
		"mysql": "int",
	}
}
