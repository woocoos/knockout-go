package casbin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/alicebob/miniredis/v2"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
	rediswatcher "github.com/casbin/redis-watcher/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/contrib/gql"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/log"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/tsingsun/woocoo/test/wctest"
	"github.com/tsingsun/woocoo/web"
	"github.com/tsingsun/woocoo/web/handler"
	"github.com/tsingsun/woocoo/web/handler/authz"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/woocoos/knockout-go/pkg/identity"
	"github.com/woocoos/knockout-go/test"
)

func casbinFilePrepare(node string) {
	p, err := conf.NewParserFromFile("testdata/casbin.yaml")
	if err != nil {
		panic(err)
	}
	cfg := conf.NewFromParse(p)
	if err := os.WriteFile(test.Tmp(node+`_policy.csv`), []byte(cfg.String(node+".policy")), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.WriteFile(test.Tmp(node+`_model.conf`), []byte(cfg.String(node+".model")), os.ModePerm); err != nil {
		panic(err)
	}
}

func TestNewAuthorization(t *testing.T) {
	type args struct {
		cnf  *conf.Configuration
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		check   func(t *testing.T, got *Authorizer)
	}{
		{
			name: "RBAC",
			args: args{
				cnf: func() *conf.Configuration {
					casbinFilePrepare("rbac")
					return conf.NewFromStringMap(map[string]any{
						"autoSave": true,
						"model":    test.Tmp(`rbac_model.conf`),
						"policy":   test.Tmp(`rbac_policy.csv`),
						"cache": map[string]any{
							"size": 1000,
							"ttl":  "1h",
						},
					})
				}(),
				opts: []Option{},
			},
			wantErr: false,
			check: func(t *testing.T, got *Authorizer) {
				assert.NoError(t, got.Enforcer.LoadPolicy())
				_, err := got.Enforcer.AddPermissionForUser("alice", "data1", "write")
				require.NoError(t, err)
				has, err := got.Enforcer.Enforce("alice", "data1", "write")
				require.NoError(t, err)
				assert.True(t, has)
				assert.NoError(t, got.Enforcer.SavePolicy())
			},
		},
		{
			name: "redis watcher",
			args: args{
				cnf: func() *conf.Configuration {
					casbinFilePrepare("redis")
					mr := miniredis.RunT(t)
					return conf.NewFromStringMap(map[string]any{
						"expireTime": 10 * time.Second,
						"watcherOptions": map[string]any{
							"options": map[string]any{
								"addr": mr.Addr(),
							},
						},
						"model":  test.Tmp(`redis_model.conf`),
						"policy": test.Tmp(`redis_policy.csv`),
					})
				}(),
			},
			wantErr: false,
			check: func(t *testing.T, got *Authorizer) {
				defer got.Watcher.Close()
				assert.NoError(t, got.Enforcer.LoadPolicy())
				_, err := got.Enforcer.AddPermissionForUser("alice", "data1", "write")
				require.NoError(t, err)
				has, err := got.Enforcer.Enforce("alice", "data1", "write")
				require.NoError(t, err)
				assert.True(t, has)
				has, err = got.Eval(context.Background(), &security.EvalArgs{
					User: security.NewGenericPrincipalByClaims(jwt.MapClaims{
						"sub": "alice",
					}),
					Action:     "data1",
					ActionVerb: "write",
				})
				require.NoError(t, err)
				assert.True(t, has)
				assert.NoError(t, got.Enforcer.SavePolicy())
			},
		},
		{
			name: "redis watcher without redis instance",
			args: args{
				cnf: func() *conf.Configuration {
					casbinFilePrepare("rbac")
					return conf.NewFromStringMap(map[string]any{
						"expireTime": 10 * time.Second,
						"watcherOptions": map[string]any{
							"options": map[string]any{
								"addr": "wrong addr",
							},
						},
						"model":  test.Tmp(`redis_model.conf`),
						"policy": test.Tmp(`redis_policy.csv`),
					})
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAuthorizer(tt.args.cnf, tt.args.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			tt.check(t, got)
		})
	}
}

func TestAuthorizer(t *testing.T) {

	var cnf = `
authz:
  autoSave: false
  model: |
    [request_definition]
    r = sub, obj, act
    [policy_definition]
    p = sub, obj, act
    [role_definition]
    g = _, _
    [policy_effect]
    e = some(where (p.eft == allow))
    [matchers]
    %s
  policy: "p, 1, test:/, read"
handler:
  appCode: "test"
  ConfigPath: "authz"
`

	gin.SetMode(gin.ReleaseMode)
	tcnf := fmt.Sprintf(cnf, "m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")
	tests := []struct {
		name  string
		cfg   *conf.Configuration
		req   *http.Request
		check func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "use global",
			cfg: func() *conf.Configuration {
				cfg := conf.NewFromBytes([]byte(tcnf))
				cfg.Parser().Set("handler", map[string]any{
					"appCode":    "test",
					"configPath": "",
				})
				return cfg
			}(),
			req: httptest.NewRequest("GET", "/", nil).
				WithContext(security.WithContext(context.Background(),
					security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "1"}))),
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "pass",
			cfg:  conf.NewFromBytes([]byte(tcnf)),
			req: httptest.NewRequest("GET", "/", nil).
				WithContext(security.WithContext(context.Background(),
					security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "1"}))),
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "no pass",
			cfg:  conf.NewFromBytes([]byte(tcnf)),
			req: httptest.NewRequest("GET", "/unauth", nil).
				WithContext(security.WithContext(context.Background(),
					security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "1"}))),
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusForbidden, w.Code)
			},
		},
		{
			name: "match error",
			cfg: func() *conf.Configuration {
				nm := fmt.Sprintf(cnf, "m = g(r.sub, p.sub) && r.obj1 == p.obj && r.act != p.act")
				cfg := conf.NewFromBytes([]byte(nm))
				return cfg
			}(),
			req: httptest.NewRequest("GET", "/", nil).
				WithContext(security.WithContext(context.Background(),
					security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "1"}))),
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusForbidden, w.Code)
			},
		},
		{
			name: "miss user",
			cfg: func() *conf.Configuration {
				nm := fmt.Sprintf(cnf, "m = g(r.sub, p.sub) && r.obj1 == p.obj && r.act != p.act")
				cfg := conf.NewFromBytes([]byte(nm))
				return cfg
			}(),
			req: httptest.NewRequest("GET", "/", nil),
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusForbidden, w.Code)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer, err := NewAuthorizer(tt.cfg.Sub("authz"))
			require.NoError(t, err)
			security.SetDefaultAuthorizer(authorizer)

			got := authz.Middleware()
			h := handler.NewSimpleMiddleware("authz", got.ApplyFunc)
			w := httptest.NewRecorder()
			_, e := gin.CreateTestContext(w)
			e.ContextWithFallback = true
			e.Use(h.ApplyFunc(tt.cfg.Sub("handler")))
			e.GET("/", func(c *gin.Context) {
				c.String(200, "ok")
			})
			e.GET("/unauth", func(c *gin.Context) {
				c.String(200, "ok")
			})

			e.ServeHTTP(w, tt.req)
			tt.check(t, w)
		})
	}
}

