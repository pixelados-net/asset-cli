package minio

import (
	"bytes"
	"context"

	sdk "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps the reusable MinIO driver for a single managed bucket.
type Client struct {
	client *sdk.Client
	bucket string
}

// New creates a MinIO client without performing network I/O.
func New(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	client, err := sdk.New(config.Endpoint, &sdk.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, err
	}
	return &Client{client: client, bucket: config.Bucket}, nil
}

// SDK returns the underlying MinIO client.
func (client *Client) SDK() *sdk.Client {
	return client.client
}

// Bucket returns the managed bucket name.
func (client *Client) Bucket() string {
	return client.bucket
}

// Ping verifies that the managed bucket is reachable.
func (client *Client) Ping(ctx context.Context) error {
	_, err := client.client.BucketExists(ctx, client.bucket)
	return err
}

// Exists reports whether any object exists under the given key prefix.
func (client *Client) Exists(ctx context.Context, prefix string) (bool, error) {
	objects := client.client.ListObjects(ctx, client.bucket, sdk.ListObjectsOptions{Prefix: prefix, MaxKeys: 1})
	object, ok := <-objects
	if !ok {
		return false, nil
	}
	if object.Err != nil {
		return false, object.Err
	}
	return true, nil
}

// Touch creates an empty placeholder object at key so the prefix becomes visible.
func (client *Client) Touch(ctx context.Context, key string) error {
	_, err := client.client.PutObject(ctx, client.bucket, key, bytes.NewReader(nil), 0, sdk.PutObjectOptions{})
	return err
}
