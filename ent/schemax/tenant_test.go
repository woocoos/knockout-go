package schemax

import (
	"context"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockQuery struct{}
type MockMutator struct{}

func (mq MockQuery) WhereP(...func(selector *sql.Selector)) {}
func (mq MockQuery) Client() MockMutator                    { return MockMutator{} }
func (mq MockQuery) SetTenantID(int)                        {}

func NewMockQuery(ent.Query) (MockQuery, error) {
	return MockQuery{}, nil
}

func (mm MockMutator) Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error) {
	return nil, nil
}

func TestNewTenantMixin(t *testing.T) {
	app := "testApp"
	mixin := NewTenantMixin[MockQuery, MockMutator](app, NewMockQuery)
	assert.Equal(t, app, mixin.app)
	assert.NotNil(t, mixin.newQueryFunc)

	mixin = NewTenantMixin[MockQuery, MockMutator](app, NewMockQuery,
		WithTenantMixinStorageKey[MockQuery, MockMutator]("org_id"))
	assert.Equal(t, "org_id", mixin.storageKey)

}
