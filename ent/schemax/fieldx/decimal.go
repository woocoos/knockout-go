package fieldx

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

// Decimal returns a new decimal field with type github.com/shopspring/decimal.
func Decimal(name string) *decimalBuilder {
	b := &decimalBuilder{&field.Descriptor{
		Name: name,
	}}
	ot := field.String("decimal-tmp").GoType(decimal.Decimal{})
	b.desc.Info = ot.Descriptor().Info
	b.desc.Err = ot.Descriptor().Err
	b.desc.SchemaType = ot.Descriptor().SchemaType
	b.desc.Annotations = append(b.desc.Annotations, entgql.Type("Decimal"))
	return b
}

// decimalBuilder is the builder for decimal field.
//
// Testing in integration/gentest/ent/schema/user.go
type decimalBuilder struct {
	desc *field.Descriptor
}

// Precision sets the precision and scale of the decimal field.
func (b *decimalBuilder) Precision(precision, scale int) *decimalBuilder {
	b.SchemaType(map[string]string{
		dialect.MySQL:    fmt.Sprintf("decimal(%d,%d)", precision, scale),
		dialect.SQLite:   fmt.Sprintf("decimal(%d,%d)", precision, scale),
		dialect.Postgres: fmt.Sprintf("decimal(%d,%d)", precision, scale),
	})
	return b
}

func (b *decimalBuilder) Unique() *decimalBuilder {
	b.desc.Unique = true
	return b
}

func (b *decimalBuilder) Range(i, j decimal.Decimal) *decimalBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v decimal.Decimal) error {
		if v.Cmp(i) == -1 || v.Cmp(j) == 1 {
			return errors.New("value out of range")
		}
		return nil
	})
	return b
}

func (b *decimalBuilder) Min(i decimal.Decimal) *decimalBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v decimal.Decimal) error {
		if v.Cmp(i) == -1 {
			return errors.New("value out of range")
		}
		return nil
	})
	return b
}

func (b *decimalBuilder) Max(i decimal.Decimal) *decimalBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v decimal.Decimal) error {
		if v.Cmp(i) == 1 {
			return errors.New("value out of range")
		}
		return nil
	})
	return b
}

// Default sets the default value of the field.
func (b *decimalBuilder) Default(d float64) *decimalBuilder {
	b.desc.Default = func() decimal.Decimal {
		return decimal.NewFromFloat(d)
	}
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *decimalBuilder) Nillable() *decimalBuilder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *decimalBuilder) Comment(c string) *decimalBuilder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *decimalBuilder) Optional() *decimalBuilder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *decimalBuilder) Immutable() *decimalBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *decimalBuilder) StructTag(s string) *decimalBuilder {
	b.desc.Tag = s
	return b
}

// Validate adds a validator for this field. Operation fails if the validation fails.
func (b *decimalBuilder) Validate(fn func(d decimal.Decimal) error) *decimalBuilder {
	b.desc.Validators = append(b.desc.Validators, fn)
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *decimalBuilder) StorageKey(key string) *decimalBuilder {
	b.desc.StorageKey = key
	return b
}

func (b *decimalBuilder) SchemaType(types map[string]string) *decimalBuilder {
	b.desc.SchemaType = types
	return b
}

func (b *decimalBuilder) Annotations(annotations ...schema.Annotation) *decimalBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *decimalBuilder) Descriptor() *field.Descriptor {
	return b.desc
}
