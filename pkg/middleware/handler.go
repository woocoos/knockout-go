package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web"
	"github.com/tsingsun/woocoo/web/handler"
	"github.com/tsingsun/woocoo/web/handler/signer"
	"github.com/woocoos/entcache"
	"github.com/woocoos/knockout-go/pkg/identity"
	"net/http"
	"strconv"
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
	return web.WithMiddlewareApplyFunc("cachectl", CacheControlMiddleware)
}

type CacheControlConfig struct {
	Lookup     string
	RootDomain string
	Exclude    []string
	Skipper    handler.Skipper
}

// CacheControlMiddleware returns middleware to set skip cache from request header
func CacheControlMiddleware(cfg *conf.Configuration) gin.HandlerFunc {
	// 更简单的方式，但不确定是否好
	//return func(c *gin.Context) {
	//	ctx := handler.GetDerivativeContext(c)
	//	req := graphql.GetOperationContext(ctx)
	//	cacheControl := req.Headers.Get("Cache-Control")
	//	if cacheControl == "no-cache" {
	//		ctx = entcache.Skip(ctx)
	//		handler.SetDerivativeContext(c, ctx)
	//	}
	//}

	opts := CacheControlConfig{
		Lookup: "header:" + "Cache-Control",
	}
	if err := cfg.Unmarshal(&opts); err != nil {
		panic(err)
	}
	if opts.Skipper == nil {
		opts.Skipper = handler.PathSkipper(opts.Exclude)
	}
	var findValue func(c *gin.Context) (string, error)
	switch opts.Lookup {
	case "host":
		findValue = func(c *gin.Context) (str string, err error) {
			host := c.Request.Host
			if len(opts.RootDomain) > 0 {
				str = host[:len(host)-len(opts.RootDomain)-1]
			}
			return
		}
	default:
		findValue = func(c *gin.Context) (str string, err error) {
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
		cacheControl, err := findValue(c)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("get cache control error: %v", err))
			return
		}
		if cacheControl == "no-cache" {
			ctx := handler.GetDerivativeContext(c)
			ctx = entcache.Skip(ctx)
			handler.SetDerivativeContext(c, ctx)
		}
	}
}
