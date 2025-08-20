package identity

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/tsingsun/woocoo/pkg/security"
)

const (
	TenantContextKey = "_woocoos/knockout/tenant_id"
	DomainContextKey = "_woocoos/knockout/domain_id"
	TenantHeaderKey  = "X-Tenant-ID"
	UserHeaderKey    = "X-User-ID"
)

var (
	ErrInvalidUserID      = errors.New("invalid user")
	ErrMisTenantID        = errors.New("miss tenant id")
	ErrMissDomainTenantID = errors.New("miss domain tenant id")
)

// UserIDFromContext returns the user id from context.
func UserIDFromContext(ctx context.Context) (int, error) {
	return UserIDFromContextAsInt(ctx)
}

// UserIDFromContextAsInt returns the user id from context, the context don't save Int UserID in
// context, we need transfer it from string to int
func UserIDFromContextAsInt(ctx context.Context) (int, error) {
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

// WithDomainID returns a new context with domain id
func WithDomainID(parent context.Context, id int) context.Context {
	return context.WithValue(parent, DomainContextKey, id)
}

// TenantIDFromContext returns the tenant id from context.
func TenantIDFromContext(ctx context.Context) (id int, err error) {
	if tid, ok := TenantIDLoadFromContext(ctx); ok {
		return tid, nil
	}
	switch tid := ctx.Value(TenantContextKey).(type) {
	case int:
		return tid, nil
	case string:
		id, err = strconv.Atoi(tid)
		if err == nil {
			return
		}
	case nil:
		return 0, ErrMisTenantID
	default:
		return 0, fmt.Errorf("invalid tenant id type %T", tid)
	}
	return
}

// TenantIDLoadFromContext returns the tenant id from context.
// tenant id has set by int format, this function simply returns the value.
func TenantIDLoadFromContext(ctx context.Context) (id int, ok bool) {
	id, ok = ctx.Value(TenantContextKey).(int)
	return
}

// DomainIDFromContext returns the domain id from context.
func DomainIDFromContext(ctx context.Context) (id int, err error) {
	switch did := ctx.Value(DomainContextKey).(type) {
	case int:
		return did, nil
	case string:
		id, err = strconv.Atoi(did)
		if err == nil {
			return
		}
	case nil:
		return 0, ErrMissDomainTenantID
	default:
		return 0, fmt.Errorf("invalid domain id type %T", did)
	}
	return
}

// DomainIDLoadFromContext returns the domain id from context.
// domain id has set by int format, this function simply returns the value.
func DomainIDLoadFromContext(ctx context.Context) (id int, ok bool) {
	id, ok = ctx.Value(DomainContextKey).(int)
	return
}
