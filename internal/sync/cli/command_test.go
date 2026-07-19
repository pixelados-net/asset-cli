package cli

import "testing"

func TestNewRealmCommandExposesFurnitureSubcommands(t *testing.T) {
	realm := NewRealmCommand()
	if realm.Use != "sync" {
		t.Fatalf("Use = %q", realm.Use)
	}

	furnitureCommands := realm.Commands()
	if len(furnitureCommands) != 1 || furnitureCommands[0].Name() != "furniture" {
		t.Fatalf("commands = %#v", furnitureCommands)
	}

	names := make(map[string]bool)
	for _, command := range furnitureCommands[0].Commands() {
		names[command.Name()] = true
	}
	if !names["check"] || !names["apply"] {
		t.Fatalf("furniture subcommands = %#v", names)
	}
}

func TestApplyCommandDefaultsToDryRun(t *testing.T) {
	command := newApplyCommand()
	flag := command.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("apply command has no --yes flag")
	}
	if flag.DefValue != "false" {
		t.Fatalf("--yes default = %q, want \"false\" (apply must default to dry-run)", flag.DefValue)
	}
}
