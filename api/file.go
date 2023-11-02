package api

import (
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/api/file"
)

type File struct {
	*file.APIClient
	cfg *file.Config
}

func NewFile() *File {
	return &File{
		cfg: file.NewConfig(),
	}
}

func (f *File) Apply(sdk *SDK, cnf *conf.Configuration) error {
	if err := cnf.Unmarshal(f.cfg); err != nil {
		return err
	}
	f.cfg.HTTPClient = sdk.client
	f.APIClient = file.NewAPIClient(f.cfg)
	f.AddInterceptor(TenantIDInterceptor)
	return nil
}
