package fs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"
	"os"
	"path/filepath"
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
	// ProviderConfig return ProviderConfig. Provider need hold a ProviderConfig.
	ProviderConfig() *ProviderConfig
	// ParseUrlKey parse url key.
	ParseUrlKey(urlStr string) (key string, err error)
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
	BasePath   string            `json:"basePath,omitempty" yaml:"basePath,omitempty"`
	Host       string            `json:"host,omitempty" yaml:"host,omitempty"`
	Headers    map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	UserAgent  string            `json:"userAgent,omitempty" yaml:"userAgent,omitempty"`
	HTTPClient *http.Client
	Providers  []ProviderConfig `json:"providers" yaml:"providers"`
}

// NewConfig create a new config.
func NewConfig() *Config {
	return &Config{
		BasePath:  "http://localhost:8080/graphql/query",
		Providers: make([]ProviderConfig, 0),
	}
}

type BizKey interface {
	int | string
}

// Client file system client.
type Client struct {
	cfg       *Config
	providers map[string]S3Provider
	// biz key -> provider key
	keys map[string]string

	interceptors    []InterceptFunc
	FileIdentityAPI *FileIdentityAPI

	mu sync.RWMutex
}

// NewClient create a new file system client.
func NewClient(cfg *Config) (*Client, error) {
	c := &Client{
		cfg:       cfg,
		providers: make(map[string]S3Provider),
		keys:      make(map[string]string),
	}

	api := api{
		client: c,
	}
	c.FileIdentityAPI = (*FileIdentityAPI)(&api)

	for _, source := range c.cfg.Providers {
		key := GetProviderKey(&source)
		err := c.RegistryProvider(&source, key)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// GetProviderByBizKey get provider by biz key
func (c *Client) GetProviderByBizKey(bizKey string) (S3Provider, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pk, ok := c.keys[bizKey]
	if !ok {
		return nil, fmt.Errorf("provider is not exist")
	}
	v, ok := c.providers[pk]
	if !ok {
		return nil, fmt.Errorf("provider is not exist")
	}
	return v, nil
}

// RegistryProvider registry a file system provider.If provider is not exist, create a new one and cache it.
// Note that the cache key is the combination of access key id, endpoint, bucket and kind, maybe change it later.
// parameter bizKey is used to set/replace a customer key for the provider, so that you can get the provider by the key.
func (c *Client) RegistryProvider(cfg *ProviderConfig, bizKey string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	pk := GetProviderKey(cfg)
	_, ok := c.providers[pk]
	if ok {
		if bizKey != "" {
			c.keys[bizKey] = pk
		}
		return nil
	}
	builder, ok := S3Builders[cfg.Kind]
	if !ok {
		return fmt.Errorf("file system kind: %s is not supported", cfg.Kind)
	}
	provider, err := builder(context.Background(), cfg)
	if err != nil {
		return err
	}
	c.providers[pk] = provider
	if bizKey != "" {
		c.keys[bizKey] = pk
	}
	return nil
}

type DownloadOption struct {
	// if true, overwrite the local file if it exists
	OverwrittenFile bool
}

type DownloadOptionFn func(options *DownloadOption)

func WithOverwrittenFile(overwrittenFile bool) DownloadOptionFn {
	return func(options *DownloadOption) {
		options.OverwrittenFile = overwrittenFile
	}
}

// DownloadObjectByKey download object by biz key(tenantID or ...).If local file exists, it will be overwritten.
func (c *Client) DownloadObjectByKey(bizKey string, url string, localFile string, optFns ...DownloadOptionFn) error {
	options := DownloadOption{
		OverwrittenFile: true,
	}
	for _, fn := range optFns {
		fn(&options)
	}
	if !options.OverwrittenFile {
		if _, err := os.Stat(localFile); !os.IsNotExist(err) {
			return nil
		}
	}
	provider, err := c.GetProviderByBizKey(bizKey)
	if err != nil {
		return err
	}
	fileKey, err := provider.ParseUrlKey(url)
	if err != nil {
		return err
	}
	getObjOutput, err := provider.S3Client().GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &provider.ProviderConfig().Bucket,
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return err
	}
	defer getObjOutput.Body.Close()
	err = os.MkdirAll(filepath.Dir(localFile), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.ReadFrom(getObjOutput.Body)
	if err != nil {
		return err
	}
	return nil
}

// GetProviderKey get provider key.
func GetProviderKey(fs *ProviderConfig) string {
	return fmt.Sprintf("%s:%s:%s:%s", fs.AccessKeyID, fs.Endpoint, fs.Bucket, fs.Kind)
}
