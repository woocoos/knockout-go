package typex

import (
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

func MarshalMapString(val map[string]string) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(val)
		if err != nil {
			panic(err)
		}
	})
}

func UnmarshalMapString(v interface{}) (map[string]string, error) {
	if v, ok := v.(map[string]any); ok {
		m := make(map[string]string, len(v))
		for k, av := range v {
			if _, ok := av.(string); !ok {
				return nil, fmt.Errorf("invalid type %T, expect string", av)
			}
			m[k] = av.(string)
		}
		return m, nil
	}
	return nil, fmt.Errorf("%T is not a map", v)
}
