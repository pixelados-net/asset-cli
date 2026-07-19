// Package minio contains the reusable MinIO object storage adapter.
package minio

import "fmt"

// Config contains MinIO connection and bucket settings.
type Config struct {
	// Endpoint is the MinIO server host and port.
	Endpoint string `env:"ENDPOINT" envDefault:"127.0.0.1:9000"`
	// AccessKey authenticates against the MinIO server.
	AccessKey string `env:"ACCESS_KEY"`
	// SecretKey authenticates against the MinIO server.
	SecretKey string `env:"SECRET_KEY"`
	// Bucket is the bucket this process manages.
	Bucket string `env:"BUCKET"`
	// Region is the optional bucket region.
	Region string `env:"REGION" envDefault:""`
	// UseSSL enables TLS when dialing the MinIO server.
	UseSSL bool `env:"USE_SSL" envDefault:"true"`
}

// Validate reports whether the mandatory MinIO settings were supplied.
func (config Config) Validate() error {
	if config.Endpoint == "" {
		return fmt.Errorf("minio endpoint is required")
	}
	if config.AccessKey == "" {
		return fmt.Errorf("minio access key is required")
	}
	if config.SecretKey == "" {
		return fmt.Errorf("minio secret key is required")
	}
	if config.Bucket == "" {
		return fmt.Errorf("minio bucket is required")
	}
	return nil
}
