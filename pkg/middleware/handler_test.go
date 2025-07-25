package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web"
	"github.com/woocoos/knockout-go/pkg/identity"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
