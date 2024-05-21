package fieldx

import (
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"errors"
	"regexp"
)

func File(name string) *fileBuilder {
	v := &fileBuilder{
		&field.Descriptor{
			Name: name,
		},
	}
	ot := field.String(name).MaxLen(255)
	v.desc = ot.Descriptor()
	return v
}

type fileBuilder struct {
	desc *field.Descriptor
}

// Match adds a regex matcher for this field. Operation fails if the regex fails.
func (b *fileBuilder) Match(re *regexp.Regexp) *fileBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v string) error {
		if !re.MatchString(v) {
			return errors.New("value does not match validation")
		}
		return nil
	})
	return b
}

// MinLen adds a length validator for this field.
// Operation fails if the length of the string is less than the given value.
func (b *fileBuilder) MinLen(i int) *fileBuilder {
	b.desc.Validators = append(b.desc.Validators, func(v string) error {
		if len(v) < i {
			return errors.New("value is less than the required length")
		}
		return nil
	})
	return b
}

// NotEmpty adds a length validator for this field.
// Operation fails if the length of the string is zero.
func (b *fileBuilder) NotEmpty() *fileBuilder {
	return b.MinLen(1)
}

// MaxLen adds a length validator for this field.
// Operation fails if the length of the string is greater than the given value.
func (b *fileBuilder) MaxLen(i int) *fileBuilder {
	b.desc.Size = i
	b.desc.Validators = append(b.desc.Validators, func(v string) error {
		if len(v) > i {
			return errors.New("value is greater than the required length")
		}
		return nil
	})
	return b
}

// SchemaType overrides the default database type with a custom
// schema type (per dialect) for string.
//
//	field.File("name").
//		SchemaType(map[string]string{
//			dialect.MySQL:    "text",
//			dialect.Postgres: "varchar",
//		})
func (b *fileBuilder) SchemaType(types map[string]string) *fileBuilder {
	b.desc.SchemaType = types
	return b
}

// Nillable indicates that this field is a nillable.
// Unlike "Optional" only fields, "Nillable" fields are pointers in the generated struct.
func (b *fileBuilder) Nillable() *fileBuilder {
	b.desc.Nillable = true
	return b
}

// Optional indicates that this field is optional on create.
// Unlike edges, fields are required by default.
func (b *fileBuilder) Optional() *fileBuilder {
	b.desc.Optional = true
	return b
}

// Immutable indicates that this field cannot be updated.
func (b *fileBuilder) Immutable() *fileBuilder {
	b.desc.Immutable = true
	return b
}

// Comment sets the comment of the field.
func (b *fileBuilder) Comment(c string) *fileBuilder {
	b.desc.Comment = c
	return b
}

// StructTag sets the struct tag of the field.
func (b *fileBuilder) StructTag(s string) *fileBuilder {
	b.desc.Tag = s
	return b
}

// StorageKey sets the storage key of the field.
// In SQL dialects is the column name and Gremlin is the property.
func (b *fileBuilder) StorageKey(key string) *fileBuilder {
	b.desc.StorageKey = key
	return b
}

// Validate adds a validator for this field. Operation fails if the validation fails.
func (b *fileBuilder) Validate(fn func(string) error) *fileBuilder {
	b.desc.Validators = append(b.desc.Validators, fn)
	return b
}

// Annotations adds a list of annotations to the field object to be used by
// codegen extensions.
//
//	field.File("dir").
//		Annotations(
//			entgql.OrderField("DIR"),
//		)
func (b *fileBuilder) Annotations(annotations ...schema.Annotation) *fileBuilder {
	b.desc.Annotations = append(b.desc.Annotations, annotations...)
	return b
}

// Descriptor implements the ent.Field interface by returning its descriptor.
func (b *fileBuilder) Descriptor() *field.Descriptor {
	return b.desc
}
