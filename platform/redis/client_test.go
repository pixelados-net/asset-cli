package redis

import "testing"

func TestNewBuildsClientWithoutNetworkIO(t *testing.T) {
	client := New(Config{Address: "127.0.0.1:6379"})
	if client.SDK() == nil {
		t.Fatal("SDK() = nil")
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
