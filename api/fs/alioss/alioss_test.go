// person test: put a .env file in the same directory.
package alioss_test

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"
	"github.com/woocoos/knockout-go/api/fs"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	aliProviderConfig = fs.ProviderConfig{
		Kind:              fs.KindAliOSS,
		AccessKeyID:       "todo",
		AccessKeySecret:   "todo",
		Endpoint:          "https://%s.aliyuncs.com",
		EndpointImmutable: false,
		StsEndpoint:       "sts.cn-shenzhen.aliyuncs.com",
		RoleArn:           "acs:ram::5755321561100682:role/devossrwrole",
		Bucket:            "todo",
		BucketUrl:         "https://%s.%s.aliyuncs.com",
		Region:            "oss-cn-shenzhen",
		Policy:            "",
		DurationSeconds:   120,
	}
)

type fsSuite struct {
	suite.Suite
}

func TestApiSuite(t *testing.T) {
	if os.Getenv("TEST_WIP") != "" {
		t.Skip("skipping test in short mode.")
	}
	suite.Run(t, &fsSuite{})
}

func (t *fsSuite) SetupSuite() {
	// check .env file in dir
	if _, err := os.Stat(".env"); err == nil {
		// Open .env file
		file, err := os.Open(".env")
		t.Require().NoError(err)
		defer file.Close()

		// Read and parse .env file
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Skip comments and empty lines
			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
				continue
			}
			// Split key-value pairs
			pair := strings.SplitN(line, "=", 2)
			if len(pair) == 2 {
				key := strings.TrimSpace(pair[0])
				value := strings.TrimSpace(pair[1])
				// Remove surrounding quotes if present
				value = strings.Trim(value, `"`)
				os.Setenv(key, value)
			}
		}
		t.Require().NoError(scanner.Err())
		aliProviderConfig.Region = os.Getenv("ALI_OSS_REGION")
		aliProviderConfig.AccessKeyID = os.Getenv("ALI_OSS_ACCESS_KEY_ID")
		aliProviderConfig.AccessKeySecret = os.Getenv("ALI_OSS_ACCESS_KEY_SECRET")
		aliProviderConfig.Bucket = os.Getenv("ALI_OSS_BUCKET")
		aliProviderConfig.RoleArn = os.Getenv("ALI_OSS_ROLE_ARN")
		aliProviderConfig.StsEndpoint = os.Getenv("ALI_OSS_STS_ENDPOINT")

		aliProviderConfig.Endpoint = fmt.Sprintf(aliProviderConfig.Endpoint, aliProviderConfig.Region)
		aliProviderConfig.BucketUrl = fmt.Sprintf(aliProviderConfig.BucketUrl, aliProviderConfig.Bucket, aliProviderConfig.Region)
	}
}

func (t *fsSuite) TestAliSTS() {
	oss, err := fs.NewClient(&fs.Config{Providers: []fs.ProviderConfig{aliProviderConfig}})
	t.Require().NoError(err)
	provider, err := oss.GetProvider(context.TODO(), &aliProviderConfig)
	t.NoError(err)
	resp, err := provider.GetSTS("test")
	t.NoError(err)
	fmt.Println(resp)
}

func (t *fsSuite) TestAliOSSPreSignedUrl() {
	oss, err := fs.NewClient(&fs.Config{Providers: []fs.ProviderConfig{aliProviderConfig}})
	t.NoError(err)
	provider, err := oss.GetProvider(context.TODO(), &aliProviderConfig)
	t.NoError(err)
	u, err := provider.GetPreSignedURL("qldevtest", "cust/159ecc5f964dfe00", time.Hour)
	t.NoError(err)
	fmt.Println(u)
}

func (t *fsSuite) TestAliS3GetObject() {
	oss, err := fs.NewClient(&fs.Config{Providers: []fs.ProviderConfig{aliProviderConfig}})
	t.NoError(err)
	provider, err := oss.GetProvider(context.TODO(), &aliProviderConfig)
	t.NoError(err)
	s3Client := provider.S3Client()
	out, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String("qldevtest"), Key: aws.String("cust/159ecc5f964dfe00")})
	t.NoError(err)
	defer out.Body.Close()
}
