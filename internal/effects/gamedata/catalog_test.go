package gamedata

import (
	"context"
	"io"
	"strings"
	"testing"
)

type fakeReader struct{ body string }

func (fake fakeReader) Get(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(fake.body)), nil
}

func TestCatalogListNames(t *testing.T) {
	catalog := newCatalog(fakeReader{body: `{"effects":[{"id":"4","lib":"Twinkle"},{"id":"5","lib":"Torch"},{"id":"6","lib":""}]}`})
	names, err := catalog.ListNames(context.Background())
	if err != nil {
		t.Fatalf("ListNames() error = %v", err)
	}
	if len(names) != 2 || names[0] != "Twinkle" || names[1] != "Torch" {
		t.Fatalf("names = %#v", names)
	}
}

func TestCatalogListNamesRejectsInvalidJSON(t *testing.T) {
	catalog := newCatalog(fakeReader{body: `{`})
	if _, err := catalog.ListNames(context.Background()); err == nil {
		t.Fatal("ListNames() error = nil")
	}
}
