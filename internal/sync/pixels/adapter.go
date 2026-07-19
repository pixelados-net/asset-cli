// Package pixels implements the sync realm's EmulatorCatalog against the Pixels
// furniture_definitions PostgreSQL table.
package pixels

import (
	"context"
	"fmt"
	"strings"

	syncrealm "github.com/pixelados-net/asset-cli/internal/sync"
	"github.com/pixelados-net/asset-cli/platform/pixels"
)

// defaultInteractionType is the value new rows get; see plan/EMULATORS.md section
// 2c for why sync never derives or overwrites this from the client.
const defaultInteractionType = "default"

// insertColumns lists the columns Adapter writes for a new definition. sprite_id
// is not present in modern FurnitureData.json (a Flash-sprite-sheet-era concept)
// and always gets 0.
const insertColumns = "name, public_name, description, kind, width, length, allow_walk, allow_sit, allow_lay, sprite_id, interaction_type"

const insertColumnsPerRow = 11

const listDefinitionsSQL = "select name, kind, public_name, description, width, length, allow_walk, allow_sit, allow_lay " +
	"from furniture_definitions where deleted_at is null"

// Adapter implements sync.EmulatorCatalog against Pixels' furniture_definitions table.
type Adapter struct {
	pool pool
}

type pool interface {
	Query(ctx context.Context, sql string, args ...any) (rows, error)
	Exec(ctx context.Context, sql string, args ...any) error
}

// New creates a Pixels-backed EmulatorCatalog.
func New(client *pixels.Client) syncrealm.EmulatorCatalog {
	return newAdapter(pgxPool{client.Pool()})
}

func newAdapter(pool pool) *Adapter {
	return &Adapter{pool: pool}
}

// ListDefinitions reads every active furniture definition from furniture_definitions.
func (adapter *Adapter) ListDefinitions(ctx context.Context) ([]syncrealm.Definition, error) {
	rows, err := adapter.pool.Query(ctx, listDefinitionsSQL)
	if err != nil {
		return nil, fmt.Errorf("list furniture_definitions: %w", err)
	}
	defer rows.Close()

	var definitions []syncrealm.Definition
	for rows.Next() {
		var definition syncrealm.Definition
		if err := rows.Scan(&definition.Classname, &definition.Kind, &definition.PublicName, &definition.Description,
			&definition.Width, &definition.Length, &definition.AllowWalk, &definition.AllowSit, &definition.AllowLay); err != nil {
			return nil, fmt.Errorf("scan furniture_definitions row: %w", err)
		}
		definitions = append(definitions, definition)
	}
	return definitions, rows.Err()
}

// InsertDefinitions batch-inserts new furniture_definitions rows. Callers only
// ever pass classnames already confirmed missing.
func (adapter *Adapter) InsertDefinitions(ctx context.Context, definitions []syncrealm.Definition) error {
	if len(definitions) == 0 {
		return nil
	}
	query, args := buildInsertQuery(definitions)
	if err := adapter.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert furniture_definitions rows: %w", err)
	}
	return nil
}

// UpdateDefinitions batch-updates public_name/description on existing
// furniture_definitions rows using a single VALUES-based statement.
func (adapter *Adapter) UpdateDefinitions(ctx context.Context, definitions []syncrealm.Definition) error {
	if len(definitions) == 0 {
		return nil
	}
	query, args := buildUpdateQuery(definitions)
	if err := adapter.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("update furniture_definitions rows: %w", err)
	}
	return nil
}

// buildInsertQuery builds the batched multi-row INSERT statement and its
// positional args for definitions. Pure and side-effect-free so it is unit
// testable without a database connection.
func buildInsertQuery(definitions []syncrealm.Definition) (string, []any) {
	placeholders := make([]string, len(definitions))
	args := make([]any, 0, len(definitions)*insertColumnsPerRow)
	for i, definition := range definitions {
		placeholders[i] = positionalGroup(i*insertColumnsPerRow, insertColumnsPerRow)
		args = append(args, definition.Classname, definition.PublicName, definition.Description, definition.Kind,
			definition.Width, definition.Length, definition.AllowWalk, definition.AllowSit, definition.AllowLay,
			0, defaultInteractionType)
	}
	query := fmt.Sprintf("insert into furniture_definitions (%s) values %s", insertColumns, strings.Join(placeholders, ", "))
	return query, args
}

// updateColumnsPerRow is (name, public_name, description) per VALUES row.
const updateColumnsPerRow = 3

// buildUpdateQuery builds a single bulk UPDATE ... FROM (VALUES ...) statement
// that sets public_name/description per classname in one round trip. Pure and
// side-effect-free so it is unit testable without a database connection.
func buildUpdateQuery(definitions []syncrealm.Definition) (string, []any) {
	placeholders := make([]string, len(definitions))
	args := make([]any, 0, len(definitions)*updateColumnsPerRow)
	for i, definition := range definitions {
		placeholders[i] = positionalGroup(i*updateColumnsPerRow, updateColumnsPerRow)
		args = append(args, definition.Classname, definition.PublicName, definition.Description)
	}
	query := "update furniture_definitions as t set public_name = v.public_name, description = v.description, " +
		"updated_at = now(), version = t.version + 1 from (values " + strings.Join(placeholders, ", ") +
		") as v(name, public_name, description) where t.name = v.name and t.deleted_at is null"
	return query, args
}

// positionalGroup builds a parenthesized "($base+1, $base+2, ..., $base+count)" placeholder group.
func positionalGroup(base, count int) string {
	params := make([]string, count)
	for i := range params {
		params[i] = fmt.Sprintf("$%d", base+i+1)
	}
	return "(" + strings.Join(params, ", ") + ")"
}
