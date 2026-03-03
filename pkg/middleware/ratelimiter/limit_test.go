package ratelimiter

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/cache/redisc"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/tsingsun/woocoo/web"
	"github.com/woocoos/knockout-go/pkg/identity"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Note: TestKeyFuncUser and TestKeyFuncTenant are tested through TestHandlerFunc_MemoryStore
// and TestHandlerFunc_MemoryStoreWithTenantKey because gin.CreateTestContext doesn't
// support ContextWithFallback which is required for reading from request context.

func TestGetKeyFunc(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
		notNil  bool
	}{
		{name: "user", keyName: "user", notNil: true},
		{name: "tenant", keyName: "tenant", notNil: true},
		{name: "ip", keyName: "ip", notNil: false}, // ip is not a valid keyFunc, will use default in RateLimiter
		{name: "unknown", keyName: "unknown", notNil: false},
		{name: "empty", keyName: "", notNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getKeyFunc(tt.keyName)
			if tt.notNil {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestHandlerFunc_MemoryStore(t *testing.T) {
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 1
keyFunc: user
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", func(c *gin.Context) {
		ctx := security.WithContext(c.Request.Context(), security.NewGenericPrincipalByClaims(
			jwt.MapClaims{"sub": "testuser"}))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}, h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)

	// Wait for rate limit to reset
	time.Sleep(time.Second + 100*time.Millisecond)

	// Third request after reset should succeed
	req3 := httptest.NewRequest("GET", "/", nil)
	w3 := httptest.NewRecorder()
	srv.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
}

func TestHandlerFunc_MemoryStoreWithExclude(t *testing.T) {
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 1
keyFunc: ip
exclude:
  - /health
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/test", h, func(c *gin.Context) {
		c.String(200, "ok")
	})
	srv.GET("/health", h, func(c *gin.Context) {
		c.String(200, "healthy")
	})

	// First /test request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second /test request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)

	// /health requests should not be rate limited (excluded)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "health check %d should not be rate limited", i)
	}
}

func TestHandlerFunc_MemoryStoreWithTenantKey(t *testing.T) {
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 2
keyFunc: tenant
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", func(c *gin.Context) {
		tidStr := c.Query("tenant_id")
		tid, _ := strconv.Atoi(tidStr)
		ctx := identity.WithTenantID(c.Request.Context(), tid)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}, h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Tenant 1: first request
	req1 := httptest.NewRequest("GET", "/?tenant_id=1", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Tenant 1: second request
	req2 := httptest.NewRequest("GET", "/?tenant_id=1", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// Tenant 1: third request should be rate limited
	req3 := httptest.NewRequest("GET", "/?tenant_id=1", nil)
	w3 := httptest.NewRecorder()
	srv.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)

	// Tenant 2: should not be affected by tenant 1's limit
	req4 := httptest.NewRequest("GET", "/?tenant_id=2", nil)
	w4 := httptest.NewRecorder()
	srv.ServeHTTP(w4, req4)
	assert.Equal(t, 200, w4.Code)
}

func TestHandlerFunc_MemoryStoreWithIPKey(t *testing.T) {
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 1
keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
}

func TestHandlerFunc_DefaultStore(t *testing.T) {
	// Test that default store is memory when store is not specified
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 1
keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
}

func TestHandlerFunc_RedisStore(t *testing.T) {
	// Create miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Use unique driver name to avoid conflicts
	driverName := "test-redis-store-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create redisc with the client using WithRedisClient option
	rc, err := redisc.New(conf.NewFromStringMap(map[string]any{
		"driverName": driverName,
	}), redisc.WithRedisClient(client))
	require.NoError(t, err)

	err = cache.RegisterCache(driverName, rc)
	// Ignore already registered error
	if err != nil {
		require.Contains(t, err.Error(), "already registered")
	}

	cfgstr := `
redisOptions:
  rate: 1s
  limit: 1
storeKey: ` + driverName + `
keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
}

func TestHandlerFunc_RedisStoreClientNotFound(t *testing.T) {
	cfgstr := `
redisOptions:
  rate: 1s
  limit: 1
storeKey: non-existent-cache-redis
keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgstr))

	assert.Panics(t, func() {
		mid := &Config{}
		mid.ApplyFunc(cfg)
	})
}

func TestGetRedisClientFromCache(t *testing.T) {
	t.Run("cache not found", func(t *testing.T) {
		_, err := getRedisClientFromCache("non-existent-cache-get")
		assert.Error(t, err)
	})
}

