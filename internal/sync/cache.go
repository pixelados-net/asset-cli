package sync

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/pixelados-net/asset-cli/platform/redis"
)

// clientDefinitionsCacheKey caches the parsed client catalog between commands so
// repeated check/apply/publish runs in a short window skip re-downloading and
// re-parsing the multi-megabyte FurnitureData.json.
const clientDefinitionsCacheKey = "asset-cli:sync:furniture:client-definitions"

// clientDefinitionsCacheTTL bounds how long a cached parse is trusted before a
// run re-fetches the file, so an uploaded gamedata change is eventually observed.
const clientDefinitionsCacheTTL = 5 * time.Minute

// applyCursorKey stores the last classname Apply successfully inserted, so an
// interrupted run's next invocation skips definitions already written instead of
// re-attempting them. It is an optimization only: Apply always recomputes the
// missing set from the emulator's current state, so a stale or missing cursor
// never causes incorrect results, only redundant work.
const applyCursorKey = "asset-cli:sync:furniture:apply:cursor"

// redisStore is the subset of Redis operations the cache/cursor depend on,
// reduced to plain Go types so it is fakeable without a real Redis server.
type redisStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// goredisStore adapts *goredis.Client to the redisStore interface.
type goredisStore struct {
	client *goredis.Client
}

func (wrapper goredisStore) Get(ctx context.Context, key string) (string, error) {
	return wrapper.client.Get(ctx, key).Result()
}

func (wrapper goredisStore) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return wrapper.client.Set(ctx, key, value, expiration).Err()
}

func (wrapper goredisStore) Del(ctx context.Context, keys ...string) error {
	return wrapper.client.Del(ctx, keys...).Err()
}

type cachedClientCatalog struct {
	catalog ClientCatalog
	store   redisStore
}

// NewCachedClientCatalog wraps catalog with a Redis cache.
func NewCachedClientCatalog(catalog ClientCatalog, client *redis.Client) ClientCatalog {
	return newCachedClientCatalog(catalog, goredisStore{client: client.SDK()})
}

func newCachedClientCatalog(catalog ClientCatalog, store redisStore) *cachedClientCatalog {
	return &cachedClientCatalog{catalog: catalog, store: store}
}

func (cached *cachedClientCatalog) ListDefinitions(ctx context.Context) ([]Definition, error) {
	if body, err := cached.store.Get(ctx, clientDefinitionsCacheKey); err == nil {
		var definitions []Definition
		if json.Unmarshal([]byte(body), &definitions) == nil {
			return definitions, nil
		}
	}

	definitions, err := cached.catalog.ListDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	if body, encodeErr := json.Marshal(definitions); encodeErr == nil {
		_ = cached.store.Set(ctx, clientDefinitionsCacheKey, body, clientDefinitionsCacheTTL)
	}
	return definitions, nil
}

// cursor persists Apply's resume point in Redis.
type cursor struct {
	store redisStore
}

func newCursor(client *redis.Client) *cursor {
	return &cursor{store: goredisStore{client: client.SDK()}}
}

// Get returns the last confirmed classname, or "" if there is none.
func (c *cursor) Get(ctx context.Context) (string, error) {
	value, err := c.store.Get(ctx, applyCursorKey)
	if errors.Is(err, goredis.Nil) {
		return "", nil
	}
	return value, err
}

// Set records the last confirmed classname.
func (c *cursor) Set(ctx context.Context, classname string) error {
	return c.store.Set(ctx, applyCursorKey, classname, 0)
}

// Clear removes the cursor once a run completes every batch.
func (c *cursor) Clear(ctx context.Context) error {
	return c.store.Del(ctx, applyCursorKey)
}
