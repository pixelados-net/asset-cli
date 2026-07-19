package structure

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeStorage struct {
	mutex    sync.Mutex
	existing map[string]bool
	nested   map[string][]string
	touched  []string
	touchErr error
	latency  time.Duration
}

func (storage *fakeStorage) Exists(_ context.Context, prefix string) (bool, error) {
	storage.sleep()
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	return storage.existing[prefix], nil
}

func (storage *fakeStorage) SubPrefixes(_ context.Context, prefix string) ([]string, error) {
	storage.sleep()
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	return storage.nested[prefix], nil
}

func (storage *fakeStorage) Touch(_ context.Context, key string) error {
	storage.sleep()
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	if storage.touchErr != nil {
		return storage.touchErr
	}
	storage.touched = append(storage.touched, key)
	return nil
}

func (storage *fakeStorage) sleep() {
	if storage.latency > 0 {
		time.Sleep(storage.latency)
	}
}

func TestServiceCheckReportsMissingAndPresent(t *testing.T) {
	storage := &fakeStorage{existing: map[string]bool{ExpectedPaths[0]: true}}
	svc := newService(storage)

	report, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(report.Present) != 1 || report.Present[0] != ExpectedPaths[0] {
		t.Fatalf("Present = %#v", report.Present)
	}
	if len(report.Missing) != len(ExpectedPaths)-1 {
		t.Fatalf("Missing = %#v", report.Missing)
	}
	if report.OK() {
		t.Fatal("OK() = true, want false")
	}
}

func TestServiceCheckReportsNestedFolders(t *testing.T) {
	storage := &fakeStorage{
		existing: map[string]bool{},
		nested:   map[string][]string{FlatPaths[0]: {FlatPaths[0] + "bundles/"}},
	}
	svc := newService(storage)

	report, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(report.Nested) != 1 || report.Nested[0] != FlatPaths[0]+"bundles/" {
		t.Fatalf("Nested = %#v", report.Nested)
	}
	if report.OK() {
		t.Fatal("OK() = true, want false with nested folder present")
	}
}

func TestServiceCheckRunsConcurrently(t *testing.T) {
	const latency = 20 * time.Millisecond
	storage := &fakeStorage{existing: map[string]bool{}, latency: latency}
	svc := newService(storage)

	start := time.Now()
	if _, err := svc.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	elapsed := time.Since(start)

	sequential := time.Duration(len(ExpectedPaths)+len(FlatPaths)) * latency
	if elapsed >= sequential {
		t.Fatalf("Check() took %v, want well under the sequential bound %v (not running concurrently?)", elapsed, sequential)
	}
}

func TestServiceCreateFillsOnlyMissingPaths(t *testing.T) {
	storage := &fakeStorage{existing: map[string]bool{ExpectedPaths[0]: true}}
	svc := newService(storage)

	created, err := svc.Create(context.Background())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	wantCreated := 0
	for _, path := range ExpectedPaths[1:] {
		if strings.HasSuffix(path, "/") {
			wantCreated++
		}
	}
	if len(created) != wantCreated {
		t.Fatalf("created = %#v, want %d entries", created, wantCreated)
	}
	if len(storage.touched) != len(created) {
		t.Fatalf("touched = %#v", storage.touched)
	}
	for _, path := range created {
		if path == ExpectedPaths[0] {
			t.Fatalf("Create() recreated already-present path %q", path)
		}
	}
}

func TestServiceCreateNeverFabricatesMissingFiles(t *testing.T) {
	storage := &fakeStorage{existing: map[string]bool{}}
	svc := newService(storage)

	created, err := svc.Create(context.Background())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	for _, path := range created {
		if !strings.HasSuffix(path, "/") {
			t.Fatalf("Create() fabricated a placeholder for exact file key %q", path)
		}
	}
	for _, key := range storage.touched {
		if strings.Contains(key, "gamedata/") && strings.HasSuffix(key, ".json"+placeholderSuffix) {
			t.Fatalf("Create() touched a placeholder for a required gamedata file: %q", key)
		}
	}
}

func TestServiceCreatePropagatesTouchError(t *testing.T) {
	storage := &fakeStorage{existing: map[string]bool{}, touchErr: errors.New("touch failed")}
	svc := newService(storage)

	if _, err := svc.Create(context.Background()); err == nil {
		t.Fatal("Create() error = nil")
	}
}

func TestReportOK(t *testing.T) {
	if !(Report{}).OK() {
		t.Fatal("OK() = false for empty report")
	}
	if (Report{Missing: []string{"x"}}).OK() {
		t.Fatal("OK() = true with missing paths")
	}
	if (Report{Nested: []string{"x"}}).OK() {
		t.Fatal("OK() = true with nested paths")
	}
}

func BenchmarkServiceCheck(b *testing.B) {
	storage := &fakeStorage{existing: map[string]bool{}}
	svc := newService(storage)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Check(ctx); err != nil {
			b.Fatalf("Check() error = %v", err)
		}
	}
}

func BenchmarkServiceCreate(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		storage := &fakeStorage{existing: map[string]bool{}}
		svc := newService(storage)
		b.StartTimer()

		if _, err := svc.Create(ctx); err != nil {
			b.Fatalf("Create() error = %v", err)
		}
	}
}
