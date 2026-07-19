package minio

import "testing"

func TestConfigValidate(t *testing.T) {
	config := Config{Endpoint: "127.0.0.1:9000", AccessKey: "key", SecretKey: "secret", Bucket: "assets"}
	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateRequiresBucket(t *testing.T) {
	config := Config{Endpoint: "127.0.0.1:9000", AccessKey: "key", SecretKey: "secret"}
	if err := config.Validate(); err == nil {
		t.Fatal("Validate() error = nil")
	}
}

func TestConfigValidateRequiresCredentials(t *testing.T) {
	config := Config{Endpoint: "127.0.0.1:9000", Bucket: "assets"}
	if err := config.Validate(); err == nil {
		t.Fatal("Validate() error = nil")
	}
}
