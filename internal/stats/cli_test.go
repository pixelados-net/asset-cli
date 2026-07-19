package stats

import "testing"

func TestNewRealmCommandExposesNitro(t *testing.T) {
	realm := NewRealmCommand()
	if realm.Use != "stats" {
		t.Fatalf("Use = %q", realm.Use)
	}
	names := make(map[string]bool)
	for _, command := range realm.Commands() {
		names[command.Name()] = true
	}
	if !names["nitro"] {
		t.Fatalf("commands = %#v", names)
	}
}
