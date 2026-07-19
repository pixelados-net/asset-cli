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

func TestNewAcceptsFullyQualifiedEndpoint(t *testing.T) {
	config := Config{Endpoint: "https://storage.example.com/", AccessKey: "key", SecretKey: "secret", Bucket: "assets"}
	if _, err := New(config); err != nil {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNormalizeEndpoint(t *testing.T) {
	cases := map[string]string{
		"https://storage.example.com":  "storage.example.com",
		"http://127.0.0.1:9000/":       "127.0.0.1:9000",
		"127.0.0.1:9000":               "127.0.0.1:9000",
		" https://storage.example.com": "storage.example.com",
	}
	for input, want := range cases {
		if got := normalizeEndpoint(input); got != want {
			t.Fatalf("normalizeEndpoint(%q) = %q, want %q", input, got, want)
		}
	}
}
