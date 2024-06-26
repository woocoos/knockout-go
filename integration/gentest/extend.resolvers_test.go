package gentest

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_userResolver_IsExtend(t *testing.T) {
	ur := userResolver{}
	assert.NotPanics(t, func() {
		ur.IsExtend(nil, nil)
	})
}

func Test_queryResolver_User(t *testing.T) {
	ur := queryResolver{}
	assert.NotPanics(t, func() {
		u, err := ur.User(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, u.ID)
	})
}
