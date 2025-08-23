package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web"
	"github.com/woocoos/knockout-go/api/auth"
	"github.com/woocoos/knockout-go/pkg/identity"
)

type mockAuthAPI struct {
	requestCount int
}

func (m *mockAuthAPI) GetDomain(ctx context.Context, req *auth.GetDomainRequest) (ret *auth.Domain, resp *http.Response, err error) {
	m.requestCount++
	switch req.OrgID {
	case 1:
		ret = &auth.Domain{
			ParentID: 1000,
		}
	case 0:
		return nil, nil, errors.New("domain not found")
	}
	return ret, nil, nil
}

func TestTenantIDMiddleware(t *testing.T) {
	t.Run("ginContext", func(t *testing.T) {
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(TenantIDMiddleware(conf.New()))
		router.GET("/test", func(c *gin.Context) {
			tid, ok := identity.TenantIDLoadFromContext(c)
			require.True(t, ok)
			assert.Equal(t, 1, tid)
			c.String(200, "test")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
	t.Run("RequestContext", func(t *testing.T) {
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(TenantIDMiddleware(conf.New()))
		router.GET("/test", func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), gin.ContextKey, c)
			func(ctx2 context.Context) {
				tid, ok := identity.TenantIDLoadFromContext(ctx2)
				require.True(t, ok)
				assert.Equal(t, 1, tid)
			}(ctx)
			c.String(200, "test")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
	t.Run("not int", func(t *testing.T) {
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(TenantIDMiddleware(conf.New()))
		router.GET("/test", func(c *gin.Context) {
			tid, ok := identity.TenantIDLoadFromContext(c)
			assert.False(t, ok)
			assert.Equal(t, 1, tid)
			c.String(200, "test")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "1xxx")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("hots", func(t *testing.T) {
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(TenantIDMiddleware(conf.NewFromStringMap(map[string]any{
			"lookup":     "host",
			"rootDomain": "woocoo.com",
		})))
		router.GET("/test", func(c *gin.Context) {
			tid, ok := identity.TenantIDLoadFromContext(c)
			assert.True(t, ok)
			assert.Equal(t, 1, tid)
			c.String(200, "test")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Host = "1.woocoo.com"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
	t.Run("validate err", func(t *testing.T) {
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(TenantIDMiddleware(conf.New()))
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "test")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "0")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		req = httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "-1")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("web", func(t *testing.T) {
		router := web.New(
			web.WithConfiguration(conf.NewFromBytes([]byte(`
engine:
  routerGroups:
  - default:
      middlewares:
      - tenant:
`))),
			RegisterTenantID(),
		)
		router.Router().GET("/test", func(c *gin.Context) {
			tid, ok := identity.TenantIDLoadFromContext(c)
			assert.True(t, ok)
			assert.Equal(t, 1, tid)
			c.String(200, "test")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "1")
		w := httptest.NewRecorder()
		router.Router().ServeHTTP(w, req)
	})
}

func TestCacheControlMiddleware(t *testing.T) {
	t.Run("web-cache-control", func(t *testing.T) {
		router := web.New(
			web.WithConfiguration(conf.NewFromBytes([]byte(`
engine:
  routerGroups:
  - default:
      middlewares:
      - cacheControl:
`))),
			RegisterCacheControl(),
		)
		router.Router().GET("/test", func(c *gin.Context) {
			c.String(200, "test")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Cache-Control", "no-cache")
		w := httptest.NewRecorder()
		router.Router().ServeHTTP(w, req)
	})
}

func TestDomainIDMiddleware(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		t.Run("local", func(t *testing.T) {
			cfg := conf.NewFromStringMap(map[string]any{
				"Domains": map[string]int{"test.com": 1},
				"cache":   map[string]any{"size": 100, "ttl": "1m"},
			})
			assert.NotPanics(t, func() {
				DomainIDMiddleware(cfg, &mockAuthAPI{})
			})
		})
	})
	t.Run("domain in domains", func(t *testing.T) {
		cfg := conf.NewFromStringMap(map[string]any{
			"Domains": map[string]int{"test.com": 1},
		})
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(DomainIDMiddleware(cfg, &mockAuthAPI{}))
		router.GET("/test", func(c *gin.Context) {
			did, ok := identity.DomainIDLoadFromContext(c)
			assert.True(t, ok)
			assert.Equal(t, 1, did)
			c.String(200, "ok")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Host = "test.com"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	})
	t.Run("domain by remote", func(t *testing.T) {
		mockApi := &mockAuthAPI{}
		cfg := conf.NewFromStringMap(map[string]any{
			"Domains": map[string]int{"test.com": 1},
		})
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(TenantIDMiddleware(conf.New()), DomainIDMiddleware(cfg, mockApi))
		router.GET("/test", func(c *gin.Context) {
			did, ok := identity.DomainIDLoadFromContext(c)
			assert.True(t, ok)
			assert.Equal(t, 1000, did)
			c.String(200, "ok")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Host = "not-in-domains.com"
		req.Header.Add("X-Tenant-ID", "1")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		router.ServeHTTP(w, req)
		assert.Equal(t, 1, mockApi.requestCount, "use cache should not request again")
	})
	t.Run("domain get error by auth request", func(t *testing.T) {
		mockApi := &mockAuthAPI{}
		cfg := conf.NewFromStringMap(map[string]any{})
		router := gin.New()
		router.ContextWithFallback = true
		router.Use(TenantIDMiddleware(conf.New()), DomainIDMiddleware(cfg, mockApi))
		router.GET("/test", func(c *gin.Context) {
			assert.Fail(t, "should not be called")
		})
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Add("X-Tenant-ID", "0")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
