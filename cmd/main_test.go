package main

import "testing"

func TestVersion(t *testing.T) {
	if version != "0.0.1" {
		t.Fatalf("version = %q", version)
	}
}
