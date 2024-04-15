package fs

import (
	"github.com/minio/minio-go/v7"
	"github.com/tsingsun/woocoo/pkg/conf"
)

// S3ProxyOptions minio client options.
// notice Minio-go is under apache-2.0 license, not like minio server which is AGPL.
type S3ProxyOptions struct {
	Endpoint        string `json:"endpoint" yaml:"endpoint"`
	AccessKeyID     string `json:"accessKeyID" yaml:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey" yaml:"secretAccessKey"`
	UseSSL          bool   `json:"useSSL,omitempty" yaml:"useSSL"`
}

// S3ProxyOption minio client option
type S3ProxyOption func(*S3ProxyOptions)

// WithConfiguration init minio client with configuration.
func WithConfiguration(cnf *conf.Configuration) S3ProxyOption {
	return func(opts *S3ProxyOptions) {
		if err := cnf.Unmarshal(opts); err != nil {
			panic(err)
		}
	}
}

// S3Proxy minio client proxy.
type S3Proxy struct {
	*minio.Client
}

// NewS3Proxy creates a new minio proxy.
func NewS3Proxy(cnf *conf.Configuration) (*S3Proxy, error) {
	cli, err := NewMinioClient(WithConfiguration(cnf))
	if err != nil {
		return nil, err
	}
	return &S3Proxy{cli}, nil
}
