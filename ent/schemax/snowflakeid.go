package schemax

import (
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/woocoos/knockout-go/pkg/snowflake"
)

// SnowFlakeID helps to generate a snowflake type id.
type SnowFlakeID struct {
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

// SchemaType override the ent.Schema.
// The SchemaType of SnowFlakeID is a map of database dialects to the SQL type.
func (SnowFlakeID) SchemaType() map[string]string {
	return map[string]string{
		dialect.MySQL: "bigint",
	}
}
