package storage

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/minio/minio-go/v7"
)

var archiveKey = regexp.MustCompile(`^sha256/[a-f0-9]{64}/[a-z0-9._-]+$`)

// Archive is used only by the administrative ingestion CLI, never the HTTP role.
func (c *Client) Archive(ctx context.Context, key string, reader io.Reader, size int64) error {
	if !archiveKey.MatchString(key) || size < 1 || size > 256<<20 {
		return fmt.Errorf("invalid artifact key or size")
	}
	_, err := c.api.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	if err == nil {
		return nil
	}
	if minio.ToErrorResponse(err).Code != "NoSuchKey" {
		return fmt.Errorf("check private artifact: %w", err)
	}
	_, err = c.api.PutObject(ctx, c.bucket, key, reader, size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return fmt.Errorf("archive private artifact: %w", err)
	}
	return nil
}
