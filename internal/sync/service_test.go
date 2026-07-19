package sync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

type fakeClientCatalog struct {
	definitions []Definition
	err         error
	latency     time.Duration
}

func (fake *fakeClientCatalog) ListDefinitions(context.Context) ([]Definition, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	if fake.err != nil {
		return nil, fake.err
	}
	return fake.definitions, nil
}

type fakeEmulatorCatalog struct {
	mutex       sync.Mutex
	definitions []Definition
	listErr     error
	insertErr   error
	updateErr   error
	inserted    []Definition
	updated     []Definition
	latency     time.Duration
}

func (fake *fakeEmulatorCatalog) ListDefinitions(context.Context) ([]Definition, error) {
	if fake.latency > 0 {
		time.Sleep(fake.latency)
	}
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	if fake.listErr != nil {
		return nil, fake.listErr
	}
	result := make([]Definition, len(fake.definitions))
	copy(result, fake.definitions)
	return result, nil
}

func (fake *fakeEmulatorCatalog) InsertDefinitions(_ context.Context, definitions []Definition) error {
	if fake.insertErr != nil {
		return fake.insertErr
	}
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.inserted = append(fake.inserted, definitions...)
	return nil
}

func (fake *fakeEmulatorCatalog) UpdateDefinitions(_ context.Context, definitions []Definition) error {
	if fake.updateErr != nil {
		return fake.updateErr
	}
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.updated = append(fake.updated, definitions...)
	return nil
}

type fakeCursor struct {
	mutex sync.Mutex
	value string
}

func (fake *fakeCursor) Get(context.Context) (string, error) {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	return fake.value, nil
}

func (fake *fakeCursor) Set(_ context.Context, classname string) error {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.value = classname
	return nil
}

func (fake *fakeCursor) Clear(context.Context) error {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.value = ""
	return nil
}

func TestServiceCheckDiff(t *testing.T) {
	client := &fakeClientCatalog{definitions: []Definition{
		{Classname: "a", PublicName: "A"},
		{Classname: "b", PublicName: "B"},
	}}
	emulator := &fakeEmulatorCatalog{definitions: []Definition{
		{Classname: "a", PublicName: "A-old"},
		{Classname: "orphan", PublicName: "Orphan"},
	}}
	svc := newService(client, emulator, &fakeCursor{})

	report, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(report.Missing) != 1 || report.Missing[0] != "b" {
		t.Fatalf("Missing = %#v", report.Missing)
	}
	if len(report.Orphaned) != 1 || report.Orphaned[0] != "orphan" {
		t.Fatalf("Orphaned = %#v", report.Orphaned)
	}
	if len(report.NameChanges) != 1 || report.NameChanges[0].Classname != "a" ||
		report.NameChanges[0].ClientName != "A" || report.NameChanges[0].EmulatorName != "A-old" {
		t.Fatalf("NameChanges = %#v", report.NameChanges)
	}
	if report.OK() {
		t.Fatal("OK() = true, want false")
	}
}

func TestServiceCheckNormalizesColorIndexVariants(t *testing.T) {
	client := &fakeClientCatalog{definitions: []Definition{
		{Classname: "item*0", PublicName: "Item"},
		{Classname: "item*1", PublicName: "Item"},
	}}
	emulator := &fakeEmulatorCatalog{definitions: []Definition{{Classname: "item", PublicName: "Item"}}}
	svc := newService(client, emulator, &fakeCursor{})

	report, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !report.OK() || len(report.Missing) != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestServiceCheckPropagatesErrors(t *testing.T) {
	svc := newService(&fakeClientCatalog{err: errors.New("client failed")}, &fakeEmulatorCatalog{}, &fakeCursor{})
	if _, err := svc.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil")
	}

	svc = newService(&fakeClientCatalog{}, &fakeEmulatorCatalog{listErr: errors.New("emulator failed")}, &fakeCursor{})
	if _, err := svc.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil")
	}
}

func TestServiceCheckRunsConcurrently(t *testing.T) {
	const latency = 30 * time.Millisecond
	svc := newService(&fakeClientCatalog{latency: latency}, &fakeEmulatorCatalog{latency: latency}, &fakeCursor{})

	start := time.Now()
	if _, err := svc.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	elapsed := time.Since(start)
	if sequential := 2 * latency; elapsed >= sequential {
		t.Fatalf("Check() took %v, want well under the sequential bound %v (not running concurrently?)", elapsed, sequential)
	}
}

func TestServiceApplyInsertsMissingAndUpdatesChanged(t *testing.T) {
	client := &fakeClientCatalog{definitions: []Definition{
		{Classname: "a", PublicName: "A-new"},
		{Classname: "b", PublicName: "B"},
	}}
	emulator := &fakeEmulatorCatalog{definitions: []Definition{{Classname: "a", PublicName: "A-old"}}}
	svc := newService(client, emulator, &fakeCursor{})

	result, err := svc.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Created) != 1 || result.Created[0] != "b" {
		t.Fatalf("Created = %#v", result.Created)
	}
	if len(result.Updated) != 1 || result.Updated[0] != "a" {
		t.Fatalf("Updated = %#v", result.Updated)
	}
	if len(emulator.inserted) != 1 || emulator.inserted[0].Classname != "b" {
		t.Fatalf("inserted = %#v", emulator.inserted)
	}
	if len(emulator.updated) != 1 || emulator.updated[0].PublicName != "A-new" {
		t.Fatalf("updated = %#v, want the client's naming written to the emulator", emulator.updated)
	}
}

