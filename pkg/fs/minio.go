package fs

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// NewMinioClient creates a new minio client.
func NewMinioClient(opts ...S3ProxyOption) (*minio.Client, error) {
	options := &S3ProxyOptions{}
	for _, opt := range opts {
		opt(options)
	}
	cli, err := minio.New(options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(options.AccessKeyID, options.SecretAccessKey, ""),
		Secure: options.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}
