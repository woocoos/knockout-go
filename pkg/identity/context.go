package identity

import (
	"context"
	"errors"
	"github.com/tsingsun/woocoo/pkg/security"
)

var (
	ErrInvalidUserID = errors.New("invalid user")
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
