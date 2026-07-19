package arcturus

import "testing"

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("New() error = nil")
	}
}

func TestNewBuildsClientWithoutNetworkIO(t *testing.T) {
	client, err := New(Config{Host: "127.0.0.1", Port: 3306, Database: "comet", User: "root"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client.DB() == nil {
		t.Fatal("DB() = nil")
	}
}
