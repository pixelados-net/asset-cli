package sync

import (
	"context"

	"github.com/pixelados-net/asset-cli/internal/furniture/gamedata"
	"github.com/pixelados-net/asset-cli/platform/minio"
)

type clientCatalog struct {
	catalog *gamedata.Catalog
}

// NewClientCatalog creates a ClientCatalog backed by the client's
// gamedata/FurnitureData.json.
func NewClientCatalog(client *minio.Client) ClientCatalog {
	return &clientCatalog{catalog: gamedata.New(client)}
}

func (catalog *clientCatalog) ListDefinitions(ctx context.Context) ([]Definition, error) {
	definitions, err := catalog.catalog.ListDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	return mapGamedataDefinitions(definitions), nil
}

// mapGamedataDefinitions maps the client's gamedata.Definition shape to sync's
// own Definition shape. Pure and side-effect-free so it is unit testable without
// a MinIO client.
func mapGamedataDefinitions(definitions []gamedata.Definition) []Definition {
	result := make([]Definition, len(definitions))
	for i, definition := range definitions {
		result[i] = Definition{
			Classname:   definition.Classname,
			Kind:        definition.Kind,
			PublicName:  definition.Name,
			Description: definition.Description,
			Width:       definition.Width,
			Length:      definition.Length,
			AllowWalk:   definition.CanStandOn,
			AllowSit:    definition.CanSitOn,
			AllowLay:    definition.CanLayOn,
		}
	}
	return result
}
