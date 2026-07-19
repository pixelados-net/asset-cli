package stats

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pixelados-net/asset-cli/internal/furniture"
)

type fakeStorage struct {
	mutex    sync.Mutex
	counts   map[string]int
	countErr error
	latency  time.Duration
}

func (storage *fakeStorage) CountByExtension(_ context.Context, prefix, _ string) (int, error) {
	if storage.latency > 0 {
		time.Sleep(storage.latency)
	}
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	if storage.countErr != nil {
		return 0, storage.countErr
	}
	return storage.counts[prefix], nil
}

type fakeFurnitureChecker struct {
	report furniture.Report
	err    error
}

func (fake *fakeFurnitureChecker) Check(context.Context) (furniture.Report, error) {
	if fake.err != nil {
		return furniture.Report{}, fake.err
	}
	return fake.report, nil
}

func TestServiceNitroCountsEveryCategory(t *testing.T) {
	storage := &fakeStorage{counts: map[string]int{
		"avatar/clothing/":   100,
		"avatar/effects/":    20,
		"furniture/bundles/": 5000,
		"pets/":              12,
	}}
	svc := newService(storage, &fakeFurnitureChecker{})

	counts, err := svc.Nitro(context.Background())
	if err != nil {
		t.Fatalf("Nitro() error = %v", err)
	}
	if len(counts) != len(NitroCategories) {
		t.Fatalf("counts = %#v", counts)
	}
	totals := make(map[string]int, len(counts))
	for _, count := range counts {
		totals[count.Name] = count.Total
	}
	if totals["clothing"] != 100 || totals["furniture"] != 5000 || totals["pets"] != 12 || totals["effects"] != 20 {
		t.Fatalf("totals = %#v", totals)
	}
}

func TestServiceNitroPropagatesError(t *testing.T) {
	storage := &fakeStorage{countErr: errors.New("list failed")}
	svc := newService(storage, &fakeFurnitureChecker{})

	if _, err := svc.Nitro(context.Background()); err == nil {
		t.Fatal("Nitro() error = nil")
	}
}

func TestServiceNitroRunsConcurrently(t *testing.T) {
	const latency = 20 * time.Millisecond
	storage := &fakeStorage{counts: map[string]int{}, latency: latency}
	svc := newService(storage, &fakeFurnitureChecker{})

	start := time.Now()
	if _, err := svc.Nitro(context.Background()); err != nil {
		t.Fatalf("Nitro() error = %v", err)
	}
	elapsed := time.Since(start)

	sequential := time.Duration(len(NitroCategories)) * latency
	if elapsed >= sequential {
		t.Fatalf("Nitro() took %v, want well under the sequential bound %v (not running concurrently?)", elapsed, sequential)
	}
}

func TestServiceOrphansReportsFurnitureCategory(t *testing.T) {
	checker := &fakeFurnitureChecker{report: furniture.Report{
		Matched:  100,
		Orphaned: []string{"a", "b"},
		Missing:  []string{"c"},
	}}
	svc := newService(&fakeStorage{}, checker)

	reports, err := svc.Orphans(context.Background())
	if err != nil {
		t.Fatalf("Orphans() error = %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("reports = %#v", reports)
	}
	report := reports[0]
	if report.Category != "furniture" || report.Matched != 100 || report.Orphaned != 2 || report.Missing != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestServiceOrphansPropagatesError(t *testing.T) {
	svc := newService(&fakeStorage{}, &fakeFurnitureChecker{err: errors.New("check failed")})

	if _, err := svc.Orphans(context.Background()); err == nil {
		t.Fatal("Orphans() error = nil")
	}
}

func BenchmarkServiceNitro(b *testing.B) {
	storage := &fakeStorage{counts: map[string]int{
		"avatar/clothing/":   100,
		"avatar/effects/":    20,
		"furniture/bundles/": 5000,
		"pets/":              12,
	}}
	svc := newService(storage, &fakeFurnitureChecker{})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Nitro(ctx); err != nil {
			b.Fatalf("Nitro() error = %v", err)
		}
	}
}
