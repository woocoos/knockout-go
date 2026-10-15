package typex

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalInt64ID(t *testing.T) {
	marshalInt64ID := func(i int64) string {
		var buf bytes.Buffer
		MarshalInt64ID(i).MarshalGQL(&buf)
		return buf.String()
	}

	assert.Equal(t, `"12"`, marshalInt64ID(12))
	assert.Equal(t, `"0"`, marshalInt64ID(0))
	assert.Equal(t, `"-12"`, marshalInt64ID(-12))
	assert.Equal(t, `"9223372036854775807"`, marshalInt64ID(math.MaxInt64))
	assert.Equal(t, `"-9223372036854775808"`, marshalInt64ID(math.MinInt64))
}

func TestUnmarshalInt64ID(t *testing.T) {
	tests := []struct {
		Name        string
		Input       any
		Expected    int64
		ShouldError bool
	}{
		{
			Name:     "string",
			Input:    "12",
			Expected: 12,
		},
		{
			Name:     "string negative",
			Input:    "-12",
			Expected: -12,
		},
		{
			Name:     "string int64 max",
			Input:    "9223372036854775807",
			Expected: math.MaxInt64,
		},
		{
			Name:     "string int64 min",
			Input:    "-9223372036854775808",
			Expected: math.MinInt64,
		},
		{
			Name:     "int",
			Input:    12,
			Expected: 12,
		},
		{
			Name:     "int64",
			Input:    int64(12),
			Expected: 12,
		},
		{
			Name:     "json.Number",
			Input:    json.Number("12"),
			Expected: 12,
		},
		{
			Name:     "json.Number int64 max",
			Input:    json.Number("9223372036854775807"),
			Expected: math.MaxInt64,
		},
		{
			Name:        "string invalid",
			Input:       "abc",
			Expected:    0,
			ShouldError: true,
		},
		{
			Name:        "string float",
			Input:       "1.2",
			Expected:    0,
			ShouldError: true,
		},
		{
			Name:        "string overflow",
			Input:       "9223372036854775808",
			Expected:    0,
			ShouldError: true,
		},
		{
			Name:        "string empty",
			Input:       "",
			Expected:    0,
			ShouldError: true,
		},
		{
			Name:        "unsupported type",
			Input:       struct{}{},
			Expected:    0,
			ShouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			id, err := UnmarshalInt64ID(tt.Input)

			if tt.ShouldError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.Expected, id)
		})
	}
}
