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
	ctx        context.Context
	stsClient  *sts.Client
	s3Client   *s3.Client
	fileSource *SourceConfig
}

// BuildAwsS3 create aws s3 provider. it matches S3ProviderBuilder
func BuildAwsS3(ctx context.Context, fileSource *SourceConfig) (S3Provider, error) {
	svc := &AwsS3{
		ctx:        ctx,
		fileSource: fileSource,
	}
	stsClient, err := initAwsSTS(svc.ctx, svc.fileSource)
	if err != nil {
		return nil, err
	}
	svc.stsClient = stsClient
	s3Client, err := InitAwsClient(svc.ctx, svc.fileSource)
	if err != nil {
		return nil, err
	}
	svc.s3Client = s3Client
	return svc, nil
}

// initAwsSTS init aws sts client
func initAwsSTS(ctx context.Context, fs *SourceConfig) (*sts.Client, error) {
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
func InitAwsClient(ctx context.Context, fs *SourceConfig) (*s3.Client, error) {
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

// GetSTS get sts
func (svc *AwsS3) GetSTS(roleSessionName string) (*STSResponse, error) {
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(svc.fileSource.RoleArn),
		Policy:          aws.String(svc.fileSource.Policy),
		RoleSessionName: aws.String(roleSessionName),
		DurationSeconds: aws.Int32(int32(svc.fileSource.DurationSeconds)),
	}
	out, err := svc.stsClient.AssumeRole(svc.ctx, input)
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
func (svc *AwsS3) GetPreSignedURL(bucket, path string, expires time.Duration) (string, error) {
	pClient := s3.NewPresignClient(svc.s3Client)
	resp, err := pClient.PresignGetObject(svc.ctx, &s3.GetObjectInput{
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
func (svc *AwsS3) S3Client() *s3.Client {
	return svc.s3Client
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
