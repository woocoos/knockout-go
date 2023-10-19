// Code generated by ent, DO NOT EDIT.

package world

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the world type in the database.
	Label = "world"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldTenantID holds the string denoting the tenant_id field in the database.
	FieldTenantID = "tenant_id"
	// FieldDeletedAt holds the string denoting the deleted_at field in the database.
	FieldDeletedAt = "deleted_at"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldPowerBy holds the string denoting the power_by field in the database.
	FieldPowerBy = "power_by"
	// Table holds the table name of the world in the database.
	Table = "worlds"
)

// Columns holds all SQL columns for world fields.
var Columns = []string{
	FieldID,
	FieldTenantID,
	FieldDeletedAt,
	FieldName,
	FieldPowerBy,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

// Note that the variables below are initialized by the runtime
// package on the initialization of the application. Therefore,
// it should be imported in the main as follows:
//
//	import _ "github.com/woocoos/knockout-go/integration/helloapp/ent/runtime"
var (
	Hooks        [2]ent.Hook
	Interceptors [2]ent.Interceptor
	// DefaultPowerBy holds the default value on creation for the "power_by" field.
	DefaultPowerBy string
)

// OrderOption defines the ordering options for the World queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByTenantID orders the results by the tenant_id field.
func ByTenantID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTenantID, opts...).ToFunc()
}

// ByDeletedAt orders the results by the deleted_at field.
func ByDeletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeletedAt, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByPowerBy orders the results by the power_by field.
func ByPowerBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPowerBy, opts...).ToFunc()
}
