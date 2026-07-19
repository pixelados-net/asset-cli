package pixels

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
}

func (fake *fakeRows) Next() bool {
	return fake.idx < len(fake.data)
}

func (fake *fakeRows) Scan(dest ...any) error {
	row := fake.data[fake.idx]
	fake.idx++
	*dest[0].(*string) = row[0].(string)
	*dest[1].(*string) = row[1].(string)
	*dest[2].(*string) = row[2].(string)
	*dest[3].(*string) = row[3].(string)
	*dest[4].(*int) = row[4].(int)
	*dest[5].(*int) = row[5].(int)
	*dest[6].(*bool) = row[6].(bool)
	*dest[7].(*bool) = row[7].(bool)
	*dest[8].(*bool) = row[8].(bool)
	return nil
}

func (fake *fakeRows) Err() error { return nil }
func (fake *fakeRows) Close()     {}

type fakePool struct {
	rows      *fakeRows
	queryErr  error
	execErr   error
	execQuery string
	execArgs  []any
}

func (fake *fakePool) Query(context.Context, string, ...any) (rows, error) {
	if fake.queryErr != nil {
		return nil, fake.queryErr
	}
	return fake.rows, nil
}

func (fake *fakePool) Exec(_ context.Context, query string, args ...any) error {
	fake.execQuery = query
	fake.execArgs = args
	return fake.execErr
}

func TestAdapterListDefinitions(t *testing.T) {
	pool := &fakePool{rows: &fakeRows{data: [][]any{
		{"chair_plasto", syncrealm.KindFloor, "Chair", "A plain chair", 1, 1, false, true, false},
	}}}
	adapter := newAdapter(pool)

	definitions, err := adapter.ListDefinitions(context.Background())
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(definitions) != 1 || definitions[0].Classname != "chair_plasto" || definitions[0].Description != "A plain chair" {
		t.Fatalf("definitions = %#v", definitions)
	}
}

func TestAdapterListDefinitionsPropagatesQueryError(t *testing.T) {
	adapter := newAdapter(&fakePool{queryErr: errors.New("query failed")})
	if _, err := adapter.ListDefinitions(context.Background()); err == nil {
		t.Fatal("ListDefinitions() error = nil")
	}
}

func TestAdapterInsertDefinitionsNoopOnEmpty(t *testing.T) {
	pool := &fakePool{}
	adapter := newAdapter(pool)
	if err := adapter.InsertDefinitions(context.Background(), nil); err != nil {
		t.Fatalf("InsertDefinitions() error = %v", err)
	}
	if pool.execQuery != "" {
		t.Fatal("InsertDefinitions() issued a query for an empty batch")
	}
}

func TestAdapterInsertDefinitionsPropagatesExecError(t *testing.T) {
	adapter := newAdapter(&fakePool{execErr: errors.New("exec failed")})
	err := adapter.InsertDefinitions(context.Background(), []syncrealm.Definition{{Classname: "a"}})
	if err == nil {
		t.Fatal("InsertDefinitions() error = nil")
	}
}

func TestAdapterInsertDefinitionsBuildsPositionalPlaceholders(t *testing.T) {
	pool := &fakePool{}
	adapter := newAdapter(pool)
	definitions := []syncrealm.Definition{
		{Classname: "a", Description: "Desc A"},
		{Classname: "b", Description: "Desc B"},
	}
	if err := adapter.InsertDefinitions(context.Background(), definitions); err != nil {
		t.Fatalf("InsertDefinitions() error = %v", err)
	}
	if !strings.Contains(pool.execQuery, "$1") || !strings.Contains(pool.execQuery, "$12") {
		t.Fatalf("query = %q, want $1..$%d placeholders", pool.execQuery, len(definitions)*insertColumnsPerRow)
	}
	if len(pool.execArgs) != len(definitions)*insertColumnsPerRow {
		t.Fatalf("args = %d, want %d", len(pool.execArgs), len(definitions)*insertColumnsPerRow)
	}
	if pool.execArgs[2] != "Desc A" {
		t.Fatalf("first row description = %v", pool.execArgs[2])
	}
}

func TestAdapterUpdateDefinitionsNoopOnEmpty(t *testing.T) {
	pool := &fakePool{}
	adapter := newAdapter(pool)
	if err := adapter.UpdateDefinitions(context.Background(), nil); err != nil {
		t.Fatalf("UpdateDefinitions() error = %v", err)
	}
	if pool.execQuery != "" {
		t.Fatal("UpdateDefinitions() issued a query for an empty batch")
	}
}

func TestAdapterUpdateDefinitionsPropagatesExecError(t *testing.T) {
	adapter := newAdapter(&fakePool{execErr: errors.New("exec failed")})
	err := adapter.UpdateDefinitions(context.Background(), []syncrealm.Definition{{Classname: "a"}})
	if err == nil {
		t.Fatal("UpdateDefinitions() error = nil")
	}
}

func TestBuildUpdateQuery(t *testing.T) {
	definitions := []syncrealm.Definition{
		{Classname: "a", PublicName: "A-new", Description: "Desc A"},
		{Classname: "b", PublicName: "B-new", Description: "Desc B"},
	}
	query, args := buildUpdateQuery(definitions)

	if !strings.Contains(query, "from (values") || !strings.Contains(query, "t.name = v.name") {
		t.Fatalf("query = %q, missing expected clauses", query)
	}
	if !strings.Contains(query, "$1") || !strings.Contains(query, "$6") {
		t.Fatalf("query = %q, want $1..$%d placeholders", query, len(definitions)*updateColumnsPerRow)
	}
	if len(args) != len(definitions)*updateColumnsPerRow {
		t.Fatalf("args = %d, want %d", len(args), len(definitions)*updateColumnsPerRow)
	}
	if args[0] != "a" || args[1] != "A-new" || args[2] != "Desc A" {
		t.Fatalf("first row args = %v", args[:3])
	}
}
