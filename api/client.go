package api

import (
	"context"
	"fmt"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/httpx"
	"github.com/woocoos/knockout-go/pkg/identity"
	"golang.org/x/oauth2"
	"net/http"
	"strconv"
	"time"
)

const (
	PluginFS   = "fs"
	PluginMsg  = "msg"
	PluginAuth = "auth"
)

// Plugin is the knockout service client plugin interface.
type Plugin interface {
	Apply(*SDK, *conf.Configuration) error
}

// SDK is the knockout service client SDK.
type SDK struct {
	client       *http.Client
	plugins      map[string]Plugin
	signer       *httpx.Signature
	signerClient *client

	tokenSource oauth2.TokenSource
}

type client struct {
	Base   http.RoundTripper
	signer *httpx.Signature
}

func (c *client) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := c.signer.Sign(req, "", time.Now()); err != nil {
		return nil, err
	}
	return c.Base.RoundTrip(req)
}

// NewSDK creates a new SDK.
func NewSDK(cnf *conf.Configuration) (sdk *SDK, err error) {
	sdk = &SDK{
		plugins: make(map[string]Plugin),
	}
	cfg, err := httpx.NewClientConfig(cnf.Sub("client"))
	if err != nil {
		return nil, err
	}
	sdk.client, err = cfg.Client(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	sdk.tokenSource = cfg.OAuth2.GetTokenSource()

	if cnf.IsSet("signer") {
		sdk.signer, err = httpx.NewSignature(
			httpx.WithConfiguration(cnf.Sub("signer")),
			httpx.WithSigner(httpx.NewTokenSigner),
		)
		if err != nil {
			return nil, err
		}
		// wrap transport, because signer use bearer jwt
		authTransPort := sdk.client.Transport.(*oauth2.Transport)
		sdk.signerClient = &client{
			signer: sdk.signer,
			Base:   authTransPort.Base,
		}
		authTransPort.Base = sdk.signerClient
	}
	if cnf.IsSet("plugin") {
		cnf.Map("plugin", func(root string, sub *conf.Configuration) {
			if err := sdk.RegisterPlugin(root, sub); err != nil {
				panic(err)
			}
		})
	}
	return
}

// DefaultSDK creates a new SDK with default configuration.
func DefaultSDK(cfg *conf.Configuration) (sdk *SDK, err error) {
	defaultCfg := map[string]any{
		"client": map[string]any{
			"timeout": "2s",
		},
		"signer": map[string]any{
			"authScheme":  "KO-HMAC-SHA1",
			"authHeaders": []string{"timestamp", "nonce"},
			"signedLookups": map[string]any{
				"accessToken": "header:authorization>bearer",
				"url":         "CanonicalUri",
			},
			"nonceLen": 12,
		},
	}
	ncfg := conf.NewFromStringMap(defaultCfg)
	if err = ncfg.ParserOperator().Merge(cfg.ParserOperator()); err != nil {
		return nil, err
	}
	return NewSDK(ncfg)
}

// GetToken returns the token of the token source.
func (sdk *SDK) GetToken() (*oauth2.Token, error) {
	return sdk.tokenSource.Token()
}

// RegisterPlugin registers a plugin. Plugins are used to extend the SDK.
func (sdk *SDK) RegisterPlugin(name string, cnf *conf.Configuration) error {
	switch name {
	case PluginFS:
		p := NewFs()
		if err := p.Apply(sdk, cnf); err != nil {
			return err
		}
		sdk.plugins[name] = p
	case PluginMsg:
		p := NewMsg()
		if err := p.Apply(sdk, cnf); err != nil {
			return err
		}
		sdk.plugins[name] = p
	case PluginAuth:
		p := NewAuth()
		if err := p.Apply(sdk, cnf); err != nil {
			return err
		}
		sdk.plugins[name] = p
	default:
		return fmt.Errorf("plugin %s is not supported", name)
	}
	return nil
}

// GetPlugin returns a plugin by name.
func (sdk *SDK) GetPlugin(name string) (Plugin, bool) {
	v, ok := sdk.plugins[name]
	return v, ok
}

// Fs returns the file system plugin.
func (sdk *SDK) Fs() *Fs {
	return sdk.plugins[PluginFS].(*Fs)
}

// Msg returns the msg plugin.
func (sdk *SDK) Msg() *Msg {
	return sdk.plugins[PluginMsg].(*Msg)
}

func (sdk *SDK) Auth() *Auth {
	return sdk.plugins[PluginAuth].(*Auth)
}

// TenantIDInterceptor is a client intercept that try to inject tenant id into request header.
// only support string and int type.
func TenantIDInterceptor(ctx context.Context, req *http.Request) error {
	switch tid := ctx.Value(identity.TenantContextKey).(type) {
	case string:
		req.Header.Add(identity.TenantHeaderKey, tid)
	case int:
		req.Header.Add(identity.TenantHeaderKey, strconv.Itoa(tid))
	}
	return nil
}