func TestServiceApplyBatchesLargeInserts(t *testing.T) {
	const total = 1200 // spans multiple writeBatchSize (500) batches
	definitions := make([]Definition, total)
	for i := range definitions {
		definitions[i] = Definition{Classname: fmt.Sprintf("item_%04d", i), PublicName: "Item"}
	}
	client := &fakeClientCatalog{definitions: definitions}
	emulator := &fakeEmulatorCatalog{}
	svc := newService(client, emulator, &fakeCursor{})

	result, err := svc.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Created) != total || len(emulator.inserted) != total {
		t.Fatalf("created = %d, inserted = %d, want %d", len(result.Created), len(emulator.inserted), total)
	}
}

func TestServiceApplyBatchesLargeUpdates(t *testing.T) {
	const total = 1200
	clientDefs := make([]Definition, total)
	emulatorDefs := make([]Definition, total)
	for i := range clientDefs {
		name := fmt.Sprintf("item_%04d", i)
		clientDefs[i] = Definition{Classname: name, PublicName: "New"}
		emulatorDefs[i] = Definition{Classname: name, PublicName: "Old"}
	}
	client := &fakeClientCatalog{definitions: clientDefs}
	emulator := &fakeEmulatorCatalog{definitions: emulatorDefs}
	svc := newService(client, emulator, &fakeCursor{})

	result, err := svc.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Updated) != total || len(emulator.updated) != total {
		t.Fatalf("updated = %d, want %d", len(result.Updated), total)
	}
}

func TestServiceApplyResumesInsertFromCursor(t *testing.T) {
	definitions := []Definition{
		{Classname: "a"}, {Classname: "b"}, {Classname: "c"},
	}
	client := &fakeClientCatalog{definitions: definitions}
	emulator := &fakeEmulatorCatalog{}
	cursor := &fakeCursor{value: "b"}
	svc := newService(client, emulator, cursor)

	result, err := svc.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Created) != 1 || result.Created[0] != "c" {
		t.Fatalf("Created = %#v, want only classnames after cursor", result.Created)
	}
}

func TestServiceApplyClearsCursorOnFullSuccess(t *testing.T) {
	client := &fakeClientCatalog{definitions: []Definition{{Classname: "a"}}}
	emulator := &fakeEmulatorCatalog{}
	cursor := &fakeCursor{}
	svc := newService(client, emulator, cursor)

	if _, err := svc.Apply(context.Background()); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if value, _ := cursor.Get(context.Background()); value != "" {
		t.Fatalf("cursor = %q, want cleared", value)
	}
}

func TestServiceApplyPropagatesInsertError(t *testing.T) {
	client := &fakeClientCatalog{definitions: []Definition{{Classname: "a"}}}
	emulator := &fakeEmulatorCatalog{insertErr: errors.New("insert failed")}
	svc := newService(client, emulator, &fakeCursor{})

	if _, err := svc.Apply(context.Background()); err == nil {
		t.Fatal("Apply() error = nil")
	}
}

func TestServiceApplyPropagatesUpdateError(t *testing.T) {
	client := &fakeClientCatalog{definitions: []Definition{{Classname: "a", PublicName: "New"}}}
	emulator := &fakeEmulatorCatalog{
		definitions: []Definition{{Classname: "a", PublicName: "Old"}},
		updateErr:   errors.New("update failed"),
	}
	svc := newService(client, emulator, &fakeCursor{})

	if _, err := svc.Apply(context.Background()); err == nil {
		t.Fatal("Apply() error = nil")
	}
}

func TestServiceApplyNoopWhenNothingChanged(t *testing.T) {
	client := &fakeClientCatalog{definitions: []Definition{{Classname: "a", PublicName: "Same"}}}
	emulator := &fakeEmulatorCatalog{definitions: []Definition{{Classname: "a", PublicName: "Same"}}}
	svc := newService(client, emulator, &fakeCursor{})

	result, err := svc.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(result.Created) != 0 || len(result.Updated) != 0 {
		t.Fatalf("result = %#v, want no-op", result)
	}
}

func TestBaseClassname(t *testing.T) {
	cases := map[string]string{
		"yordi_val_c24_catplushie*9": "yordi_val_c24_catplushie",
		"item*0":                     "item",
		"no_index_item":              "no_index_item",
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
	if (Report{Missing: []string{"x"}}).OK() {
		t.Fatal("OK() = true with missing entries")
	}
}

func BenchmarkServiceCheck(b *testing.B) {
	const size = 80000
	definitions := make([]Definition, size)
	for i := range definitions {
		definitions[i] = Definition{Classname: fmt.Sprintf("item_%d", i), PublicName: "Item"}
	}
	client := &fakeClientCatalog{definitions: definitions}
	emulator := &fakeEmulatorCatalog{definitions: definitions}
	svc := newService(client, emulator, &fakeCursor{})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Check(ctx); err != nil {
			b.Fatalf("Check() error = %v", err)
		}
	}
}
