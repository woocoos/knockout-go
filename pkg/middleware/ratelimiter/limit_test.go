package ratelimiter

import (
	"net/http/httptest"
	"strconv"
	"sync"
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

func TestHandlerFunc_RedisStore_WindowReset(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	driverName := "test-redis-window-reset-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	rc, err := redisc.New(conf.NewFromStringMap(map[string]any{
		"driverName": driverName,
	}), redisc.WithRedisClient(client))
	require.NoError(t, err)
	err = cache.RegisterCache(driverName, rc)
	if err != nil {
		require.Contains(t, err.Error(), "already registered")
	}

	// rate=1s, limit=3
	cfgstr := `
redisOptions:
  rate: 1s
  limit: 3
storeKey: ` + driverName + `
keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "request %d should succeed", i+1)
	}

	// 4th request should be rate limited
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code, "request 4 should be rate limited")

	// Wait for window to expire
	time.Sleep(1100 * time.Millisecond)

	// After window expires, requests should succeed again
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		srv.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "request %d after reset should succeed", i+1)
	}
}

func TestHandlerFunc_RedisStore_NoWindowSliding(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	driverName := "test-redis-no-sliding-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	rc, err := redisc.New(conf.NewFromStringMap(map[string]any{
		"driverName": driverName,
	}), redisc.WithRedisClient(client))
	require.NoError(t, err)
	err = cache.RegisterCache(driverName, rc)
	if err != nil {
		require.Contains(t, err.Error(), "already registered")
	}

	// rate=2s, limit=3 - if window slides, 3 requests at T=0,1,2s would extend window to T=5s
	cfgstr := `
redisOptions:
  rate: 2s
  limit: 3
storeKey: ` + driverName + `
keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// T=0: request 1 succeeds
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// T=1s: request 2 succeeds (window not expired, but tokens remain)
	time.Sleep(1 * time.Second)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// T=1s: request 3 succeeds (window not expired, but tokens remain)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// T=1s: request 4 should be rate limited (all tokens consumed)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)

	// T=2s: window should expire (started at T=0, rate=2s)
	// If window was sliding, it would have been extended to T=1+2=3s
	time.Sleep(1 * time.Second)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, "window should reset at T=2s, not slide to T=3s")
}

func TestHandlerFunc_RedisStore_Concurrent(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	driverName := "test-redis-concurrent-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	rc, err := redisc.New(conf.NewFromStringMap(map[string]any{
		"driverName": driverName,
	}), redisc.WithRedisClient(client))
	require.NoError(t, err)
	err = cache.RegisterCache(driverName, rc)
	if err != nil {
		require.Contains(t, err.Error(), "already registered")
	}

	// limit=5, rate=10s (long window to ensure all requests hit same window)
	cfgstr := `
redisOptions:
  rate: 10s
  limit: 5
storeKey: ` + driverName + `
keyFunc: ip
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)

	srv := gin.New()
	srv.ContextWithFallback = true
	srv.GET("/", h, func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Launch 20 concurrent requests
	const totalRequests = 20
	var wg sync.WaitGroup
	results := make([]int, totalRequests)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			srv.ServeHTTP(w, req)
			results[idx] = w.Code
		}(i)
	}

	wg.Wait()

	// Count successes and rate limits
	successCount := 0
	rateLimitedCount := 0
	for _, code := range results {
		switch code {
		case 200:
			successCount++
		case 429:
			rateLimitedCount++
		}
	}

	// Exactly 5 should succeed, 15 should be rate limited
	assert.Equal(t, 5, successCount, "exactly 5 requests should succeed")
	assert.Equal(t, 15, rateLimitedCount, "exactly 15 requests should be rate limited")
}

func TestHandlerFunc_RedisStore_MultiUserConcurrent(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	driverName := "test-redis-multi-user-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	rc, err := redisc.New(conf.NewFromStringMap(map[string]any{
		"driverName": driverName,
	}), redisc.WithRedisClient(client))
	require.NoError(t, err)
	err = cache.RegisterCache(driverName, rc)
	if err != nil {
		require.Contains(t, err.Error(), "already registered")
	}

	// limit=3 per user, rate=10s (long window so all requests hit same window)
	cfgstr := `
redisOptions:
  rate: 10s
  limit: 3
