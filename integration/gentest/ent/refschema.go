// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/woocoos/knockout-go/integration/gentest/ent/refschema"
	"github.com/woocoos/knockout-go/integration/gentest/ent/user"
)

// RefSchema is the model entity for the RefSchema schema.
type RefSchema struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// UserID holds the value of the "user_id" field.
	UserID int `json:"user_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the RefSchemaQuery when eager-loading is set.
	Edges        RefSchemaEdges `json:"edges"`
	selectValues sql.SelectValues
}

// RefSchemaEdges holds the relations/edges for other nodes in the graph.
type RefSchemaEdges struct {
	// User holds the value of the user edge.
	User *User `json:"user,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
	// totalCount holds the count of the edges above.
	totalCount [1]map[string]int
}

// UserOrErr returns the User value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RefSchemaEdges) UserOrErr() (*User, error) {
	if e.User != nil {
		return e.User, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "user"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*RefSchema) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case refschema.FieldID, refschema.FieldUserID:
			values[i] = new(sql.NullInt64)
		case refschema.FieldName:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the RefSchema fields.
func (rs *RefSchema) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case refschema.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			rs.ID = int(value.Int64)
		case refschema.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				rs.Name = value.String
			}
		case refschema.FieldUserID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value.Valid {
				rs.UserID = int(value.Int64)
			}
		default:
			rs.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the RefSchema.
// This includes values selected through modifiers, order, etc.
func (rs *RefSchema) Value(name string) (ent.Value, error) {
	return rs.selectValues.Get(name)
}

// QueryUser queries the "user" edge of the RefSchema entity.
func (rs *RefSchema) QueryUser() *UserQuery {
	return NewRefSchemaClient(rs.config).QueryUser(rs)
}

// Update returns a builder for updating this RefSchema.
// Note that you need to call RefSchema.Unwrap() before calling this method if this RefSchema
// was returned from a transaction, and the transaction was committed or rolled back.
func (rs *RefSchema) Update() *RefSchemaUpdateOne {
	return NewRefSchemaClient(rs.config).UpdateOne(rs)
}

// Unwrap unwraps the RefSchema entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (rs *RefSchema) Unwrap() *RefSchema {
	_tx, ok := rs.config.driver.(*txDriver)
	if !ok {
		panic("ent: RefSchema is not a transactional entity")
	}
	rs.config.driver = _tx.drv
	return rs
}

// String implements the fmt.Stringer.
func (rs *RefSchema) String() string {
	var builder strings.Builder
	builder.WriteString("RefSchema(")
	builder.WriteString(fmt.Sprintf("id=%v, ", rs.ID))
	builder.WriteString("name=")
	builder.WriteString(rs.Name)
	builder.WriteString(", ")
	builder.WriteString("user_id=")
	builder.WriteString(fmt.Sprintf("%v", rs.UserID))
	builder.WriteByte(')')
	return builder.String()
}

// RefSchemas is a parsable slice of RefSchema.
type RefSchemas []*RefSchema
