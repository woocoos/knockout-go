package schemax

import (
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/woocoos/knockout-go/pkg/snowflake"
)

// SnowFlakeID 是采用雪花算法生成的ID.
type SnowFlakeID struct {
	// ID is the unique identifier of the user in the database.
	mixin.Schema
}

func (id SnowFlakeID) Fields() []ent.Field {
	Incremental := false
	return []ent.Field{
		field.Int("id").SchemaType(id.SchemaType()).
			Annotations(entsql.Annotation{Incremental: &Incremental},
				entproto.Field(1)).
			DefaultFunc(func() int {
				return int(snowflake.New().Int64())
			}),
	}
}

func (SnowFlakeID) SchemaType() map[string]string {
	return map[string]string{
		"mysql": "bigint",
	}
}
