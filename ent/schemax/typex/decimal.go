package typex

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/shopspring/decimal"
	"io"
)

func MarshalDecimal(t decimal.Decimal) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, t.String())
	})
}

func UnmarshalDecimal(v interface{}) (decimal.Decimal, error) {
	switch v := v.(type) {
	case string:
		return decimal.NewFromString(v)
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case int64:
		return decimal.NewFromInt(v), nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case json.Number:
		return decimal.NewFromString(string(v))
	default:
		return decimal.Decimal{}, fmt.Errorf("%T is not an decimal", v)
	}
}
