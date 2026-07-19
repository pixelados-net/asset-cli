package e2e

import "testing"

func TestVersionE2E(t *testing.T) {
	result := runHarness(t, []string{"version"})
	if result.err != nil {
		t.Fatalf("version error = %v, output = %q", result.err, result.output)
	}
	if result.output != "asset-cli v0.0.1\n" {
		t.Fatalf("output = %q", result.output)
	}
}
