package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"corridor/internal/config"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
)

// Setup uses separate administrator credentials; HTTP startup never calls it.
func Setup(ctx context.Context, cfg config.Config, rootKey, rootSecret string) error {
	adminCfg := config.Config{S3Endpoint: cfg.S3Endpoint, S3AccessKey: rootKey, S3SecretKey: rootSecret, S3Bucket: cfg.S3Bucket}
	client, err := New(adminCfg)
	if err != nil {
		return err
	}
	exists, err := client.api.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		return fmt.Errorf("check setup bucket: %w", err)
	}
	if !exists {
		if err := client.api.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
			return fmt.Errorf("create private bucket: %w", err)
		}
	}
	if err := client.api.SetBucketPolicy(ctx, cfg.S3Bucket, ""); err != nil {
		return fmt.Errorf("enforce private bucket: %w", err)
	}
	endpoint, _ := url.Parse(cfg.S3Endpoint)
	admin, err := madmin.New(endpoint.Host, rootKey, rootSecret, endpoint.Scheme == "https")
	if err != nil {
		return fmt.Errorf("setup admin client: %w", err)
	}
	policy, err := json.Marshal(map[string]any{"Version": "2012-10-17", "Statement": []any{map[string]any{"Effect": "Allow", "Action": []string{"s3:ListBucket"}, "Resource": []string{"arn:aws:s3:::" + cfg.S3Bucket}}}})
	if err != nil {
		return fmt.Errorf("encode bucket policy: %w", err)
	}
	if err := admin.AddCannedPolicy(ctx, "corridor-health", policy); err != nil {
		return fmt.Errorf("install health policy: %w", err)
	}
	if err := admin.AddUser(ctx, cfg.S3AccessKey, cfg.S3SecretKey); err != nil {
		return fmt.Errorf("provision storage user: %w", err)
	}
	if _, err := admin.AttachPolicy(ctx, madmin.PolicyAssociationReq{Policies: []string{"corridor-health"}, User: cfg.S3AccessKey}); err != nil {
		if madmin.ToErrorResponse(err).Code != "XMinioAdminPolicyChangeAlreadyApplied" {
			return fmt.Errorf("attach health policy: %w", err)
		}
	}
	return nil
}
