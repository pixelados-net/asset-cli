package arcturus

import (
	"context"
	"errors"
	"strings"
	"testing"

	syncrealm "github.com/pixelados-net/asset-cli/internal/sync"
)

type fakeRows struct {
	data [][]any
	idx  int
	err  error
}

func (fake *fakeRows) Next() bool {
	if fake.err != nil {
		return false
	}
	return fake.idx < len(fake.data)
}

func (fake *fakeRows) Scan(dest ...any) error {
	row := fake.data[fake.idx]
	fake.idx++
	*dest[0].(*string) = row[0].(string)
	*dest[1].(*string) = row[1].(string)
	*dest[2].(*string) = row[2].(string)
	*dest[3].(*int) = row[3].(int)
	*dest[4].(*int) = row[4].(int)
	*dest[5].(*int) = row[5].(int)
	*dest[6].(*int) = row[6].(int)
	*dest[7].(*int) = row[7].(int)
	return nil
}

func (fake *fakeRows) Err() error   { return fake.err }
func (fake *fakeRows) Close() error { return nil }

type fakeQuerier struct {
	rows      *fakeRows
	queryErr  error
	execErr   error
	execQuery string
	execArgs  []any
}

func (fake *fakeQuerier) QueryContext(context.Context, string, ...any) (rows, error) {
	if fake.queryErr != nil {
		return nil, fake.queryErr
	}
	return fake.rows, nil
}

func (fake *fakeQuerier) ExecContext(_ context.Context, query string, args ...any) error {
	fake.execQuery = query
	fake.execArgs = args
	return fake.execErr
}

func TestAdapterListDefinitionsMapsTypeToKind(t *testing.T) {
	querier := &fakeQuerier{rows: &fakeRows{data: [][]any{
		{"throne_gold", "s", "Gold Throne", 2, 2, 0, 1, 0},
		{"post.it", "i", "Post-it", 0, 0, 0, 0, 0},
	}}}
	adapter := newAdapter(querier)

	definitions, err := adapter.ListDefinitions(context.Background())
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(definitions) != 2 {
		t.Fatalf("definitions = %#v", definitions)
	}
	if definitions[0].Kind != syncrealm.KindFloor || definitions[0].Classname != "throne_gold" || !definitions[0].AllowSit {
		t.Fatalf("definitions[0] = %#v", definitions[0])
	}
	if definitions[1].Kind != syncrealm.KindWall || definitions[1].Classname != "post.it" {
		t.Fatalf("definitions[1] = %#v", definitions[1])
	}
}

// TestAdapterListDefinitionsTreatsNonZeroAllowFlagAsTrue guards against a real
// production bug: some items_base rows carry a stray tinyint value like 2 in an
// allow_* column instead of a clean 0/1, which the MySQL driver refuses to scan
// directly into bool.
func TestAdapterListDefinitionsTreatsNonZeroAllowFlagAsTrue(t *testing.T) {
	querier := &fakeQuerier{rows: &fakeRows{data: [][]any{
		{"weird_item", "s", "Weird Item", 1, 1, 2, 0, 0},
	}}}
	adapter := newAdapter(querier)

	definitions, err := adapter.ListDefinitions(context.Background())
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(definitions) != 1 || !definitions[0].AllowWalk {
		t.Fatalf("definitions = %#v, want AllowWalk = true for a stray value of 2", definitions)
	}
}

func TestAdapterListDefinitionsPropagatesQueryError(t *testing.T) {
	adapter := newAdapter(&fakeQuerier{queryErr: errors.New("query failed")})
	if _, err := adapter.ListDefinitions(context.Background()); err == nil {
		t.Fatal("ListDefinitions() error = nil")
	}
}

func TestAdapterInsertDefinitionsNoopOnEmpty(t *testing.T) {
	querier := &fakeQuerier{}
	adapter := newAdapter(querier)
	if err := adapter.InsertDefinitions(context.Background(), nil); err != nil {
		t.Fatalf("InsertDefinitions() error = %v", err)
	}
	if querier.execQuery != "" {
		t.Fatal("InsertDefinitions() issued a query for an empty batch")
	}
}

func TestAdapterInsertDefinitionsPropagatesExecError(t *testing.T) {
	adapter := newAdapter(&fakeQuerier{execErr: errors.New("exec failed")})
	err := adapter.InsertDefinitions(context.Background(), []syncrealm.Definition{{Classname: "a"}})
	if err == nil {
		t.Fatal("InsertDefinitions() error = nil")
	}
}

func TestBuildInsertQuery(t *testing.T) {
	definitions := []syncrealm.Definition{
		{Classname: "a", PublicName: "A", Kind: syncrealm.KindFloor, Width: 1, Length: 1},
		{Classname: "b", PublicName: "B", Kind: syncrealm.KindWall, Width: 2, Length: 2},
	}
	query, args := buildInsertQuery(definitions)

	if got := strings.Count(query, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"); got != len(definitions) {
		t.Fatalf("placeholder groups = %d, want %d; query = %q", got, len(definitions), query)
	}
	if len(args) != len(definitions)*insertColumnsPerRow {
		t.Fatalf("args = %d, want %d", len(args), len(definitions)*insertColumnsPerRow)
	}
	if args[2] != typeFloor {
		t.Fatalf("first row type = %v, want %q", args[2], typeFloor)
	}
	if args[2+insertColumnsPerRow] != typeWall {
		t.Fatalf("second row type = %v, want %q", args[2+insertColumnsPerRow], typeWall)
	}
}

func TestAdapterUpdateDefinitionsNoopOnEmpty(t *testing.T) {
	querier := &fakeQuerier{}
	adapter := newAdapter(querier)
	if err := adapter.UpdateDefinitions(context.Background(), nil); err != nil {
		t.Fatalf("UpdateDefinitions() error = %v", err)
	}
	if querier.execQuery != "" {
		t.Fatal("UpdateDefinitions() issued a query for an empty batch")
	}
}

func TestAdapterUpdateDefinitionsPropagatesExecError(t *testing.T) {
	adapter := newAdapter(&fakeQuerier{execErr: errors.New("exec failed")})
	err := adapter.UpdateDefinitions(context.Background(), []syncrealm.Definition{{Classname: "a"}})
	if err == nil {
		t.Fatal("UpdateDefinitions() error = nil")
	}
}

func TestBuildUpdateQuery(t *testing.T) {
	definitions := []syncrealm.Definition{
		{Classname: "a", PublicName: "A-new"},
		{Classname: "b", PublicName: "B-new"},
	}
	query, args := buildUpdateQuery(definitions)

	if got := strings.Count(query, "WHEN ? THEN ?"); got != len(definitions) {
		t.Fatalf("CASE branches = %d, want %d; query = %q", got, len(definitions), query)
	}
	if !strings.Contains(query, "public_name = CASE item_name") || !strings.Contains(query, "WHERE item_name IN") {
		t.Fatalf("query = %q, missing expected clauses", query)
	}
	// 2 args per CASE branch (name, value) + 1 arg per IN placeholder.
	want := len(definitions)*2 + len(definitions)
	if len(args) != want {
		t.Fatalf("args = %d, want %d", len(args), want)
	}
	if args[0] != "a" || args[1] != "A-new" {
		t.Fatalf("first CASE args = %v, %v", args[0], args[1])
	}
}
