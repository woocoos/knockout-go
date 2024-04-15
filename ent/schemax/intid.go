package schemax

import (
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// IntID helps to generate an int type id. It is used for the primary key of the table.
type IntID struct {
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
