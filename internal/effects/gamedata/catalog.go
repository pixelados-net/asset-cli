// Package gamedata reads effect libraries from gamedata/EffectMap.json.
package gamedata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

// effectMapKey is the bucket key holding the avatar effect map.
const effectMapKey = "gamedata/EffectMap.json"

type effectMap struct {
	Effects []struct {
		Library string `json:"lib"`
	} `json:"effects"`
}

type objectReader interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

// Catalog reads effect library names from EffectMap.json.
type Catalog struct{ client objectReader }

// New creates an EffectMap-backed effects catalog.
func New(client *minio.Client) *Catalog { return newCatalog(client) }

func newCatalog(client objectReader) *Catalog { return &Catalog{client: client} }

// ListNames returns every non-empty library declared by EffectMap.json.
func (catalog *Catalog) ListNames(ctx context.Context) ([]string, error) {
	reader, err := catalog.client.Get(ctx, effectMapKey)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var data effectMap
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode effect map: %w", err)
	}
	names := make([]string, 0, len(data.Effects))
	for _, effect := range data.Effects {
		if effect.Library != "" {
			names = append(names, effect.Library)
		}
	}
	return names, nil
}
