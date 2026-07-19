package pets

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type fakeSource struct {
	names   []string
	err     error
	latency time.Duration
}

func (fake *fakeSource) ListNames(context.Context) ([]string, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	return fake.names, fake.err
}

func TestServiceCheck(t *testing.T) {
	svc := newService(&fakeSource{names: []string{"dog", "custom"}},
		&fakeSource{names: []string{"dog", "cat", "cat"}})
	report, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if report.Matched != 1 || !equalStrings(report.Orphaned, []string{"custom"}) ||
		!equalStrings(report.Missing, []string{"cat"}) {
		t.Fatalf("report = %#v", report)
	}
	if report.OK() {
		t.Fatal("OK() = true with a missing bundle")
	}
}

func TestServiceCheckPropagatesErrors(t *testing.T) {
	svc := newService(&fakeSource{err: errors.New("list failed")}, &fakeSource{})
	if _, err := svc.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil")
	}
}

func TestServiceCheckRunsConcurrently(t *testing.T) {
	const latency = 30 * time.Millisecond
	svc := newService(&fakeSource{latency: latency}, &fakeSource{latency: latency})
	start := time.Now()
	if _, err := svc.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 2*latency {
		t.Fatalf("Check() took %v, want under %v", elapsed, 2*latency)
	}
}

func TestClientCatalogCopiesStandardNames(t *testing.T) {
	catalog := NewClientCatalog()
	names, err := catalog.ListNames(context.Background())
	if err != nil {
		t.Fatalf("ListNames() error = %v", err)
	}
	names[0] = "changed"
	if StandardBundleNames[0] != "dog" || len(StandardBundleNames) != 36 {
		t.Fatalf("StandardBundleNames = %#v", StandardBundleNames)
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

func BenchmarkServiceCheck(b *testing.B) {
	const size = 80000
	names := make([]string, size)
	for i := range names {
		names[i] = fmt.Sprintf("pet_%d", i)
	}
	svc := newService(&fakeSource{names: names}, &fakeSource{names: names})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Check(context.Background()); err != nil {
			b.Fatalf("Check() error = %v", err)
		}
	}
}
