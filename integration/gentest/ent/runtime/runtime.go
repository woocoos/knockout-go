// Code generated by ent, DO NOT EDIT.

package runtime

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/woocoos/knockout-go/integration/gentest/ent/schema"
	"github.com/woocoos/knockout-go/integration/gentest/ent/user"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	userMixin := schema.User{}.Mixin()
	userMixinHooks0 := userMixin[0].Hooks()
	user.Hooks[0] = userMixinHooks0[0]
	userFields := schema.User{}.Fields()
	_ = userFields
	// userDescName is the schema descriptor for name field.
	userDescName := userFields[1].Descriptor()
	// user.NameValidator is a validator for the "name" field. It is called by the builders before save.
	user.NameValidator = userDescName.Validators[0].(func(string) error)
	// userDescCreatedAt is the schema descriptor for created_at field.
	userDescCreatedAt := userFields[2].Descriptor()
	// user.DefaultCreatedAt holds the default value on creation for the created_at field.
	user.DefaultCreatedAt = userDescCreatedAt.Default.(func() time.Time)
	// userDescMoney is the schema descriptor for money field.
	userDescMoney := userFields[3].Descriptor()
	// user.DefaultMoney holds the default value on creation for the money field.
	user.DefaultMoney = userDescMoney.Default.(func() decimal.Decimal)
	// user.MoneyValidator is a validator for the "money" field. It is called by the builders before save.
	user.MoneyValidator = func() func(decimal.Decimal) error {
		validators := userDescMoney.Validators
		fns := [...]func(decimal.Decimal) error{
			validators[0].(func(decimal.Decimal) error),
			validators[1].(func(decimal.Decimal) error),
		}
		return func(money decimal.Decimal) error {
			for _, fn := range fns {
				if err := fn(money); err != nil {
					return err
				}
			}
			return nil
		}
	}()
}

const (
	Version = "v0.12.5"                                         // Version of ent codegen.
	Sum     = "h1:KREM5E4CSoej4zeGa88Ou/gfturAnpUv0mzAjch1sj4=" // Sum of ent codegen.
)
