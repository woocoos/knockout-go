package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/cache/lfu"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web"
	"github.com/tsingsun/woocoo/web/handler"
	"github.com/tsingsun/woocoo/web/handler/signer"
	"github.com/woocoos/entcache"
	"github.com/woocoos/knockout-go/api/auth"
	"github.com/woocoos/knockout-go/pkg/identity"
)

// RegisterTokenSigner register middleware to sign request
func RegisterTokenSigner() web.Option {
	return web.WithMiddlewareNewFunc(signer.TokenSignerName, func() handler.Middleware {
		mw := signer.NewMiddleware(signer.TokenSignerName, handler.WithMiddlewareConfig(func(config any) {
			c := config.(*signer.Config)
			c.SignerConfig.UnsignedPayload = true
			c.SignerConfig.AuthScheme = "KO-HMAC-SHA1"
			c.SignerConfig.AuthHeaders = []string{"timestamp", "nonce"}
			c.SignerConfig.SignedLookups = map[string]string{
				"accessToken": "header:Authorization>Bearer",
				"timestamp":   "",
				"nonce":       "",
				"url":         "CanonicalUri",
			}
			c.Skipper = func(c *gin.Context) bool {
				if c.IsWebsocket() {
					return true
				}
				return false
			}
		}))
		return mw
	})
}

// RegisterTenantID register middleware to get tenant id from request header
func RegisterTenantID() web.Option {
	return web.WithMiddlewareApplyFunc("tenant", TenantIDMiddleware)
}

// TenantConfig is the configuration for TenantIDMiddleware
type TenantConfig struct {
	Lookup     string
	RootDomain string
	Exclude    []string
	Skipper    handler.Skipper
}

// TenantIDMiddleware returns middleware to get tenant id from http request
func TenantIDMiddleware(cfg *conf.Configuration) gin.HandlerFunc {
	opts := TenantConfig{
		Lookup: "header:" + identity.TenantHeaderKey,
	}
	if err := cfg.Unmarshal(&opts); err != nil {
		panic(err)
	}
	if opts.Skipper == nil {
		opts.Skipper = handler.PathSkipper(opts.Exclude)
	}
	var findTenantValue func(c *gin.Context) (string, error)
	switch opts.Lookup {
	case "host":
		findTenantValue = func(c *gin.Context) (str string, err error) {
			host := c.Request.Host
			if len(opts.RootDomain) > 0 {
				str = host[:len(host)-len(opts.RootDomain)-1]
			}
			return
		}
	default:
		findTenantValue = func(c *gin.Context) (str string, err error) {
			extr, err := handler.CreateExtractors(opts.Lookup, "")
			if err != nil {
				return
			}
			for _, extractor := range extr {
				ts, err := extractor(c)
				if err == nil && len(ts) != 0 {
					str = ts[0]
					break
				}
			}
			return
		}
	}
	return func(c *gin.Context) {
		if opts.Skipper(c) {
			return
		}
		tid, err := findTenantValue(c)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("get tenant id error: %v", err))
			return
		}
		v, err := strconv.Atoi(tid)
		if err != nil || v <= 0 {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid tenant id %s:%v", tid, err))
			return
		}
		handler.DerivativeContextWithValue(c, identity.TenantContextKey, v)
	}
}

// RegisterCacheControl register middleware to set skip cache from request header
func RegisterCacheControl() web.Option {
	return web.WithMiddlewareApplyFunc("cacheControl", CacheControlMiddleware)
}

type CacheControlConfig struct {
	Exclude []string
	Skipper handler.Skipper
}

// CacheControlMiddleware returns middleware to set skip cache from request header
func CacheControlMiddleware(cfg *conf.Configuration) gin.HandlerFunc {
	opts := CacheControlConfig{}
	if err := cfg.Unmarshal(&opts); err != nil {
		panic(err)
	}
	if opts.Skipper == nil {
		opts.Skipper = handler.PathSkipper(opts.Exclude)
	}
	return func(c *gin.Context) {
		if opts.Skipper(c) {
			return
		}
		cacheControl := c.GetHeader("Cache-Control")
		if cacheControl == "no-cache" {
			c.Request = c.Request.WithContext(entcache.Skip(c.Request.Context()))
		}
	}
}

type DomainRequest interface {
	GetDomain(ctx context.Context, req *auth.GetDomainRequest) (ret *auth.Domain, resp *http.Response, err error)
}

// DomainConfig is the configuration for DomainIDMiddleware
type DomainConfig struct {
	// Domains is a map of domain name to domain id
	// e.g. {"example.com": 1, "test.com": 2}
	// If the domain is not in the map, it will try to get the domain id
	// from the auth service using the tenant id from the request context.
	Domains map[string]int
	Skipper handler.Skipper
	// StoreKey is the key to get the cache from the cache manager
	// If it is empty, a default cache will be created with size 100000 and
	// ttl 5 minutes.
	// The cache will be used to store the domain id for the tenant id.
	StoreKey string `json:"storeKey" yaml:"storeKey"`
	request  DomainRequest
	cache    cache.Cache
}

// DomainIDMiddleware returns middleware to get domain id from http request
func DomainIDMiddleware(cfg *conf.Configuration, req DomainRequest) gin.HandlerFunc {
	var err error
	opts := DomainConfig{
		Domains: make(map[string]int),
		request: req,
	}
	if err := cfg.Unmarshal(&opts); err != nil {
		panic(err)
	}
	if opts.Skipper == nil {
		opts.Skipper = handler.PathSkipper(nil)
	}
	if opts.StoreKey != "" {
		if opts.cache, err = cache.GetCache(opts.StoreKey); err != nil {
			panic(err)
		}
	} else {
		size := 1000
		ttl := time.Minute * 5
		if cfg.IsSet("cache") {
			size = cfg.Int("cache.size")
			ttl = cfg.Duration("cache.ttl")
		}
		opts.cache, err = lfu.NewTinyLFU(conf.NewFromStringMap(map[string]any{
			"size": size,
			"ttl":  ttl,
		}))
		if err != nil {
			panic(err)
		}
	}
	return func(c *gin.Context) {
		if opts.Skipper(c) {
			return
		}
		host := c.Request.Host
		domainID, ok := opts.Domains[host]
		ctx := c.Request.Context()
		if !ok {
			tid, ok := identity.TenantIDLoadFromContext(ctx)
			if !ok {
				c.AbortWithError(http.StatusBadRequest, identity.ErrMisTenantID)
				return
			}
			// try to get domain id from cache
			key := "domain:" + strconv.Itoa(tid)
			err = opts.cache.Get(ctx, key, &domainID, cache.WithGetter(func(ctx context.Context, key string) (any, error) {
				ret, _, err := opts.request.GetDomain(c.Request.Context(), &auth.GetDomainRequest{
					OrgID: tid,
				})
				if err != nil {
					return nil, fmt.Errorf("get domain id from auth service error: %v", err)
				}
				return ret.ParentID, nil
			}))
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		}
		handler.DerivativeContextWithValue(c, identity.DomainContextKey, domainID)
	}
}
