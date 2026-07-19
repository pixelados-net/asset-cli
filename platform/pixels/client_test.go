package pixels

import "testing"

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("New() error = nil")
	}
}

func TestNewBuildsClientWithoutNetworkIO(t *testing.T) {
	client, err := New(Config{Host: "127.0.0.1", Port: 5432, Database: "pixels", User: "pixels", SSLMode: "disable"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client.Pool() == nil {
		t.Fatal("Pool() = nil")
	}
	client.Close()
}
