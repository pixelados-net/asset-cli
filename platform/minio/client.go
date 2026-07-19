package minio

import (
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