func TestRedisCallback(t *testing.T) {
	casbinFilePrepare("callback")
	redis := miniredis.RunT(t)
	authorizer, err := NewAuthorizer(conf.NewFromStringMap(map[string]any{
		"expireTime": 10 * time.Second,
		"watcherOptions": map[string]any{
			"options": map[string]any{
				"addr":    redis.Addr(),
				"channel": "/casbin",
			},
		},
		"model":  test.Tmp(`callback_model.conf`),
		"policy": test.Tmp(`callback_policy.csv`),
	}))

	require.NoError(t, err)
	t.Run("UpdateForAddPolicy", func(t *testing.T) {
		msg := rediswatcher.MSG{ID: uuid.New().String(), Method: "UpdateForAddPolicy",
			Sec: "g", Ptype: "g", NewRule: []string{"alice", "admin"},
		}
		m, err := json.Marshal(msg)
		require.NoError(t, err)
		redis.Publish("/casbin", string(m))
		assert.NoError(t, wctest.RunWait(t.Log, time.Second*3, func() error {
			time.Sleep(time.Second * 2)
			return nil
		}))
	})
	// file adapter does not support UpdateForRemovePolicy
	t.Run("UpdateForRemovePolicy", func(t *testing.T) {
		msg := rediswatcher.MSG{ID: uuid.New().String(), Method: "UpdateForRemovePolicy",
			Sec: "p", Ptype: "p", NewRule: []string{"alice", "data1", "remove"},
		}
		m, err := json.Marshal(msg)
		require.NoError(t, err)
		ok, err := authorizer.Enforcer.HasPolicy("alice", "data1", "remove")
		require.NoError(t, err)
		assert.True(t, ok)
		redis.Publish("/casbin", string(m))
		assert.NoError(t, wctest.RunWait(t.Log, time.Second*3, func() error {
			time.Sleep(time.Second * 2)
			return nil
		}))
	})
	authorizer.Watcher.Close()
}