func TestKeySkip(t *testing.T) {
	t.Run("includeKeys only - key in list", func(t *testing.T) {
		mid := &Config{
			IncludeKeys: []string{"user1", "user2"},
		}
		mid.includeMap = make(map[string]struct{})
		for _, k := range mid.IncludeKeys {
			mid.includeMap[k] = struct{}{}
		}

		// Key in include list should NOT skip
		assert.False(t, mid.KeySkip("user1"))
		assert.False(t, mid.KeySkip("user2"))
	})

	t.Run("includeKeys only - key not in list", func(t *testing.T) {
		mid := &Config{
			IncludeKeys: []string{"user1", "user2"},
		}
		mid.includeMap = make(map[string]struct{})
		for _, k := range mid.IncludeKeys {
			mid.includeMap[k] = struct{}{}
		}

		// Key not in include list should skip
		assert.True(t, mid.KeySkip("otheruser"))
	})

	t.Run("excludeKeys only", func(t *testing.T) {
		mid := &Config{
			ExcludeKeys: []string{"admin", "system"},
		}
		mid.excludeMap = make(map[string]struct{})
		for _, k := range mid.ExcludeKeys {
			mid.excludeMap[k] = struct{}{}
		}

		// Key in exclude list should skip
		assert.True(t, mid.KeySkip("admin"))
		assert.True(t, mid.KeySkip("system"))
		// Key not in exclude list should NOT skip
		assert.False(t, mid.KeySkip("normaluser"))
	})

	t.Run("empty key", func(t *testing.T) {
		mid := &Config{}
		// Empty key should skip
		assert.True(t, mid.KeySkip(""))
	})

	t.Run("no skip lists", func(t *testing.T) {
		mid := &Config{}
		// Should not skip (apply rate limiting)
		assert.False(t, mid.KeySkip("anyuser"))
	})

	t.Run("excludeKeys has higher priority", func(t *testing.T) {
		mid := &Config{
			IncludeKeys: []string{"admin", "user1"},
			ExcludeKeys: []string{"admin"},
		}
		mid.includeMap = make(map[string]struct{})
		for _, k := range mid.IncludeKeys {
			mid.includeMap[k] = struct{}{}
		}
		mid.excludeMap = make(map[string]struct{})
		for _, k := range mid.ExcludeKeys {
			mid.excludeMap[k] = struct{}{}
		}

		// admin is in both include and exclude, exclude wins
		assert.True(t, mid.KeySkip("admin"))
		// user1 is only in include, should NOT skip
		assert.False(t, mid.KeySkip("user1"))
		// otheruser is not in include, should skip
		assert.True(t, mid.KeySkip("otheruser"))
	})
}

func TestHandlerFunc_IncludeKeys(t *testing.T) {
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 1
keyFunc: user
includeKeys:
  - limiteduser
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", func(c *gin.Context) {
		user := c.Query("user")
		ctx := security.WithContext(c.Request.Context(), security.NewGenericPrincipalByClaims(
			jwt.MapClaims{"sub": user}))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}, h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// limiteduser: first request should succeed
	req1 := httptest.NewRequest("GET", "/?user=limiteduser", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// limiteduser: second request should be rate limited
	req2 := httptest.NewRequest("GET", "/?user=limiteduser", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)

	// otheruser: should NOT be rate limited (not in includeKeys)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/?user=otheruser", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "otheruser request %d should not be rate limited", i)
	}
}

func TestHandlerFunc_ExcludeKeys(t *testing.T) {
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 1
keyFunc: user
excludeKeys:
  - admin
  - system
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", func(c *gin.Context) {
		user := c.Query("user")
		ctx := security.WithContext(c.Request.Context(), security.NewGenericPrincipalByClaims(
			jwt.MapClaims{"sub": user}))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}, h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// normaluser: first request should succeed
	req1 := httptest.NewRequest("GET", "/?user=normaluser", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// normaluser: second request should be rate limited
	req2 := httptest.NewRequest("GET", "/?user=normaluser", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)

	// admin: should NOT be rate limited (in excludeKeys)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/?user=admin", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "admin request %d should not be rate limited", i)
	}

	// system: should NOT be rate limited (in excludeKeys)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/?user=system", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "system request %d should not be rate limited", i)
	}
}

func TestHandlerFunc_IncludeAndExcludeKeys(t *testing.T) {
	// Test that excludeKeys has higher priority than includeKeys
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 1
keyFunc: user
includeKeys:
  - user1
  - user2
  - admin
excludeKeys:
  - admin
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)
	assert.NotNil(t, h)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", func(c *gin.Context) {
		user := c.Query("user")
		ctx := security.WithContext(c.Request.Context(), security.NewGenericPrincipalByClaims(
			jwt.MapClaims{"sub": user}))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}, h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// user1: first request should succeed (in includeKeys, not in excludeKeys)
	req1 := httptest.NewRequest("GET", "/?user=user1", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// user1: second request should be rate limited
	req2 := httptest.NewRequest("GET", "/?user=user1", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)

	// admin: should NOT be rate limited (in excludeKeys, even though in includeKeys)
	// excludeKeys has higher priority
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/?user=admin", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "admin request %d should not be rate limited (excludeKeys priority)", i)
	}

	// otheruser: should NOT be rate limited (not in includeKeys)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/?user=otheruser", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "otheruser request %d should not be rate limited", i)
	}
}

func TestRegisterMiddleware(t *testing.T) {
	result := RegisterMiddleware()
	cfgStr := `
server:
  addr: 127.0.0.1:0
engine:
  routerGroups:
    - default:
        middlewares:
          - rateLimit:
              inMemoryOptions:
                rate: 1s
                limit: 1
              keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgStr))
	webSrv := web.New(web.WithConfiguration(cfg), result)
	assert.NotNil(t, result)
	assert.NotNil(t, webSrv)

	// Add a test route
	webSrv.Router().Engine.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	webSrv.Router().Engine.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	webSrv.Router().Engine.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
}
