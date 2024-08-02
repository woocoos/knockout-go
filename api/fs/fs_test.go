package fs

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	minioFileSource = SourceConfig{
		Kind:              KindMinio,
		AccessKeyID:       "minioadmin",
		AccessKeySecret:   "minioadmin",
		Endpoint:          "http://127.0.0.1:9000",
		EndpointImmutable: false,
		StsEndpoint:       "http://127.0.0.1:9000",
		Region:            "us-east-1",
		RoleArn:           "arn:aws:s3:::*",
		Policy:            "",
		DurationSeconds:   3600,
		Bucket:            "knockout-go",
		BucketUrl:         "http://127.0.0.1:32650/knockout-go",
	}
)

type fsSuite struct {
	suite.Suite
	client *Client
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, &fsSuite{})
}

func (t *fsSuite) SetupSuite() {
	cli, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds:  credentials.NewStaticV4(minioFileSource.AccessKeyID, minioFileSource.AccessKeySecret, ""),
		Secure: false,
	})
	t.Require().NoError(err)
	err = cli.MakeBucket(context.Background(), minioFileSource.Bucket, minio.MakeBucketOptions{
		Region: minioFileSource.Region,
	})
	if err != nil {
		var merr minio.ErrorResponse
		ok := errors.As(err, &merr)
		t.Require().True(ok)
		t.Assert().Equal(merr.Code, "BucketAlreadyOwnedByYou")
	}
	t.client, err = NewClient(&Config{Sources: []SourceConfig{minioFileSource}})
	t.Require().NoError(err)
}

func (t *fsSuite) TestMinioSTS() {
	provider, err := t.client.GetProvider(context.TODO(), &minioFileSource)
	t.NoError(err)
	resp, err := provider.GetSTS("")
	t.NoError(err)
	fmt.Println(resp)
}

func (t *fsSuite) TestMinioPreSignedUrl() {
	provider, err := t.client.GetProvider(context.TODO(), &minioFileSource)
	t.NoError(err)
	u, err := provider.GetPreSignedURL("knockout-go", "3a9809ba339ec87f1636c7878685f616.jpeg", time.Hour)
	t.NoError(err)
	fmt.Println(u)
}

func (t *fsSuite) TestMinioS3GetObject_NoSuchKey() {
	provider, err := t.client.GetProvider(context.TODO(), &minioFileSource)
	t.NoError(err)
	s3Client := provider.S3Client()
	_, err = s3Client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String("knockout-go"), Key: aws.String("/NoSuchKey.jpeg")})
	t.ErrorContains(err, "NoSuchKey")
}
