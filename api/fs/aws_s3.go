package fs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	awsCredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"net/url"
	"strings"
	"time"
)

// AwsS3 aws s3 service for file storage.
type AwsS3 struct {
	stsClient *sts.Client
	s3Client  *s3.Client
	config    *ProviderConfig
}

// BuildAwsS3 create aws s3 provider. it matches S3ProviderBuilder
func BuildAwsS3(ctx context.Context, fileSource *ProviderConfig) (S3Provider, error) {
	svc := &AwsS3{
		config: fileSource,
	}
	stsClient, err := initAwsSTS(ctx, svc.config)
	if err != nil {
		return nil, err
	}
	svc.stsClient = stsClient
	s3Client, err := InitAwsClient(ctx, svc.config)
	if err != nil {
		return nil, err
	}
	svc.s3Client = s3Client
	return svc, nil
}

// initAwsSTS init aws sts client
func initAwsSTS(ctx context.Context, fs *ProviderConfig) (*sts.Client, error) {
	creds := awsCredentials.NewStaticCredentialsProvider(fs.AccessKeyID, fs.AccessKeySecret, "")
	cfg, err := awsCfg.LoadDefaultConfig(ctx, awsCfg.WithCredentialsProvider(creds))
	if err != nil {
		return nil, err
	}

	stsClient := sts.NewFromConfig(cfg, func(options *sts.Options) {
		options.Region = fs.Region
		options.BaseEndpoint = aws.String(fs.StsEndpoint)
	})
	return stsClient, nil
}

// InitAwsClient init aws s3 client.
func InitAwsClient(ctx context.Context, fs *ProviderConfig) (*s3.Client, error) {
	creds := awsCredentials.NewStaticCredentialsProvider(fs.AccessKeyID, fs.AccessKeySecret, "")
	cfg, err := awsCfg.LoadDefaultConfig(ctx, awsCfg.WithCredentialsProvider(creds))
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.Region = fs.Region
		// customer endpoint resolver
		options.EndpointResolverV2 = &EndpointResolverV2{
			EndpointImmutable: fs.EndpointImmutable, // if true，use custom endpoint
			endpoint:          fs.Endpoint,
		}
		options.BaseEndpoint = aws.String(fs.Endpoint)
		// if minio，need to use path style
		if fs.Kind == KindMinio {
			options.UsePathStyle = true
		}
	})
	return s3Client, nil
}

// ProviderConfig return ProviderConfig. Provider need hold a ProviderConfig.
func (p *AwsS3) ProviderConfig() *ProviderConfig {
	return p.config
}

// ParseUrlKey parse url key.
func (p *AwsS3) ParseUrlKey(urlStr string) (key string, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	if p.config.Kind == KindMinio {
		key = strings.TrimPrefix(u.Path, "/"+p.config.Bucket)
	} else {
		key = strings.TrimPrefix(u.Path, "/")
	}
	return key, nil
}

// GetSTS get sts
func (p *AwsS3) GetSTS(ctx context.Context, roleSessionName string) (*STSResponse, error) {
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(p.config.RoleArn),
		Policy:          aws.String(p.config.Policy),
		RoleSessionName: aws.String(roleSessionName),
		DurationSeconds: aws.Int32(int32(p.config.DurationSeconds)),
	}
	out, err := p.stsClient.AssumeRole(ctx, input)
	if err != nil {
		return nil, err
	}
	return &STSResponse{
		AccessKeyID:     *out.Credentials.AccessKeyId,
		SecretAccessKey: *out.Credentials.SecretAccessKey,
		SessionToken:    *out.Credentials.SessionToken,
		Expiration:      *out.Credentials.Expiration,
	}, nil
}

// GetPreSignedURL get pre-signed url
func (p *AwsS3) GetPreSignedURL(ctx context.Context, bucket, path string, expires time.Duration) (string, error) {
	pClient := s3.NewPresignClient(p.s3Client)
	resp, err := pClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(strings.TrimLeft(path, "/")),
	}, func(options *s3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("signUrl response is nil")
	}
	return resp.URL, nil
}

// S3Client get s3Client
func (p *AwsS3) S3Client() *s3.Client {
	return p.s3Client
}

// EndpointResolverV2 customer endpoint resolver
type EndpointResolverV2 struct {
	endpoint          string
	EndpointImmutable bool
}

// ResolveEndpoint resolve endpoint
func (r *EndpointResolverV2) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	// if the endpoint is immutable, return the configured endpoint
	if r.EndpointImmutable {
		u, err := url.Parse(r.endpoint)
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}
	// delegate back to the default v2 resolver otherwise
	return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
