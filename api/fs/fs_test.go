package fs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	minioProviderConfig = ProviderConfig{
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
	cli, err := InitAwsClient(context.Background(), &minioProviderConfig)
	t.Require().NoError(err)
	_, err = cli.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(minioProviderConfig.Bucket)},
	)
	if err != nil {
		t.Require().ErrorContains(err, "BucketAlreadyOwnedByYou")
	}
	t.client, err = NewClient(&Config{Providers: []ProviderConfig{minioProviderConfig}})
	t.Require().NoError(err)
}

func (t *fsSuite) TestMinioSTS() {
	provider, err := t.client.GetProvider(context.TODO(), &minioProviderConfig)
	t.NoError(err)
	resp, err := provider.GetSTS("")
	t.NoError(err)
	fmt.Println(resp)
}

func (t *fsSuite) TestMinioPreSignedUrl() {
	provider, err := t.client.GetProvider(context.TODO(), &minioProviderConfig)
	t.NoError(err)
	u, err := provider.GetPreSignedURL("knockout-go", "3a9809ba339ec87f1636c7878685f616.jpeg", time.Hour)
	t.NoError(err)
	fmt.Println(u)
}

func (t *fsSuite) TestMinioS3GetObject_NoSuchKey() {
	provider, err := t.client.GetProvider(context.TODO(), &minioProviderConfig)
	t.NoError(err)
	s3Client := provider.S3Client()
	_, err = s3Client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String("knockout-go"), Key: aws.String("/NoSuchKey.jpeg")})
	t.ErrorContains(err, "NoSuchKey")
}
