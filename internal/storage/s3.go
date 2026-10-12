package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

// PutPhoto uploads photo bytes to private storage.
func (c *Client) PutPhoto(ctx context.Context, key string, data []byte) error {
	_, err := c.api.PutObject(ctx, c.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return fmt.Errorf("store private photo: %w", err)
	}
	return nil
}

// GetPhoto retrieves raw photo bytes from private storage.
func (c *Client) GetPhoto(ctx context.Context, key string) ([]byte, error) {
	obj, err := c.api.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get private photo: %w", err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

