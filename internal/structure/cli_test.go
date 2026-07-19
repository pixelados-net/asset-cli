package structure

import "testing"

func TestNewRealmCommandExposesCheckAndCreate(t *testing.T) {
	realm := NewRealmCommand()
	if realm.Use != "structure" {
		t.Fatalf("Use = %q", realm.Use)
	}
	names := make(map[string]bool)
	for _, command := range realm.Commands() {
		names[command.Name()] = true
	}
	if !names["check"] || !names["create"] {
		t.Fatalf("commands = %#v", names)
	}
}
