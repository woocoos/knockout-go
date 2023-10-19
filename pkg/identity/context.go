package identity

import (
	"context"
	"errors"
	"fmt"
	"github.com/tsingsun/woocoo/pkg/security"
	"strconv"
)

var (
	TenantContextKey = "_woocoos/knockout/tenant_id"
	TenantHeaderKey  = "X-Tenant-ID"

	ErrInvalidUserID = errors.New("invalid user")
	ErrMisTenantID   = errors.New("miss tenant id")
)

func UserIDFromContext(ctx context.Context) (int, error) {
	gp, ok := security.GenericPrincipalFromContext(ctx)
	if !ok {
		return 0, ErrInvalidUserID
	}
	id := gp.GenericIdentity.NameInt()
	if id == 0 {
		return 0, ErrInvalidUserID
	}
	return id, nil
}

func WithTenantID(parent context.Context, id int) context.Context {
	return context.WithValue(parent, TenantContextKey, id)
}

// TenantIDFromContext returns the tenant id from context.tenant id is int format
func TenantIDFromContext(ctx context.Context) (id int, err error) {
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
