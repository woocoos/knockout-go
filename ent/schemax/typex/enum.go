package typex

import (
	"entgo.io/contrib/entproto"
	"entgo.io/ent/schema"
	"fmt"
	"io"
	"strconv"
)

// SimpleStatus 用于简单型状态字段枚举型
type SimpleStatus string

// UserType values.
const (
	SimpleStatusActive     SimpleStatus = "active"
	SimpleStatusInactive   SimpleStatus = "inactive"
	SimpleStatusProcessing SimpleStatus = "processing"
	SimpleStatusDisabled   SimpleStatus = "disabled"
)

func (st SimpleStatus) String() string {
	return string(st)
}

// SimpleStatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func SimpleStatusValidator(st SimpleStatus) error {
	switch st {
	case SimpleStatusActive, SimpleStatusInactive, SimpleStatusProcessing, SimpleStatusDisabled:
		return nil
	default:
		return fmt.Errorf("status: invalid enum value for status field: %q", st)
	}
}

// Values implements field.EnumValues interface
func (SimpleStatus) Values() []string {
	return []string{
		SimpleStatusActive.String(),
		SimpleStatusInactive.String(),
		SimpleStatusProcessing.String(),
		SimpleStatusDisabled.String(),
	}
}

// MarshalGQL implements graphql.Marshaler interface.
func (st SimpleStatus) MarshalGQL(w io.Writer) {
	io.WriteString(w, strconv.Quote(st.String()))
}

// UnmarshalGQL implements graphql.Unmarshaler interface.
func (st *SimpleStatus) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("enum %T must be a string", val)
	}
	*st = SimpleStatus(str)
	if err := SimpleStatusValidator(*st); err != nil {
		return fmt.Errorf("%s is not a valid SimpleStatus", str)
	}
	return nil
}

func (st SimpleStatus) ProtoAnnotation() schema.Annotation {
	return entproto.Enum(map[string]int32{
		SimpleStatusActive.String():     1,
		SimpleStatusInactive.String():   2,
		SimpleStatusProcessing.String(): 3,
		SimpleStatusDisabled.String():   4,
	})
}
