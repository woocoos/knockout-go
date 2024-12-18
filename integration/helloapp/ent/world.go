// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/world"
)

// World is the model entity for the World schema.
type World struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// DeletedAt holds the value of the "deleted_at" field.
	DeletedAt time.Time `json:"deleted_at,omitempty"`
	// TenantID holds the value of the "tenant_id" field.
	TenantID int `json:"tenant_id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// PowerBy holds the value of the "power_by" field.
	PowerBy      string `json:"power_by,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*World) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case world.FieldID, world.FieldTenantID:
			values[i] = new(sql.NullInt64)
		case world.FieldName, world.FieldPowerBy:
			values[i] = new(sql.NullString)
		case world.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the World fields.
func (w *World) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case world.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			w.ID = int(value.Int64)
		case world.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				w.DeletedAt = value.Time
			}
		case world.FieldTenantID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field tenant_id", values[i])
			} else if value.Valid {
				w.TenantID = int(value.Int64)
			}
		case world.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				w.Name = value.String
			}
		case world.FieldPowerBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field power_by", values[i])
			} else if value.Valid {
				w.PowerBy = value.String
			}
		default:
			w.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the World.
// This includes values selected through modifiers, order, etc.
func (w *World) Value(name string) (ent.Value, error) {
	return w.selectValues.Get(name)
}

// Update returns a builder for updating this World.
// Note that you need to call World.Unwrap() before calling this method if this World
// was returned from a transaction, and the transaction was committed or rolled back.
func (w *World) Update() *WorldUpdateOne {
	return NewWorldClient(w.config).UpdateOne(w)
}

// Unwrap unwraps the World entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (w *World) Unwrap() *World {
	_tx, ok := w.config.driver.(*txDriver)
	if !ok {
		panic("ent: World is not a transactional entity")
	}
	w.config.driver = _tx.drv
	return w
}

// String implements the fmt.Stringer.
func (w *World) String() string {
	var builder strings.Builder
	builder.WriteString("World(")
	builder.WriteString(fmt.Sprintf("id=%v, ", w.ID))
	builder.WriteString("deleted_at=")
	builder.WriteString(w.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("tenant_id=")
	builder.WriteString(fmt.Sprintf("%v", w.TenantID))
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(w.Name)
	builder.WriteString(", ")
	builder.WriteString("power_by=")
	builder.WriteString(w.PowerBy)
	builder.WriteByte(')')
	return builder.String()
}

// Worlds is a parsable slice of World.
type Worlds []*World
