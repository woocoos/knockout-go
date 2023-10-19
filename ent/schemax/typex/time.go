package typex

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"strconv"
	"time"
)

func MarshalDuration(t time.Duration) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatInt(int64(t), 10))
	})
}

func UnmarshalDuration(v interface{}) (time.Duration, error) {
	switch v := v.(type) {
	case int64:
		return time.Duration(v), nil
	case string:
		return time.ParseDuration(v)
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return time.Duration(i), nil
	default:
		return 0, fmt.Errorf("invalid type %T, expect string", v)
	}
}
