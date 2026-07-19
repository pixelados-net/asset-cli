package pixels

import "testing"

func TestConfigValidateRequiresDatabase(t *testing.T) {
	if err := (Config{User: "pixels"}).Validate(); err == nil {
		t.Fatal("Validate() error = nil")
	}
}

func TestConfigValidateRequiresUser(t *testing.T) {
	if err := (Config{Database: "pixels"}).Validate(); err == nil {
		t.Fatal("Validate() error = nil")
	}
}

func TestConfigValidate(t *testing.T) {
	config := Config{Host: "127.0.0.1", Port: 5432, Database: "pixels", User: "pixels", SSLMode: "disable"}
	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
