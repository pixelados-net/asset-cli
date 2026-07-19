package gamedata

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type fakeReader struct {
	body string
	err  error
}

func (reader *fakeReader) Get(context.Context, string) (io.ReadCloser, error) {
	if reader.err != nil {
		return nil, reader.err
	}
	return io.NopCloser(strings.NewReader(reader.body)), nil
}

const sampleFurnitureData = `{
	"roomitemtypes": {
		"furnitype": [
			{"id": 1, "classname": "throne_gold", "revision": 10, "unrelated": {"nested": true}},
			{"id": 2, "classname": "chair_basic"},
			{"id": 3, "classname": ""}
		]
	}
}`

func TestCatalogListClassnamesExtractsOnlyClassnames(t *testing.T) {
	catalog := newCatalog(&fakeReader{body: sampleFurnitureData})

	classnames, err := catalog.ListClassnames(context.Background())
	if err != nil {
		t.Fatalf("ListClassnames() error = %v", err)
	}
	if len(classnames) != 2 || classnames[0] != "throne_gold" || classnames[1] != "chair_basic" {
		t.Fatalf("classnames = %#v", classnames)
	}
}

func TestCatalogListClassnamesPropagatesGetError(t *testing.T) {
	catalog := newCatalog(&fakeReader{err: errors.New("get failed")})

	if _, err := catalog.ListClassnames(context.Background()); err == nil {
		t.Fatal("ListClassnames() error = nil")
	}
}

func TestCatalogListClassnamesRejectsInvalidJSON(t *testing.T) {
	catalog := newCatalog(&fakeReader{body: "not json"})

	if _, err := catalog.ListClassnames(context.Background()); err == nil {
		t.Fatal("ListClassnames() error = nil")
	}
}
