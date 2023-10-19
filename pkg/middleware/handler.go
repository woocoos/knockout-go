package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/tsingsun/woocoo/web"
	"github.com/tsingsun/woocoo/web/handler"
	"github.com/tsingsun/woocoo/web/handler/signer"
)

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
