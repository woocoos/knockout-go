package identity

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web"
	"github.com/tsingsun/woocoo/web/handler"
	"net/http"
	"strconv"
)

var (
	tenantContextKey = "_woocoos/knockout/tenant_id"
	TenantHeaderKey  = "X-Tenant-ID"

	ErrMisTenantID = errors.New("miss tenant id")
)

type TenantOptions struct {
	Lookup     string
	RootDomain string
	Exclude    []string
	Skipper    handler.Skipper
}

// TenantIDMiddleware returns middleware to get tenant id from http request
func TenantIDMiddleware(cfg *conf.Configuration) gin.HandlerFunc {
	opts := TenantOptions{
		Lookup: "header:" + TenantHeaderKey,
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
		c.Set(tenantContextKey, v)
	}
}

// RegistryTenantIDMiddleware register a middleware to get tenant id from request header
func RegistryTenantIDMiddleware() web.Option {
	return web.WithMiddlewareApplyFunc("tenant", TenantIDMiddleware)
}

func WithTenantID(parent context.Context, id int) context.Context {
	return context.WithValue(parent, tenantContextKey, id)
}

// TenantIDFromContext returns the tenant id from context.tenant id is int format
func TenantIDFromContext(ctx context.Context) (id int, err error) {
	var tid any
	ginCtx, ok := ctx.Value(gin.ContextKey).(*gin.Context)
	if ok {
		tid = ginCtx.Value(tenantContextKey)
	} else {
		tid = ctx.Value(tenantContextKey)
	}

	switch tid.(type) {
	case int:
		return tid.(int), nil
	case string:
		id, err = strconv.Atoi(tid.(string))
		if err == nil {
			return
		}
	}
	return 0, fmt.Errorf("invalid tenant id type %T", tid)
}
