package api

import (
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/api/msg"
)

type Msg struct {
	*msg.APIClient
	cfg *msg.Config
}

func NewMsg() *Msg {
	return &Msg{
		cfg: msg.NewConfig(),
	}
}

func (f *Msg) Apply(sdk *SDK, cnf *conf.Configuration) error {
	if err := cnf.Unmarshal(f.cfg); err != nil {
		return err
	}
	f.cfg.HTTPClient = sdk.client
	f.APIClient = msg.NewAPIClient(f.cfg)
	return nil
}
