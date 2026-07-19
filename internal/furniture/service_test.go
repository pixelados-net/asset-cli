package furniture

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type fakeBundles struct {
	names   []string
	err     error
	latency time.Duration
}

func (fake *fakeBundles) ListClassnames(context.Context) ([]string, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	if fake.err != nil {
		return nil, fake.err
	}
	return fake.names, nil
}

type fakeCatalog struct {
	names   []string
	err     error
	latency time.Duration
}

func (fake *fakeCatalog) ListClassnames(context.Context) ([]string, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	if fake.err != nil {
		return nil, fake.err
	}
	return fake.names, nil
}

func TestServiceCheckDiff(t *testing.T) {
	cases := []struct {
		name         string
		bundles      []string
		catalog      []string
		wantMatched  int
		wantOrphaned []string
		wantMissing  []string
	}{
		{
			name:         "orphaned only",
			bundles:      []string{"a", "b"},
			catalog:      []string{"a"},
			wantMatched:  1,
			wantOrphaned: []string{"b"},
		},
		{
			name:        "missing only",
			bundles:     []string{"a"},
			catalog:     []string{"a", "b"},
			wantMatched: 1,
			wantMissing: []string{"b"},
		},
		{
			name:         "both",
			bundles:      []string{"a", "orphan"},
			catalog:      []string{"a", "missing"},
			wantMatched:  1,
			wantOrphaned: []string{"orphan"},
			wantMissing:  []string{"missing"},
		},
		{
			name:        "neither",
			bundles:     []string{"a", "b"},
			catalog:     []string{"a", "b"},
			wantMatched: 2,
		},
		{
			name:        "catalog color index variants share one bundle",
			bundles:     []string{"yordi_val_c24_catplushie"},
			catalog:     []string{"yordi_val_c24_catplushie*0", "yordi_val_c24_catplushie*1", "yordi_val_c24_catplushie*2"},
			wantMatched: 1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			svc := newService(&fakeBundles{names: testCase.bundles}, &fakeCatalog{names: testCase.catalog})

			report, err := svc.Check(context.Background())
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if report.Matched != testCase.wantMatched {
				t.Fatalf("Matched = %d, want %d", report.Matched, testCase.wantMatched)
			}
			if !equalStrings(report.Orphaned, testCase.wantOrphaned) {
				t.Fatalf("Orphaned = %#v, want %#v", report.Orphaned, testCase.wantOrphaned)
			}
			if !equalStrings(report.Missing, testCase.wantMissing) {
				t.Fatalf("Missing = %#v, want %#v", report.Missing, testCase.wantMissing)
			}
		})
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestServiceCheckPropagatesBundleError(t *testing.T) {
	svc := newService(&fakeBundles{err: errors.New("list failed")}, &fakeCatalog{})
	if _, err := svc.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil")
	}
}

func TestServiceCheckPropagatesCatalogError(t *testing.T) {
	svc := newService(&fakeBundles{}, &fakeCatalog{err: errors.New("catalog failed")})
	if _, err := svc.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil")
	}
}

func TestServiceCheckRunsConcurrently(t *testing.T) {
	const latency = 30 * time.Millisecond
	svc := newService(&fakeBundles{latency: latency}, &fakeCatalog{latency: latency})

	start := time.Now()
	if _, err := svc.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	elapsed := time.Since(start)

	sequential := 2 * latency
	if elapsed >= sequential {
		t.Fatalf("Check() took %v, want well under the sequential bound %v (not running concurrently?)", elapsed, sequential)
	}
}

func TestBaseClassname(t *testing.T) {
	cases := map[string]string{
		"yordi_val_c24_catplushie*9": "yordi_val_c24_catplushie",
		"item*0":                     "item",
		"no_index_item":              "no_index_item",
		"":                           "",
	}
	for input, want := range cases {
		if got := baseClassname(input); got != want {
			t.Fatalf("baseClassname(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestReportOK(t *testing.T) {
	if !(Report{}).OK() {
		t.Fatal("OK() = false for empty report")
	}
	if !(Report{Orphaned: []string{"x"}}).OK() {
		t.Fatal("OK() = false with only orphaned entries, want true")
	}
	if (Report{Missing: []string{"x"}}).OK() {
		t.Fatal("OK() = true with missing entries")
	}
}

func BenchmarkServiceCheck(b *testing.B) {
	const size = 80000
	bundleNames := make([]string, size)
	catalogNames := make([]string, size)
	for i := 0; i < size; i++ {
		name := fmt.Sprintf("item_%d", i)
		bundleNames[i] = name
		catalogNames[i] = name
	}
	svc := newService(&fakeBundles{names: bundleNames}, &fakeCatalog{names: catalogNames})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Check(ctx); err != nil {
			b.Fatalf("Check() error = %v", err)
		}
	}
}
