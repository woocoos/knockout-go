package api

import (
	"context"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/httpx"
	"golang.org/x/oauth2"
	"net/http"
	"time"
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
	return
}

// RegisterPlugin registers a plugin. Plugins are used to extend the SDK.
func (sdk *SDK) RegisterPlugin(name string, cnf *conf.Configuration) error {
	switch name {
	case "file":
		p := NewFile()
		if err := p.Apply(sdk, cnf); err != nil {
			return err
		}
		sdk.plugins[name] = p
	case "msg":
		p := NewMsg()
		if err := p.Apply(sdk, cnf); err != nil {
			return err
		}
		sdk.plugins[name] = p
	}
	return nil
}

// GetPlugin returns a plugin by name.
func (sdk *SDK) GetPlugin(name string) (Plugin, bool) {
	v, ok := sdk.plugins[name]
	return v, ok
}

// File returns the file plugin.
func (sdk *SDK) File() *File {
	return sdk.plugins["file"].(*File)
}

// Msg returns the msg plugin.
func (sdk *SDK) Msg() *Msg {
	return sdk.plugins["msg"].(*Msg)
}
