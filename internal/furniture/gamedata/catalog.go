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

// KindFloor and KindWall name which top-level FurnitureData.json array a
// Definition came from: roomitemtypes or wallitemtypes. Neither array's own
// entries carry a kind field — this is a parse-time tag, not a mapped value.
const (
	KindFloor = "floor"
	KindWall  = "wall"
)

// furnitureItem decodes only the fields Definition needs; encoding/json discards
// everything else instead of building a full in-memory tree of the
// multi-megabyte source file.
type furnitureItem struct {
	Classname    string `json:"classname"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	XDim         int    `json:"xdim"`
	YDim         int    `json:"ydim"`
	CanStandOn   bool   `json:"canstandon"`
	CanSitOn     bool   `json:"cansiton"`
	CanLayOn     bool   `json:"canlayon"`
	CustomParams string `json:"customparams"`
}

type furnitureData struct {
	RoomItemTypes struct {
		FurniType []furnitureItem `json:"furnitype"`
	} `json:"roomitemtypes"`
	WallItemTypes struct {
		FurniType []furnitureItem `json:"furnitype"`
	} `json:"wallitemtypes"`
}

// Definition is one classname's full furniture-catalog record, as read from the
// client's FurnitureData.json.
type Definition struct {
	// Classname is the stable technical identifier (the sync key).
	Classname string
	// Kind is KindFloor or KindWall.
	Kind string
	// Name is the client's display name (see docs/wiki for why this is real text,
	// not a localization key, for most items).
	Name string
	// Description is the client's display description.
	Description string
	// Width is the footprint width (xdim).
	Width int
	// Length is the footprint length (ydim).
	Length int
	// CanStandOn reports whether units can stand on the item.
	CanStandOn bool
	// CanSitOn reports whether the item produces a sit status.
	CanSitOn bool
	// CanLayOn reports whether the item produces a lay status.
	CanLayOn bool
	// CustomParams stores deferred item-specific parameters.
	CustomParams string
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

// ListDefinitions decodes FurnitureData.json and returns every declared
// definition from both the floor (roomitemtypes) and wall (wallitemtypes) arrays.
func (catalog *Catalog) ListDefinitions(ctx context.Context) ([]Definition, error) {
	reader, err := catalog.client.Get(ctx, furnitureDataKey)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var data furnitureData
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode furniture data: %w", err)
	}

	definitions := make([]Definition, 0, len(data.RoomItemTypes.FurniType)+len(data.WallItemTypes.FurniType))
	for _, item := range data.RoomItemTypes.FurniType {
		if item.Classname != "" {
			definitions = append(definitions, toDefinition(item, KindFloor))
		}
	}
	for _, item := range data.WallItemTypes.FurniType {
		if item.Classname != "" {
			definitions = append(definitions, toDefinition(item, KindWall))
		}
	}
	return definitions, nil
}

// ListClassnames decodes FurnitureData.json and returns every declared classname.
func (catalog *Catalog) ListClassnames(ctx context.Context) ([]string, error) {
	definitions, err := catalog.ListDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	classnames := make([]string, len(definitions))
	for i, definition := range definitions {
		classnames[i] = definition.Classname
	}
	return classnames, nil
}

func toDefinition(item furnitureItem, kind string) Definition {
	return Definition{
		Classname:    item.Classname,
		Kind:         kind,
		Name:         item.Name,
		Description:  item.Description,
		Width:        item.XDim,
		Length:       item.YDim,
		CanStandOn:   item.CanStandOn,
		CanSitOn:     item.CanSitOn,
		CanLayOn:     item.CanLayOn,
		CustomParams: item.CustomParams,
	}
}
