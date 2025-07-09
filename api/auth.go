package api

import (
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/api/auth"
)

type Auth struct {
	*auth.APIClient
	cfg *auth.Config
}

func NewAuth() *Auth {
	return &Auth{
		cfg: auth.NewConfig(),
	}
}

func (f *Auth) Apply(sdk *SDK, cnf *conf.Configuration) error {
	if err := cnf.Unmarshal(f.cfg); err != nil {
		return err
	}
	f.cfg.HTTPClient = sdk.client
	f.APIClient = auth.NewAPIClient(f.cfg)
	f.AddInterceptor(TenantIDInterceptor)
	return nil
}