storeKey: ` + driverName + `
keyFunc: user
`
	cfg := conf.NewFromBytes([]byte(cfgstr))
	mid := &Config{}
	h := mid.ApplyFunc(cfg)

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

	const (
		numUsers       = 5
		reqsPerUser    = 10
		limitPerUser   = 3
		totalRequests  = numUsers * reqsPerUser
	)

	var wg sync.WaitGroup
	type result struct {
		user string
		code int
	}
	results := make([]result, totalRequests)

	for u := 0; u < numUsers; u++ {
		user := "user" + strconv.Itoa(u)
		for r := 0; r < reqsPerUser; r++ {
			wg.Add(1)
			idx := u*reqsPerUser + r
			go func(i int, usr string) {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/?user="+usr, nil)
				w := httptest.NewRecorder()
				srv.ServeHTTP(w, req)
				results[i] = result{user: usr, code: w.Code}
			}(idx, user)
		}
	}

	wg.Wait()

	// Verify per-user results: each user should have exactly limitPerUser successes
	userSuccess := make(map[string]int)
	userLimited := make(map[string]int)
	for _, r := range results {
		switch r.code {
		case 200:
			userSuccess[r.user]++
		case 429:
			userLimited[r.user]++
		}
	}

	for u := 0; u < numUsers; u++ {
		user := "user" + strconv.Itoa(u)
		assert.Equal(t, limitPerUser, userSuccess[user],
			"%s: exactly %d requests should succeed", user, limitPerUser)
		assert.Equal(t, reqsPerUser-limitPerUser, userLimited[user],
			"%s: %d requests should be rate limited", user, reqsPerUser-limitPerUser)
	}
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

func TestHandlerFunc_ConcurrentRateLimit(t *testing.T) {
	// Test that concurrent requests are properly rate limited
	cfgstr := `
inMemoryOptions:
  rate: 1s
  limit: 5
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

	// Send 20 concurrent requests
	var successCount, failCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			mu.Lock()
			defer mu.Unlock()
			if w.Code == 200 {
				successCount++
			} else if w.Code == 429 {
				failCount++
			}
		}()
	}
	wg.Wait()

	// With limit=5, exactly 5 requests should succeed
	assert.Equal(t, 5, successCount, "Exactly 5 requests should succeed with limit=5")
	assert.Equal(t, 15, failCount, "15 requests should be rate limited")
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

func TestGetShard(t *testing.T) {
	sm := newShardMap()

	t.Run("空 key 返回 shard 0", func(t *testing.T) {
		shard := sm.getShard("")
		assert.Equal(t, &sm.shards[0], shard)
	})

	t.Run("相同 key 总是返回同一个 shard", func(t *testing.T) {
		keys := []string{"1001", "1234", "192.168.1.1", "alice", "admin"}
		for _, key := range keys {
			shard1 := sm.getShard(key)
			shard2 := sm.getShard(key)
			assert.Equal(t, shard1, shard2, "key %q should always map to the same shard", key)
		}
	})

	t.Run("相同前缀的用户 ID 分布到不同 shard", func(t *testing.T) {
		// 用户 ID 都以 "1" 开头，旧算法 key[0] % 16 会全部落在 shard 1
		userIDs := []string{"1001", "1234", "1999", "100", "101", "1500", "1888"}
		shards := make(map[int]bool)
		for _, id := range userIDs {
			shard := sm.getShard(id)
			for i := range sm.shards {
				if shard == &sm.shards[i] {
					shards[i] = true
					break
				}
			}
		}
		// 至少分布到 2 个以上 shard（FNV 应该能分散）
		assert.Greater(t, len(shards), 2, "user IDs with same prefix should be distributed across multiple shards")
	})

	t.Run("IP 地址分布到不同 shard", func(t *testing.T) {
		ips := []string{
			"192.168.1.1",
			"192.168.1.2",
			"192.168.1.100",
			"10.0.0.1",
			"10.0.0.2",
			"172.16.0.1",
		}
		shards := make(map[int]bool)
		for _, ip := range ips {
			shard := sm.getShard(ip)
			for i := range sm.shards {
				if shard == &sm.shards[i] {
					shards[i] = true
					break
				}
			}
		}
		// 至少分布到 2 个以上 shard
		assert.Greater(t, len(shards), 2, "IP addresses should be distributed across multiple shards")
	})

	t.Run("大量 key 均匀分布", func(t *testing.T) {
		// 生成 1000 个用户 ID，检查分布均匀性
		const totalKeys = 1000
		shardCounts := make([]int, 16)

		for i := 0; i < totalKeys; i++ {
			key := strconv.Itoa(1000 + i) // "1000", "1001", ..., "1999"
			shard := sm.getShard(key)
			for j := range sm.shards {
				if shard == &sm.shards[j] {
					shardCounts[j]++
					break
				}
			}
		}

		// 每个 shard 平均应该有 totalKeys/16 = 62.5 个 key
		// 允许一定偏差，但不应有 shard 超过平均值的 3 倍
		avgCount := float64(totalKeys) / 16
		for i, count := range shardCounts {
			assert.LessOrEqual(t, float64(count), avgCount*3,
				"shard %d has %d keys, which is too many (avg: %.1f)", i, count, avgCount)
		}

		// 所有 shard 都应该被使用到
		emptyShards := 0
		for _, count := range shardCounts {
			if count == 0 {
				emptyShards++
			}
		}
		assert.LessOrEqual(t, emptyShards, 3, "too many empty shards (%d), distribution is poor", emptyShards)
	})
}
