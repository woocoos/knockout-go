package schemax

import (
	"context"
	"errors"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/assert"
	"github.com/woocoos/knockout-go/pkg/identity"
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
	assert.Equal(t, "org_id", mixin.tenantStorageKey)

}

func TestGetTenantRules(t *testing.T) {
	selector := sql.Select("tenant_id").From(sql.Table("test_table"))
	mixin := TenantMixin[MockQuery, MockMutator]{tenantStorageKey: "tenant_id"}

	tests := []struct {
		name    string
		filers  []string
		tid     int
		wantLen int
	}{
		{
			name:    "empty filter",
			filers:  []string{""},
			tid:     123,
			wantLen: 1,
		},
		{
			name:    "single field filter",
			filers:  []string{"foo/1"},
			tid:     123,
			wantLen: 1,
		},
		{
			name:    "multi field filter",
			filers:  []string{"foo/1:bar/2"},
			tid:     123,
			wantLen: 1,
		},
		{
			name:    "invalid filter",
			filers:  []string{"foo"},
			tid:     123,
			wantLen: 1,
		},
		{
			name:    "all empty filters and tenant ID",
			filers:  []string{},
			tid:     0,
			wantLen: 1,
		},
		{
			name:    `empty tenant ID should tenant_id == ""`,
			filers:  []string{"foo/1"},
			tid:     0,
			wantLen: 1,
		},
		{
			name:    "attach tenant",
			filers:  []string{"", "tenant_id/123"},
			tid:     0,
			wantLen: 1,
		},
		{
			name:    "attach tenant",
			filers:  []string{"", "tenant_id/[345,678]"},
			tid:     123,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mixin.getTenantRules(tt.filers, tt.tid, 0, selector)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

// 跟踪WhereP是否被调用的mock
type trackingQuery struct {
	MockQuery
	wherePCalled bool
	wherePCount  int
}

func (tq *trackingQuery) WhereP(fns ...func(selector *sql.Selector)) {
	tq.wherePCalled = true
	tq.wherePCount += len(fns)
}

func TestTenantOnlyInterceptors(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		domainKey  string
		wantErr    error
		wantWhereP bool
	}{
		{
			name:       "跳过租户隐私检查",
			ctx:        SkipTenantPrivacy(context.Background()),
			wantErr:    nil,
			wantWhereP: false,
		},
		{
			name:    "缺少租户ID",
			ctx:     context.Background(),
			wantErr: identity.ErrMisTenantID,
		},
		{
			name:       "正常情况-无domain",
			ctx:        identity.WithTenantID(context.Background(), 123),
			wantErr:    nil,
			wantWhereP: true,
		},
		{
			name:       "正常情况-有domain",
			ctx:        identity.WithDomainID(identity.WithTenantID(context.Background(), 123), 456),
			domainKey:  "domain_id",
			wantErr:    nil,
			wantWhereP: true,
		},
		{
			name:      "有domain配置但缺少domainID",
			ctx:       identity.WithTenantID(context.Background(), 123),
			domainKey: "domain_id",
			wantErr:   identity.ErrMissDomainTenantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 用于捕获newQueryFunc返回的查询对象
			var capturedQuery *trackingQuery

			mixin := NewTenantMixin[*trackingQuery, MockMutator]("testApp",
				func(q ent.Query) (*trackingQuery, error) {
					capturedQuery = &trackingQuery{}
					return capturedQuery, nil
				},
				WithTenantMixinDomainStorageKey[*trackingQuery, MockMutator](tt.domainKey),
			)

			interceptors := mixin.TenantOnlyInterceptors()
			assert.Len(t, interceptors, 1)

			// 获取TraverseFunc并执行
			traverseFunc, ok := interceptors[0].(ent.TraverseFunc)
			assert.True(t, ok)

			err := traverseFunc.Traverse(tt.ctx, &trackingQuery{})

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}

			if tt.wantWhereP {
				assert.NotNil(t, capturedQuery)
				assert.Equal(t, tt.wantWhereP, capturedQuery.wherePCalled)
			}
		})
	}
}

func TestTenantOnlyInterceptors_PredicateLogic(t *testing.T) {
	tests := []struct {
		name      string
		tid       int
		did       int
		domainKey string
	}{
		{
			name:      "tid等于did时使用domain过滤",
			tid:       100,
			did:       100,
			domainKey: "domain_id",
		},
		{
			name:      "tid不等于did时使用tenant过滤",
			tid:       100,
			did:       200,
			domainKey: "domain_id",
		},
		{
			name:      "无domain配置时使用tenant过滤",
			tid:       100,
			did:       0,
			domainKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedQuery *trackingQuery

			mixin := NewTenantMixin[*trackingQuery, MockMutator]("testApp",
				func(q ent.Query) (*trackingQuery, error) {
					capturedQuery = &trackingQuery{}
					return capturedQuery, nil
				},
				WithTenantMixinDomainStorageKey[*trackingQuery, MockMutator](tt.domainKey),
			)

			ctx := identity.WithTenantID(context.Background(), tt.tid)
			if tt.domainKey != "" {
				ctx = identity.WithDomainID(ctx, tt.did)
			}

			interceptors := mixin.TenantOnlyInterceptors()
			traverseFunc := interceptors[0].(ent.TraverseFunc)

			err := traverseFunc.Traverse(ctx, &trackingQuery{})
			assert.NoError(t, err)
			assert.NotNil(t, capturedQuery)
			assert.True(t, capturedQuery.wherePCalled)
			assert.Equal(t, 1, capturedQuery.wherePCount)
		})
	}
}
