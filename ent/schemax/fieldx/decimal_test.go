package fieldx

import (
	"entgo.io/contrib/entgql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecimal(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		got := Decimal("test")
		assert.False(t, got.desc.Nillable)
		assert.False(t, got.desc.Immutable)
		assert.False(t, got.desc.Unique)
		assert.False(t, got.desc.Optional)
		assert.Equal(t, "test", got.desc.Name)
		assert.Equal(t, "string", got.desc.Info.Type.String())
		assert.Nil(t, got.desc.ValueScanner)
		assert.Nil(t, got.desc.Validators)
		assert.Nil(t, got.desc.Err)
		assert.Nil(t, got.desc.Default)
		assert.Nil(t, got.desc.UpdateDefault)
		assert.Equal(t, entgql.Type("Decimal"), got.desc.Annotations[0])
		assert.Empty(t, got.desc.Comment)
		assert.Empty(t, got.desc.StorageKey)
		assert.Empty(t, got.desc.Tag)
		assert.Nil(t, got.desc.SchemaType)
		assert.Nil(t, got.desc.Enums)
		assert.Empty(t, got.desc.DeprecatedReason)
		assert.False(t, got.desc.Deprecated)
		assert.Zero(t, got.desc.Size)
		assert.False(t, got.desc.Sensitive)
	})
}
