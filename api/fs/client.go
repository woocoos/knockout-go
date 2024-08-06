package fs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"sync"
	"time"
)

var (
	// S3Builders s3 provider builders.
	S3Builders = map[Kind]S3ProviderBuilder{}
)

func init() {
	RegisterS3Provider(KindMinio, BuildAwsS3)
	RegisterS3Provider(KindAwsS3, BuildAwsS3)
}

// RegisterS3Provider register s3 provider builder. A builder is a function that creates a new s3 provider.
func RegisterS3Provider(kind Kind, builder S3ProviderBuilder) {
	S3Builders[kind] = builder
}

// S3ProviderBuilder s3 provider builder.
type S3ProviderBuilder func(context.Context, *ProviderConfig) (S3Provider, error)

// S3Provider s3 provider interface.
type S3Provider interface {
	// GetSTS get sts response.
	GetSTS(ctx context.Context, roleSessionName string) (*STSResponse, error)
	// GetPreSignedURL get pre-signed url to make request format match each S3Provider.
	GetPreSignedURL(ctx context.Context, bucket, path string, expires time.Duration) (string, error)
	// S3Client return s3 client.
	S3Client() *s3.Client
}

// STSResponse sts response.
type STSResponse struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// Kind file source kind
type Kind string

// Kind values.
const (
	KindMinio  Kind = "minio"
	KindAliOSS Kind = "aliOSS"
	KindAwsS3  Kind = "awsS3"
)

// ProviderConfig define file system source config.
type ProviderConfig struct {
	// file source kind
	Kind Kind `json:"kind" yaml:"kind"`
	// access key id
	AccessKeyID string `json:"accessKeyID,omitempty" yaml:"accessKeyID"`
	// access key secret
	AccessKeySecret string `json:"accessKeySecret,omitempty" yaml:"accessKeySecret,omitempty"`
	// access endpoint, used for public access
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	// endpoint is immutable, if it's custom domain, set to true
	EndpointImmutable bool `json:"endpointImmutable,omitempty" yaml:"endpointImmutable,omitempty"`
	// sts endpoint, used for sts
	StsEndpoint string `json:"stsEndpoint,omitempty" yaml:"stsEndpoint,omitempty"`
	// region, physical location of data storage
	Region string `json:"region,omitempty" yaml:"region,omitempty"`
	// file storage bucket
	Bucket string `json:"bucket" yaml:"bucket"`
	// file storage bucket url, used for matching url
	BucketUrl string `json:"bucketUrl" yaml:"bucketUrl"`
	// role arn, used for STS
	RoleArn string `json:"roleArn,omitempty" yaml:"roleArn,omitempty"`
	// specify the policy for the returned STS token
	Policy string `json:"policy,omitempty" yaml:"policy,omitempty"`
	// duration of the STS token, default 3600s
	DurationSeconds int `json:"durationSeconds,omitempty" yaml:"durationSeconds,omitempty"`
}

// Config file system config. if you want to add source config in config file.
type Config struct {
	Providers []ProviderConfig `json:"providers" yaml:"providers"`
}

// NewConfig create a new config.
func NewConfig() *Config {
	return &Config{
		Providers: make([]ProviderConfig, 0),
	}
}

// Client file system client.
type Client struct {
	cfg       *Config
	providers map[string]S3Provider

	mu sync.RWMutex
}

// NewClient create a new file system client.
func NewClient(cfg *Config) (*Client, error) {
	c := &Client{
		cfg:       cfg,
		providers: make(map[string]S3Provider),
	}
	for _, source := range c.cfg.Providers {
		_, err := c.GetProvider(&source)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// GetProvider get file system provider.If provider is not exist, create a new one and cache it.
// If you want to create a new provider in this method, you should pass all config value.
// Note that the cache key is the combination of access key id, endpoint, bucket and kind, maybe change it later.
func (c *Client) GetProvider(fs *ProviderConfig) (S3Provider, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pk := getProviderKey(fs)
	v, ok := c.providers[pk]
	if ok {
		return v, nil
	}
	builder, ok := S3Builders[fs.Kind]
	if !ok {
		return nil, fmt.Errorf("file system kind: %s is not supported", fs.Kind)
	}
	provider, err := builder(context.Background(), fs)
	if err != nil {
		return nil, err
	}
	c.providers[pk] = provider
	return provider, nil
}

func getProviderKey(fs *ProviderConfig) string {
	return fmt.Sprintf("%s:%s:%s:%s", fs.AccessKeyID, fs.Endpoint, fs.Bucket, fs.Kind)
}
