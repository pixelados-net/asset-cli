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
	catalog := newCatalog(fakeReader{body: `{"libraries":[{"id":"hh_human_body"},{"id":"acc_hat_U_cap"},{"id":""}]}`})
	names, err := catalog.ListNames(context.Background())
	if err != nil {
		t.Fatalf("ListNames() error = %v", err)
	}
	if len(names) != 2 || names[0] != "hh_human_body" || names[1] != "acc_hat_U_cap" {
		t.Fatalf("names = %#v", names)
	}
}

func TestCatalogListNamesRejectsInvalidJSON(t *testing.T) {
	catalog := newCatalog(fakeReader{body: `{`})
	if _, err := catalog.ListNames(context.Background()); err == nil {
		t.Fatal("ListNames() error = nil")
	}
}
