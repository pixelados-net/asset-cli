// Package gamedata reads clothing libraries from gamedata/FigureMap.json.
package gamedata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

// figureMapKey is the bucket key holding the avatar library map.
const figureMapKey = "gamedata/FigureMap.json"

type figureMap struct {
	Libraries []struct {
		ID string `json:"id"`
	} `json:"libraries"`
}

type objectReader interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

// Catalog reads clothing library names from FigureMap.json.
type Catalog struct{ client objectReader }

// New creates a FigureMap-backed clothing catalog.
func New(client *minio.Client) *Catalog { return newCatalog(client) }

func newCatalog(client objectReader) *Catalog { return &Catalog{client: client} }

// ListNames returns every non-empty library ID declared by FigureMap.json.
func (catalog *Catalog) ListNames(ctx context.Context) ([]string, error) {
	reader, err := catalog.client.Get(ctx, figureMapKey)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var data figureMap
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode figure map: %w", err)
	}
	names := make([]string, 0, len(data.Libraries))
	for _, library := range data.Libraries {
		if library.ID != "" {
			names = append(names, library.ID)
		}
	}
	return names, nil
}
