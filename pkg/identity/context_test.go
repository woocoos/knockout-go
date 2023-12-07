package identity

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTenantIDFromContext(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		tid, ok := TenantIDFromContext(context.Background())
		assert.False(t, ok)
		assert.Equal(t, 0, tid)
	})
	t.Run("int", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), 1)
		tid, ok := TenantIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, 1, tid)
	})
	t.Run("string", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), TenantContextKey, "1")
		_, ok := TenantIDFromContext(ctx)
		assert.False(t, ok)
	})
}
