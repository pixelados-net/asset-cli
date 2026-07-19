// Package arcturus implements the sync realm's EmulatorCatalog against the
// Arcturus items_base MySQL table.
package arcturus

import (
	"context"
	"fmt"
	"strings"

	syncrealm "github.com/pixelados-net/asset-cli/internal/sync"
	"github.com/pixelados-net/asset-cli/platform/arcturus"
)

// Arcturus's items_base.type is a legacy free-form code (confirmed against a real
// install: values include "s", "i", "b", "r", "h", "e", "1".."5" — far more than
// floor/wall). This adapter only ever needs to translate the two kinds the client
// declares; it never reads or reinterprets the other codes on existing rows.
const (
	typeFloor = "s"
	typeWall  = "i"
)

// defaultInteractionType is the value new rows get; see plan/EMULATORS.md section
// 2c for why sync never derives or overwrites this from the client.
const defaultInteractionType = "default"

// insertColumns lists the columns Adapter writes for a new definition. sprite_id
// is not present in modern FurnitureData.json (a Flash-sprite-sheet-era concept)
// and always gets 0; description has no column in this schema at all.
const insertColumns = "item_name, public_name, type, width, length, allow_walk, allow_sit, allow_lay, sprite_id, interaction_type"

const insertColumnsPerRow = 10

const listDefinitionsSQL = "SELECT item_name, type, public_name, width, length, allow_walk, allow_sit, allow_lay FROM items_base"

// Adapter implements sync.EmulatorCatalog against Arcturus's items_base table.
type Adapter struct {
	db querier
}

// New creates an Arcturus-backed EmulatorCatalog.
func New(client *arcturus.Client) syncrealm.EmulatorCatalog {
	return newAdapter(sqlDB{db: client.DB()})
}

func newAdapter(db querier) *Adapter {
	return &Adapter{db: db}
}

// ListDefinitions reads every furniture definition from items_base.
func (adapter *Adapter) ListDefinitions(ctx context.Context) ([]syncrealm.Definition, error) {
	rows, err := adapter.db.QueryContext(ctx, listDefinitionsSQL)
	if err != nil {
		return nil, fmt.Errorf("list items_base: %w", err)
	}
	defer rows.Close()

	var definitions []syncrealm.Definition
	for rows.Next() {
		var (
			itemName, itemType, publicName string
			width, length                  int
			allowWalk, allowSit, allowLay  int
		)
		if err := rows.Scan(&itemName, &itemType, &publicName, &width, &length, &allowWalk, &allowSit, &allowLay); err != nil {
			return nil, fmt.Errorf("scan items_base row: %w", err)
		}
		definitions = append(definitions, syncrealm.Definition{
			Classname:  itemName,
			Kind:       kindFromType(itemType),
			PublicName: publicName,
			Width:      width,
			Length:     length,
			// Arcturus's allow_* columns are tinyint but not always a clean 0/1 in
			// real data (some rows carry stray values like 2) — treat non-zero as
			// true rather than scanning straight into bool, which the driver
			// rejects for anything but 0/1.
			AllowWalk: allowWalk != 0,
			AllowSit:  allowSit != 0,
			AllowLay:  allowLay != 0,
		})
	}
	return definitions, rows.Err()
}

// InsertDefinitions batch-inserts new items_base rows. Callers only ever pass
// classnames already confirmed missing.
func (adapter *Adapter) InsertDefinitions(ctx context.Context, definitions []syncrealm.Definition) error {
	if len(definitions) == 0 {
		return nil
	}
	query, args := buildInsertQuery(definitions)
	if err := adapter.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("insert items_base rows: %w", err)
	}
	return nil
}

// UpdateDefinitions batch-updates public_name on existing items_base rows using
// a single CASE-based statement. items_base has no description column, so only
// public_name is writable here; see plan/EMULATORS.md.
func (adapter *Adapter) UpdateDefinitions(ctx context.Context, definitions []syncrealm.Definition) error {
	if len(definitions) == 0 {
		return nil
	}
	query, args := buildUpdateQuery(definitions)
	if err := adapter.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update items_base rows: %w", err)
	}
	return nil
}

func kindFromType(itemType string) string {
	if itemType == typeWall {
		return syncrealm.KindWall
	}
	return syncrealm.KindFloor
}

func typeFromKind(kind string) string {
	if kind == syncrealm.KindWall {
		return typeWall
	}
	return typeFloor
}

// buildInsertQuery builds the batched multi-row INSERT statement and its
// positional args for definitions. Pure and side-effect-free so it is unit
// testable without a database connection.
func buildInsertQuery(definitions []syncrealm.Definition) (string, []any) {
	placeholders := make([]string, len(definitions))
	args := make([]any, 0, len(definitions)*insertColumnsPerRow)
	for i, definition := range definitions {
		placeholders[i] = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		args = append(args, definition.Classname, definition.PublicName, typeFromKind(definition.Kind), definition.Width,
			definition.Length, definition.AllowWalk, definition.AllowSit, definition.AllowLay, 0, defaultInteractionType)
	}
	query := fmt.Sprintf("INSERT INTO items_base (%s) VALUES %s", insertColumns, strings.Join(placeholders, ", "))
	return query, args
}

// buildUpdateQuery builds a single CASE-based UPDATE statement that sets
// public_name per item_name in one round trip. Pure and side-effect-free so it
// is unit testable without a database connection.
func buildUpdateQuery(definitions []syncrealm.Definition) (string, []any) {
	cases := make([]string, len(definitions))
	inPlaceholders := make([]string, len(definitions))
	args := make([]any, 0, len(definitions)*3)
	for i, definition := range definitions {
		cases[i] = "WHEN ? THEN ?"
		inPlaceholders[i] = "?"
		args = append(args, definition.Classname, definition.PublicName)
	}
	for _, definition := range definitions {
		args = append(args, definition.Classname)
	}
	query := fmt.Sprintf(
		"UPDATE items_base SET public_name = CASE item_name %s ELSE public_name END WHERE item_name IN (%s)",
		strings.Join(cases, " "), strings.Join(inPlaceholders, ", "),
	)
	return query, args
}
