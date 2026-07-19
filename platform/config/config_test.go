package config

import (
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/pixelados-net/asset-cli/platform/logger"
)

func TestLoadParsesEveryModule(t *testing.T) {
	t.Setenv("ASSET_CLI_LOG_LEVEL", "debug")
	t.Setenv("ASSET_CLI_LOG_FORMAT", "json")
	t.Setenv("ASSET_CLI_MINIO_ENDPOINT", "127.0.0.1:9000")
	t.Setenv("ASSET_CLI_MINIO_ACCESS_KEY", "key")
	t.Setenv("ASSET_CLI_MINIO_SECRET_KEY", "secret")
	t.Setenv("ASSET_CLI_MINIO_BUCKET", "assets")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Logger.Level != zapcore.DebugLevel || config.Logger.Format != logger.FormatJSON {
		t.Fatalf("Logger = %#v", config.Logger)
	}
	if config.MinIO.Endpoint != "127.0.0.1:9000" || config.MinIO.Bucket != "assets" {
		t.Fatalf("MinIO = %#v", config.MinIO)
	}
}

func TestLoadRejectsMissingMinIOBucket(t *testing.T) {
	t.Setenv("ASSET_CLI_MINIO_ENDPOINT", "127.0.0.1:9000")
	t.Setenv("ASSET_CLI_MINIO_ACCESS_KEY", "key")
	t.Setenv("ASSET_CLI_MINIO_SECRET_KEY", "secret")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil")
	}
}
