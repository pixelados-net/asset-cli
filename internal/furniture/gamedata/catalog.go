// Package gamedata reads the furniture catalog from gamedata/FurnitureData.json.
package gamedata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

// furnitureDataKey is the bucket key holding the furniture catalog.
const furnitureDataKey = "gamedata/FurnitureData.json"

// furnitureData decodes only the fields ListClassnames needs; encoding/json
// discards everything else instead of building a full in-memory tree of the
// multi-megabyte source file.
type furnitureData struct {
	RoomItemTypes struct {
		FurniType []struct {
			Classname string `json:"classname"`
		} `json:"furnitype"`
	} `json:"roomitemtypes"`
}

// objectReader is the subset of the MinIO client Catalog needs.
type objectReader interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

// Catalog reads the furniture catalog from FurnitureData.json in MinIO.
type Catalog struct {
	client objectReader
}

// New creates a gamedata-backed furniture catalog.
func New(client *minio.Client) *Catalog {
	return newCatalog(client)
}

func newCatalog(client objectReader) *Catalog {
	return &Catalog{client: client}
}

// ListClassnames decodes FurnitureData.json and returns every declared classname.
func (catalog *Catalog) ListClassnames(ctx context.Context) ([]string, error) {
	reader, err := catalog.client.Get(ctx, furnitureDataKey)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var data furnitureData
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode furniture data: %w", err)
	}

	classnames := make([]string, 0, len(data.RoomItemTypes.FurniType))
	for _, item := range data.RoomItemTypes.FurniType {
		if item.Classname != "" {
			classnames = append(classnames, item.Classname)
		}
	}
	return classnames, nil
}