func TestGraphqlCheckPermissions(t *testing.T) {
	log.InitGlobalLogger()
	var cfgStr = `
authz:
  autoSave: false
  model: |
    [request_definition]
    r = sub, obj, act
    [policy_definition]
    p = sub, obj, act
    [role_definition]
    g = _, _
    [policy_effect]
    e = some(where (p.eft == allow))
    [matchers]
    m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act

web:
  server:
  engine:
    routerGroups:
      - default:
          middlewares:
            - graphql:
                withAuthorization: true
`
	cfg := conf.NewFromBytes([]byte(cfgStr)).AsGlobal()
	auth, err := NewAuthorizer(cfg.Sub("authz"), WithAdapter(stringadapter.NewAdapter(`p, 1, :hello, read`)))
	require.NoError(t, err)
	security.SetDefaultAuthorizer(auth)
	srv := web.New(web.WithConfiguration(cfg.Sub("web")),
		web.WithMiddlewareNewFunc("graphql", gql.Middleware))
	mock := graphql.ExecutableSchemaMock{
		ComplexityFunc: func(typeName string, fieldName string, childComplexity int, args map[string]any) (int, bool) {
			panic("mock out the Complexity method")
		},
		ExecFunc: func(ctx context.Context) graphql.ResponseHandler {
			return func(ctx context.Context) *graphql.Response {
				return &graphql.Response{
					Data: []byte("{}"),
				}
			}
		},
		SchemaFunc: func() *ast.Schema {
			return &ast.Schema{
				Query: &ast.Definition{
					Kind: ast.Object,
					Name: "Query",
					Fields: []*ast.FieldDefinition{
						{
							Name:     "hello",
							Type:     ast.NamedType("Boolean", &ast.Position{}),
							Position: &ast.Position{},
						},
					},
				},
				Types: map[string]*ast.Definition{
					"Boolean": {
						Kind:     ast.Scalar,
						Name:     "Boolean",
						Position: &ast.Position{},
					},
				},
			}
		},
	}

	_, err = gql.RegisterSchema(srv, &mock)
	require.NoError(t, err)
	var reuqest = func(target, uid string) *http.Request {
		r := httptest.NewRequest("POST", target, bytes.NewReader([]byte(`{"query":"query hello { hello }"}`)))
		if uid != "" {
			r = r.WithContext(security.WithContext(context.Background(), security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": uid})))
		}
		r.Header.Set("Content-Type", "application/json")
		return r
	}
	t.Run("allow", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := reuqest("/query", "1")
		srv.Router().ServeHTTP(w, r)
		if !assert.Equal(t, http.StatusOK, w.Code) {
			t.Log(w.Body.String())
		}
	})
	t.Run("reject", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := reuqest("/query", "2")
		srv.Router().ServeHTTP(w, r)
		if assert.Equal(t, http.StatusForbidden, w.Code) {
			assert.Contains(t, w.Body.String(), "action hello is not allowed")
		}
	})
}

func TestAuthorizer_QueryAllowedResourceConditions(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		casbinFilePrepare("resources")
		authorizer, err := NewAuthorizer(conf.NewFromStringMap(map[string]any{
			"expireTime": 10 * time.Second,
			"model":      test.Tmp(`resources_model.conf`),
			"policy":     test.Tmp(`resources_policy.csv`),
			"cache": map[string]any{
				"size": 100,
			},
		}))
		require.NoError(t, err)
		ctx := identity.WithTenantID(context.Background(), 10000)
		ea := &security.EvalArgs{
			User:     security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "alice"}),
			Action:   "schema",
			Resource: ":10000:World",
		}
		condions, err := authorizer.QueryAllowedResourceConditions(ctx, ea)
		require.NoError(t, err)
		assert.Equal(t, []string{":name/cba:power_by/0"}, condions)
	})
	t.Run("cache", func(t *testing.T) {
		casbinFilePrepare("resources")
		authorizer, err := NewAuthorizer(conf.NewFromStringMap(map[string]any{
			"expireTime": 10 * time.Second,
			"model":      test.Tmp(`resources_model.conf`),
			"policy":     test.Tmp(`resources_policy.csv`),
			"cache": map[string]any{
				"size": 100,
			},
		}))
		require.NoError(t, err)
		ctx := identity.WithTenantID(context.Background(), 10000)
		ea := &security.EvalArgs{
			User:     security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "alice"}),
			Action:   "schema",
			Resource: ":10000:World",
		}
		condions, err := authorizer.QueryAllowedResourceConditions(ctx, ea)
		require.NoError(t, err)
		assert.Equal(t, []string{":name/cba:power_by/0"}, condions)
		// from cache
		condions, err = authorizer.QueryAllowedResourceConditions(ctx, ea)
		require.NoError(t, err)
		assert.Equal(t, []string{":name/cba:power_by/0"}, condions)
	})
}
