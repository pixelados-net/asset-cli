package arcturus

import "testing"

func TestConfigValidateRequiresDatabase(t *testing.T) {
	if err := (Config{}).Validate(); err == nil {
		t.Fatal("Validate() error = nil")
	}
}

func TestConfigValidate(t *testing.T) {
	config := Config{Host: "127.0.0.1", Port: 3306, Database: "comet", User: "root"}
	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
