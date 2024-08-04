package alioss_test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/woocoos/knockout-go/api/fs"
	"testing"
	"time"
)

var (
	aliProviderConfig = fs.ProviderConfig{
		Kind:              fs.KindAliOSS,
		AccessKeyID:       "todo",
		AccessKeySecret:   "todo",
		Endpoint:          "https://oss-cn-shenzhen.aliyuncs.com",
		EndpointImmutable: false,
		StsEndpoint:       "sts.cn-shenzhen.aliyuncs.com",
		RoleArn:           "acs:ram::5755321561100682:role/devossrwrole",
		Bucket:            "todo",
		BucketUrl:         "https://todo.oss-cn-shenzhen.aliyuncs.com",
		Region:            "oss-cn-shenzhen",
		Policy:            "",
		DurationSeconds:   3600,
	}
)

func TestAliSTS(t *testing.T) {
	oss, err := fs.NewClient(&fs.Config{Providers: []fs.ProviderConfig{aliProviderConfig}})
	assert.NoError(t, err)
	provider, err := oss.GetProvider(context.TODO(), &aliProviderConfig)
	assert.NoError(t, err)
	resp, err := provider.GetSTS("test")
	assert.NoError(t, err)
	fmt.Println(resp)
}

func TestAliOSSPreSignedUrl(t *testing.T) {
	oss, err := fs.NewClient(&fs.Config{Providers: []fs.ProviderConfig{aliProviderConfig}})
	assert.NoError(t, err)
	provider, err := oss.GetProvider(context.TODO(), &aliProviderConfig)
	assert.NoError(t, err)
	u, err := provider.GetPreSignedURL("qldevtest", "cust/159ecc5f964dfe00", time.Hour)
	assert.NoError(t, err)
	fmt.Println(u)
}

func TestAliS3GetObject(t *testing.T) {
	oss, err := fs.NewClient(&fs.Config{Providers: []fs.ProviderConfig{aliProviderConfig}})
	assert.NoError(t, err)
	provider, err := oss.GetProvider(context.TODO(), &aliProviderConfig)
	assert.NoError(t, err)
	s3Client := provider.S3Client()
	out, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String("qldevtest"), Key: aws.String("cust/159ecc5f964dfe00")})
	assert.NoError(t, err)
	defer out.Body.Close()
}
