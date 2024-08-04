package alioss

import (
	"context"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts20150401 "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/woocoos/knockout-go/api/fs"
	"strings"
	"time"
)

func init() {
	fs.RegisterS3Provider(fs.KindAliOSS, BuildProvider)
}

// Provider for ali yun. it implements fs.S3Provider
// At first, you shoulr register this provider before using it.
//
//	 import (
//	 	"github.com/woocoos/knockout-go/api/fs/alioss"
//	 )
//		fs.RegisterS3Provider(fs.KindAliOSS, alioss.BuildProvider)
type Provider struct {
	ctx        context.Context
	stsClient  *sts20150401.Client
	ossClient  *oss.Client
	s3Client   *s3.Client
	fileSource *fs.ProviderConfig
}

// BuildProvider create aws s3 provider. it matches fs.S3ProviderBuilder
func BuildProvider(ctx context.Context, fileSource *fs.ProviderConfig) (fs.S3Provider, error) {
	svc := &Provider{
		ctx:        ctx,
		fileSource: fileSource,
	}
	err := svc.initAliSTS()
	if err != nil {
		return nil, err
	}
	err = svc.initAliOSS()
	if err != nil {
		return nil, err
	}
	err = svc.initAwsClient()
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// initAliSTS init ali yun oss sts client
func (svc *Provider) initAliSTS() error {
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(svc.fileSource.AccessKeyID),
		AccessKeySecret: tea.String(svc.fileSource.AccessKeySecret),
		RegionId:        tea.String(svc.fileSource.Region),
	}
	cfg.Endpoint = tea.String(svc.fileSource.StsEndpoint)
	stsClient, err := sts20150401.NewClient(cfg)
	if err != nil {
		return err
	}
	svc.stsClient = stsClient
	return nil
}

// initAliOSS init ali yun oss client
func (svc *Provider) initAliOSS() error {
	useCname := false
	// use custom domain
	if svc.fileSource.EndpointImmutable {
		useCname = true
	}
	client, err := oss.New(svc.fileSource.Endpoint, svc.fileSource.AccessKeyID, svc.fileSource.AccessKeySecret, oss.UseCname(useCname))
	if err != nil {
		return err
	}
	svc.ossClient = client
	return nil
}

// initAwsClient init s3 compatible client
func (svc *Provider) initAwsClient() error {
	s3Client, err := fs.InitAwsClient(svc.ctx, svc.fileSource)
	if err != nil {
		return err
	}
	svc.s3Client = s3Client
	return nil
}

// GetSTS get STS
// note: roleSessionName is required, but you can pass an any string(by zmm).
func (svc *Provider) GetSTS(roleSessionName string) (*fs.STSResponse, error) {
	assumeRoleRequest := &sts20150401.AssumeRoleRequest{
		RoleSessionName: tea.String(roleSessionName),
		RoleArn:         tea.String(svc.fileSource.RoleArn),
		DurationSeconds: tea.Int64(int64(svc.fileSource.DurationSeconds)),
	}
	if svc.fileSource.Policy != "" {
		assumeRoleRequest.Policy = tea.String(svc.fileSource.Policy)
	}

	resp, err := svc.stsClient.AssumeRoleWithOptions(assumeRoleRequest, &service.RuntimeOptions{})
	if err != nil {
		return nil, err
	}

	expiration, err := time.Parse(time.RFC3339, *resp.Body.Credentials.Expiration)
	if err != nil {
		return nil, err
	}
	return &fs.STSResponse{
		AccessKeyID:     *resp.Body.Credentials.AccessKeyId,
		SecretAccessKey: *resp.Body.Credentials.AccessKeySecret,
		SessionToken:    *resp.Body.Credentials.SecurityToken,
		Expiration:      expiration,
	}, nil
}

// GetPreSignedURL get aliyun presign url
func (svc *Provider) GetPreSignedURL(bucket, path string, expires time.Duration) (string, error) {
	bk, err := svc.ossClient.Bucket(bucket)
	if err != nil {
		return "", err
	}
	signedURL, err := bk.SignURL(strings.TrimLeft(path, "/"), oss.HTTPGet, int64(expires.Seconds()))
	if err != nil {
		return "", err
	}
	return signedURL, nil
}

// S3Client get s3 client
func (svc *Provider) S3Client() *s3.Client {
	return svc.s3Client
}
