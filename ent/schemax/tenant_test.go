package schemax

import (
	"context"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/assert"
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

func TestGetTenantRules(t *testing.T) {
	selector := sql.Select("tenant_id").From(sql.Table("test_table"))
	mixin := TenantMixin[MockQuery, MockMutator]{storageKey: "tenant_id"}

	tests := []struct {
		name    string
		filers  []string
		tid     string
		wantLen int
	}{
		{
			name:    "empty filter",
			filers:  []string{""},
			tid:     "123",
			wantLen: 1,
		},
		{
			name:    "single field filter",
			filers:  []string{"foo/1"},
			tid:     "123",
			wantLen: 1,
		},
		{
			name:    "multi field filter",
			filers:  []string{"foo/1:bar/2"},
			tid:     "123",
			wantLen: 1,
		},
		{
			name:    "invalid filter",
			filers:  []string{"foo"},
			tid:     "123",
			wantLen: 1,
		},
		{
			name:    "all empty filters and tenant ID",
			filers:  []string{},
			tid:     "",
			wantLen: 1,
		},
		{
			name:    `empty tenant ID should tenant_id == ""`,
			filers:  []string{"foo/1"},
			tid:     "",
			wantLen: 1,
		},
		{
			name:    "attach tenant",
			filers:  []string{"", "tenant_id/123"},
			tid:     "123",
			wantLen: 1,
		},
		{
			name:    "attach tenant",
			filers:  []string{"", "tenant_id/[345,678]"},
			tid:     "123",
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mixin.getTenantRules(tt.filers, tt.tid, selector)
			assert.Len(t, got, tt.wantLen)
		})
	}
}
