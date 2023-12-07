package identity

import (
	"context"
	"errors"
	"github.com/tsingsun/woocoo/pkg/security"
	"strconv"
)

const (
	TenantContextKey = "_woocoos/knockout/tenant_id"
	TenantHeaderKey  = "X-Tenant-ID"
)

var (
	ErrInvalidUserID = errors.New("invalid user")
	ErrMisTenantID   = errors.New("miss tenant id")
)

// UserIDFromContext returns the user id from context
func UserIDFromContext(ctx context.Context) (int, error) {
	user, ok := security.FromContext(ctx)
	if !ok {
		return 0, ErrInvalidUserID
	}
	id, _ := strconv.Atoi(user.Identity().Name())
	if id == 0 {
		return 0, ErrInvalidUserID
	}
	return id, nil
}

// WithTenantID returns a new context with tenant id
func WithTenantID(parent context.Context, id int) context.Context {
	return context.WithValue(parent, TenantContextKey, id)
}

// TenantIDFromContext returns the tenant id from context.tenant id is int format
func TenantIDFromContext(ctx context.Context) (id int, ok bool) {
	id, ok = ctx.Value(TenantContextKey).(int)
	return
}
