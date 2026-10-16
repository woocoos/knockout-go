package typex

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalInt64ID(i int64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = fmt.Fprintf(w, `"%d"`, i)
	})
}

func UnmarshalInt64ID(v interface{}) (int64, error) {
	switch v := v.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("%T is not an int64", v)
	}
}
