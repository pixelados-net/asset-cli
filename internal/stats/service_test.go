package stats

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/pixelados-net/asset-cli/internal/clothing"
	"github.com/pixelados-net/asset-cli/internal/effects"
	"github.com/pixelados-net/asset-cli/internal/furniture"
	"github.com/pixelados-net/asset-cli/internal/pets"
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
	report  furniture.Report
	err     error
	latency time.Duration
}

func (fake *fakeFurnitureChecker) Check(context.Context) (furniture.Report, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	if fake.err != nil {
		return furniture.Report{}, fake.err
	}
	return fake.report, nil
}

type fakeClothingChecker struct {
	report  clothing.Report
	err     error
	latency time.Duration
}

func (fake *fakeClothingChecker) Check(context.Context) (clothing.Report, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	return fake.report, fake.err
}

type fakeEffectsChecker struct {
	report  effects.Report
	err     error
	latency time.Duration
}

func (fake *fakeEffectsChecker) Check(context.Context) (effects.Report, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	return fake.report, fake.err
}

type fakePetsChecker struct {
	report  pets.Report
	err     error
	latency time.Duration
}

func (fake *fakePetsChecker) Check(context.Context) (pets.Report, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	return fake.report, fake.err
}

func newTestService(storage Storage, furnitureChecker *fakeFurnitureChecker) *service {
	return newService(storage, &fakeClothingChecker{}, &fakeEffectsChecker{}, furnitureChecker, &fakePetsChecker{})
}

func TestServiceNitroCountsEveryCategory(t *testing.T) {
	storage := &fakeStorage{counts: map[string]int{
		"avatar/clothing/":   100,
		"avatar/effects/":    20,
		"furniture/bundles/": 5000,
		"pets/":              12,
	}}
	svc := newTestService(storage, &fakeFurnitureChecker{})

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
	svc := newTestService(storage, &fakeFurnitureChecker{})

	if _, err := svc.Nitro(context.Background()); err == nil {
		t.Fatal("Nitro() error = nil")
	}
}

func TestServiceNitroRunsConcurrently(t *testing.T) {
	const latency = 20 * time.Millisecond
	storage := &fakeStorage{counts: map[string]int{}, latency: latency}
	svc := newTestService(storage, &fakeFurnitureChecker{})

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

func TestServiceOrphansReportsEveryCategory(t *testing.T) {
	svc := newService(&fakeStorage{},
		&fakeClothingChecker{report: clothing.Report{Matched: 10, Orphaned: []string{"a"}}},
		&fakeEffectsChecker{report: effects.Report{Matched: 20, Missing: []string{"b"}}},
		&fakeFurnitureChecker{report: furniture.Report{Matched: 30, Orphaned: []string{"c", "d"}}},
		&fakePetsChecker{report: pets.Report{Matched: 40}},
	)

	reports, err := svc.Orphans(context.Background())
	if err != nil {
		t.Fatalf("Orphans() error = %v", err)
	}
	if len(reports) != 4 {
		t.Fatalf("reports = %#v", reports)
	}
	if reports[0] != (OrphanReport{"clothing", 10, 1, 0}) ||
		reports[1] != (OrphanReport{"effects", 20, 0, 1}) ||
		reports[2] != (OrphanReport{"furniture", 30, 2, 0}) ||
		reports[3] != (OrphanReport{"pets", 40, 0, 0}) {
		t.Fatalf("reports = %#v", reports)
	}
}

func TestServiceOrphansPropagatesError(t *testing.T) {
	svc := newTestService(&fakeStorage{}, &fakeFurnitureChecker{err: errors.New("check failed")})

	if _, err := svc.Orphans(context.Background()); err == nil {
		t.Fatal("Orphans() error = nil")
	}
}

func TestServiceOrphansRunsConcurrently(t *testing.T) {
	const latency = 20 * time.Millisecond
	svc := newService(&fakeStorage{}, &fakeClothingChecker{latency: latency},
		&fakeEffectsChecker{latency: latency}, &fakeFurnitureChecker{latency: latency},
		&fakePetsChecker{latency: latency})
	start := time.Now()
	if _, err := svc.Orphans(context.Background()); err != nil {
		t.Fatalf("Orphans() error = %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 4*latency {
		t.Fatalf("Orphans() took %v, want under %v", elapsed, 4*latency)
	}
}

func BenchmarkServiceNitro(b *testing.B) {
	storage := &fakeStorage{counts: map[string]int{
		"avatar/clothing/":   100,
		"avatar/effects/":    20,
		"furniture/bundles/": 5000,
		"pets/":              12,
	}}
	svc := newTestService(storage, &fakeFurnitureChecker{})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Nitro(ctx); err != nil {
			b.Fatalf("Nitro() error = %v", err)
		}
	}
}
