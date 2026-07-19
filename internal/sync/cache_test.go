package sync

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type fakeRedisStore struct {
	mutex  sync.Mutex
	values map[string]string
}

func newFakeRedisStore() *fakeRedisStore {
	return &fakeRedisStore{values: map[string]string{}}
}

func (fake *fakeRedisStore) Get(_ context.Context, key string) (string, error) {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	value, ok := fake.values[key]
	if !ok {
		return "", goredis.Nil
	}
	return value, nil
}

func (fake *fakeRedisStore) Set(_ context.Context, key string, value any, _ time.Duration) error {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	switch typed := value.(type) {
	case string:
		fake.values[key] = typed
	case []byte:
		fake.values[key] = string(typed)
	default:
		return errors.New("unsupported value type")
	}
	return nil
}

func (fake *fakeRedisStore) Del(_ context.Context, keys ...string) error {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	for _, key := range keys {
		delete(fake.values, key)
	}
	return nil
}

func TestCachedClientCatalogCachesBetweenCalls(t *testing.T) {
	calls := 0
	inner := &countingCatalog{definitions: []Definition{{Classname: "a"}}, calls: &calls}
	cached := newCachedClientCatalog(inner, newFakeRedisStore())

	for i := 0; i < 3; i++ {
		definitions, err := cached.ListDefinitions(context.Background())
		if err != nil {
			t.Fatalf("ListDefinitions() error = %v", err)
		}
		if len(definitions) != 1 || definitions[0].Classname != "a" {
			t.Fatalf("definitions = %#v", definitions)
		}
	}
	if calls != 1 {
		t.Fatalf("inner ListDefinitions called %d times, want 1 (cache should serve the rest)", calls)
	}
}

type countingCatalog struct {
	definitions []Definition
	calls       *int
}

func (catalog *countingCatalog) ListDefinitions(context.Context) ([]Definition, error) {
	*catalog.calls++
	return catalog.definitions, nil
}

func TestCachedClientCatalogFallsBackOnStoreMiss(t *testing.T) {
	inner := &fakeClientCatalog{definitions: []Definition{{Classname: "a"}}}
	cached := newCachedClientCatalog(inner, newFakeRedisStore())

	definitions, err := cached.ListDefinitions(context.Background())
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(definitions) != 1 {
		t.Fatalf("definitions = %#v", definitions)
	}
}

func TestCursorGetSetClear(t *testing.T) {
	store := newFakeRedisStore()
	c := &cursor{store: store}

	if value, err := c.Get(context.Background()); err != nil || value != "" {
		t.Fatalf("Get() = (%q, %v), want empty with no error", value, err)
	}
	if err := c.Set(context.Background(), "item_5"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if value, err := c.Get(context.Background()); err != nil || value != "item_5" {
		t.Fatalf("Get() = (%q, %v), want %q", value, err, "item_5")
	}
	if err := c.Clear(context.Background()); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if value, err := c.Get(context.Background()); err != nil || value != "" {
		t.Fatalf("Get() after Clear() = (%q, %v), want empty", value, err)
	}
}

func TestCachedClientCatalogPropagatesInnerError(t *testing.T) {
	inner := &fakeClientCatalog{err: errors.New("client failed")}
	cached := newCachedClientCatalog(inner, newFakeRedisStore())
	if _, err := cached.ListDefinitions(context.Background()); err == nil {
		t.Fatal("ListDefinitions() error = nil")
	}
}

func TestCachedClientCatalogSurvivesCorruptCacheEntry(t *testing.T) {
	store := newFakeRedisStore()
	_ = store.Set(context.Background(), clientDefinitionsCacheKey, "not json", 0)
	inner := &fakeClientCatalog{definitions: []Definition{{Classname: "a"}}}
	cached := newCachedClientCatalog(inner, store)

	definitions, err := cached.ListDefinitions(context.Background())
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(definitions) != 1 {
		t.Fatalf("definitions = %#v", definitions)
	}
}

func TestCachedClientCatalogStoresValidJSON(t *testing.T) {
	store := newFakeRedisStore()
	inner := &fakeClientCatalog{definitions: []Definition{{Classname: "a", PublicName: "A"}}}
	cached := newCachedClientCatalog(inner, store)

	if _, err := cached.ListDefinitions(context.Background()); err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	raw, err := store.Get(context.Background(), clientDefinitionsCacheKey)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	var definitions []Definition
	if err := json.Unmarshal([]byte(raw), &definitions); err != nil {
		t.Fatalf("cached value is not valid JSON: %v", err)
	}
	if len(definitions) != 1 || definitions[0].Classname != "a" {
		t.Fatalf("cached definitions = %#v", definitions)
	}
}
