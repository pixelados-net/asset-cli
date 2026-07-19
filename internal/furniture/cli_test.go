package furniture

import "testing"

func TestNewRealmCommandExposesCheck(t *testing.T) {
	realm := NewRealmCommand()
	if realm.Use != "furniture" {
		t.Fatalf("Use = %q", realm.Use)
	}
	names := make(map[string]bool)
	for _, command := range realm.Commands() {
		names[command.Name()] = true
	}
	if !names["check"] {
		t.Fatalf("commands = %#v", names)
	}
}
