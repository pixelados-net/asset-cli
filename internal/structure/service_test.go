package structure

import (
	"context"
	"errors"
	"testing"
)

type fakeStorage struct {
	existing map[string]bool
	touched  []string
	touchErr error
}

func (storage *fakeStorage) Exists(_ context.Context, prefix string) (bool, error) {
	return storage.existing[prefix], nil
}

func (storage *fakeStorage) Touch(_ context.Context, key string) error {
	if storage.touchErr != nil {
		return storage.touchErr
	}
	storage.touched = append(storage.touched, key)
	return nil
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

func TestServiceCreateFillsOnlyMissingPaths(t *testing.T) {
	storage := &fakeStorage{existing: map[string]bool{ExpectedPaths[0]: true}}
	svc := newService(storage)

	created, err := svc.Create(context.Background())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(created) != len(ExpectedPaths)-1 {
		t.Fatalf("created = %#v", created)
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
}
