package identity

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tsingsun/woocoo/pkg/security"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "empty",
			args:    args{ctx: context.Background()},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "int",
			args: args{ctx: security.WithContext(context.Background(), security.NewGenericPrincipalByClaims(
				jwt.MapClaims{"sub": "1"},
			))},
			want:    1,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UserIDFromContext(tt.args.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("UserIDFromContextAsInt(%v)", tt.args.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "UserIDFromContextAsInt(%v)", tt.args.ctx)
		})
	}
}

func TestTenantIDFromContext(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		tid, err := TenantIDFromContext(context.Background())
		assert.Equal(t, 0, tid)
		assert.Equal(t, ErrMisTenantID, err)
	})
	t.Run("int", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), 1)
		tid, err := TenantIDFromContext(ctx)
		assert.Equal(t, 1, tid)
		assert.Nil(t, err)
	})
	t.Run("string", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), TenantContextKey, "1")
		tid, err := TenantIDFromContext(ctx)
		assert.Equal(t, 1, tid)
		assert.Nil(t, err)
	})
	t.Run("invalid", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), TenantContextKey, struct{}{})
		tid, err := TenantIDFromContext(ctx)
		assert.Equal(t, 0, tid)
		assert.Error(t, err)
	})
}
