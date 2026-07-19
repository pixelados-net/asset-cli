package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	var output bytes.Buffer
	root := NewRootCommand("0.0.1")
	root.SetOut(&output)
	root.SetArgs([]string{"version"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.String() != "asset-cli v0.0.1\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRootCommandRejectsUnknownCommand(t *testing.T) {
	root := NewRootCommand("0.0.1")
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"unknown"})
	err := root.ExecuteContext(context.Background())
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRootCommandExposesEveryRealm(t *testing.T) {
	root := NewRootCommand("0.0.1")
	names := make(map[string]bool)
	for _, command := range root.Commands() {
		names[command.Name()] = true
	}
	for _, name := range []string{"clothing", "effects", "furniture", "pets", "stats", "structure", "sync"} {
		if !names[name] {
			t.Fatalf("commands = %#v, missing %q", names, name)
		}
	}
}
