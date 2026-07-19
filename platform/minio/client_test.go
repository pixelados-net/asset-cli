package minio

import "testing"

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("New() error = nil")
	}
}

func TestNewBuildsClientWithoutNetworkIO(t *testing.T) {
	config := Config{Endpoint: "127.0.0.1:9000", AccessKey: "key", SecretKey: "secret", Bucket: "assets"}
	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client.Bucket() != "assets" {
		t.Fatalf("Bucket() = %q", client.Bucket())
	}
	if client.SDK() == nil {
		t.Fatal("SDK() = nil")
	}
}
