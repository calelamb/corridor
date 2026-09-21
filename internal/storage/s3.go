package storage

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"corridor/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	api    *minio.Client
	bucket string
}

func New(cfg config.Config) (*Client, error) {
	endpoint, err := url.Parse(cfg.S3Endpoint)
	if err != nil || endpoint.Host == "" || endpoint.User != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return nil, errors.New("invalid storage endpoint")
	}
	if cfg.S3AccessKey == "" || cfg.S3SecretKey == "" || cfg.S3Bucket == "" {
		return nil, errors.New("storage credentials and bucket required")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 3 * time.Second
	api, err := minio.New(endpoint.Host, &minio.Options{Creds: credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""), Secure: endpoint.Scheme == "https", Region: "us-east-1", Transport: transport})
	if err != nil {
		return nil, errors.New("storage client configuration invalid")
	}
	return &Client{api: api, bucket: cfg.S3Bucket}, nil
}
func (c *Client) Check(ctx context.Context) error {
	exists, err := c.api.BucketExists(ctx, c.bucket)
	if err != nil || !exists {
		return errors.New("private storage unavailable")
	}
	return nil
}
