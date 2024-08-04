package api

import (
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/api/fs"
)

// Fs file system client. the file system is a file storage service, such as s3, oss, etc.
type Fs struct {
	*fs.Client
	cfg *fs.Config
}

// NewFs creates a new file system client.
func NewFs() *Fs {
	return &Fs{
		cfg: fs.NewConfig(),
	}
}

func (f *Fs) Apply(_ *SDK, cnf *conf.Configuration) error {
	err := cnf.Unmarshal(f.cfg)
	if err != nil {
		return err
	}
	f.Client, err = fs.NewClient(f.cfg)
	return err
}
