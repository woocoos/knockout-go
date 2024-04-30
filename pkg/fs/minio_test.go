package fs

import (
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/conf"
	"testing"
)

var (
	testMinioOptions = S3ProxyOptions{
		Endpoint:        "127.0.0.1:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
	}
	minioBucketName = "knockout-go"
	minioLocaltion  = "us-east-1"
)

func TestNewMinio(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		cli, err := NewMinioClient(WithConfiguration(conf.NewFromStringMap(map[string]any{
			"endpoint":        testMinioOptions.Endpoint,
			"accessKeyID":     testMinioOptions.AccessKeyID,
			"secretAccessKey": testMinioOptions.SecretAccessKey,
		})))
		require.NoError(t, err)

		err = cli.MakeBucket(context.Background(), minioBucketName, minio.MakeBucketOptions{
			Region: minioLocaltion,
		})
		if err != nil {
			var merr minio.ErrorResponse
			ok := errors.As(err, &merr)
			require.True(t, ok)
			assert.Equal(t, merr.Code, "BucketAlreadyOwnedByYou")
		}
	})
}
