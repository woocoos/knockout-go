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

// Decimal creates a new decimal field.
func Decimal(name string) *DecimalBuilder {
	b := &DecimalBuilder{&field.Descriptor{
		Name: name,
	}}
	ot := field.String("decimal-tmp").GoType(decimal.Decimal{})
	b.desc.Info = ot.Descriptor().Info
	b.desc.Err = ot.Descriptor().Err
	b.desc.SchemaType = ot.Descriptor().SchemaType
	b.desc.Annotations = append(b.desc.Annotations, entgql.Type("Decimal"))
	return b
}

type DecimalBuilder struct {
	desc *field.Descriptor
}

// Precision sets the precision and scale of the decimal field.
func (b *DecimalBuilder) Precision(precision, scale int) *DecimalBuilder {
	b.SchemaType(map[string]string{
		dialect.MySQL:    fmt.Sprintf("decimal(%d,%d)", precision, scale),
		dialect.SQLite:   fmt.Sprintf("decimal(%d,%d)", precision, scale),
		dialect.Postgres: fmt.Sprintf("decimal(%d,%d)", precision, scale),
	})
	return b
}

func (b *DecimalBuilder) Unique() *DecimalBuilder {
	b.desc.Unique = true
	return b
}

func (b *DecimalBuilder) Range(i, j decimal.Decimal) *DecimalBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v decimal.Decimal) error {
		if v.Cmp(i) == -1 || v.Cmp(j) == 1 {
			return errors.New("value out of range")
		}
		return nil
	})
	return b
}

func (b *DecimalBuilder) Min(i decimal.Decimal) *DecimalBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v decimal.Decimal) error {
		if v.Cmp(i) == -1 {
			return errors.New("value out of range")
		}
		return nil
	})
	return b
}

func (b *DecimalBuilder) Max(i decimal.Decimal) *DecimalBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v decimal.Decimal) error {
		if v.Cmp(i) == 1 {
			return errors.New("value out of range")
		}
		return nil
	})
	return b
}

// Default sets the default value of the field.
func (b *DecimalBuilder) Default(d float64) *DecimalBuilder {
	b.desc.Default = func() decimal.Decimal {
		return decimal.NewFromFloat(d)
	}
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *DecimalBuilder) Nillable() *DecimalBuilder {
	b.desc.Nillable = true
	return b
}

// Comment sets the comment of the field.
func (b *DecimalBuilder) Comment(c string) *DecimalBuilder {
	b.desc.Comment = c
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *DecimalBuilder) Optional() *DecimalBuilder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *DecimalBuilder) Immutable() *DecimalBuilder {
	b.desc.Immutable = true
	return b
}

// StructTag sets the struct tag of the field.
func (b *DecimalBuilder) StructTag(s string) *DecimalBuilder {
	b.desc.Tag = s
	return b
}

// Validate adds a validator for this field. Operation fails if the validation fails.
func (b *DecimalBuilder) Validate(fn func(d decimal.Decimal) error) *DecimalBuilder {
	b.desc.Validators = append(b.desc.Validators, fn)
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *DecimalBuilder) StorageKey(key string) *DecimalBuilder {
	b.desc.StorageKey = key
	return b
}

func (b *DecimalBuilder) SchemaType(types map[string]string) *DecimalBuilder {
	b.desc.SchemaType = types
	return b
}

func (b *DecimalBuilder) Annotations(annotations ...schema.Annotation) *DecimalBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *DecimalBuilder) Descriptor() *field.Descriptor {
	return b.desc
}
