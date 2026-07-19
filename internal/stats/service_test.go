package stats

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
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

func TestServiceNitroCountsEveryCategory(t *testing.T) {
	storage := &fakeStorage{counts: map[string]int{
		"avatar/clothing/":   100,
		"avatar/effects/":    20,
		"furniture/bundles/": 5000,
		"pets/":              12,
	}}
	svc := newService(storage)

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
	svc := newService(storage)

	if _, err := svc.Nitro(context.Background()); err == nil {
		t.Fatal("Nitro() error = nil")
	}
}

func TestServiceNitroRunsConcurrently(t *testing.T) {
	const latency = 20 * time.Millisecond
	storage := &fakeStorage{counts: map[string]int{}, latency: latency}
	svc := newService(storage)

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

func BenchmarkServiceNitro(b *testing.B) {
	storage := &fakeStorage{counts: map[string]int{
		"avatar/clothing/":   100,
		"avatar/effects/":    20,
		"furniture/bundles/": 5000,
		"pets/":              12,
	}}
	svc := newService(storage)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Nitro(ctx); err != nil {
			b.Fatalf("Nitro() error = %v", err)
		}
	}
}
