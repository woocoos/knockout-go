package fmterr

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/tsingsun/woocoo/pkg/conf"
	"testing"
)

func TestParseCodeMap(t *testing.T) {
	t.Run("empty conf", func(t *testing.T) {
		err := ParseCodeMap(conf.New())
		assert.NoError(t, err)
		assert.Empty(t, codeMap)
	})
	t.Run("normal conf", func(t *testing.T) {
		err := InitErrorHandler(conf.NewFromBytes([]byte(`
1000: "test"
`)))
		assert.NoError(t, err)
		assert.Equal(t, "test", codeMap[1000])
	})
}

func TestNew(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		err := New(1000, errors.New("test"))
		assert.NotNil(t, err.Err)
		assert.ErrorContains(t, err.Err, "test")
	})
	t.Run("newf", func(t *testing.T) {
		err := Newf(1000, "test %s", "string")
		assert.NotNil(t, err.Err)
		assert.ErrorContains(t, err.Err, "test string")
	})
}

func TestCodef(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		err := Codef(1000, "test")
		assert.NotNil(t, err.Err)
		assert.ErrorContains(t, err.Err, "test")
	})
	t.Run("args", func(t *testing.T) {
		err := Codef(1000, "test", "test2", "test3")
		assert.Len(t, err.Meta, 1)
		assert.Equal(t, "test2", err.Meta.(map[string]any)["test"])
		err = Codef(1000, "test", "test2", "test3", "test4")
		assert.Len(t, err.Meta, 2)
		assert.Equal(t, "test2", err.Meta.(map[string]any)["test"])
		assert.Equal(t, "test4", err.Meta.(map[string]any)["test3"])
	})
}
