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
			{"id": 1, "classname": "throne_gold", "revision": 10, "unrelated": {"nested": true},
			 "name": "Gold Throne", "description": "A shiny throne", "xdim": 2, "ydim": 2,
			 "canstandon": false, "cansiton": true, "canlayon": false, "customparams": "gold"},
			{"id": 2, "classname": "chair_basic"},
			{"id": 3, "classname": ""}
		]
	},
	"wallitemtypes": {
		"furnitype": [
			{"id": 4, "classname": "poster500", "name": "poster500_name", "description": "poster500_desc"}
		]
	}
}`

func TestCatalogListClassnamesExtractsOnlyClassnames(t *testing.T) {
	catalog := newCatalog(&fakeReader{body: sampleFurnitureData})

	classnames, err := catalog.ListClassnames(context.Background())
	if err != nil {
		t.Fatalf("ListClassnames() error = %v", err)
	}
	if len(classnames) != 3 || classnames[0] != "throne_gold" || classnames[1] != "chair_basic" || classnames[2] != "poster500" {
		t.Fatalf("classnames = %#v", classnames)
	}
}

func TestCatalogListDefinitionsTagsKindByArray(t *testing.T) {
	catalog := newCatalog(&fakeReader{body: sampleFurnitureData})

	definitions, err := catalog.ListDefinitions(context.Background())
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(definitions) != 3 {
		t.Fatalf("definitions = %#v", definitions)
	}
	throne := definitions[0]
	if throne.Classname != "throne_gold" || throne.Kind != KindFloor || throne.Name != "Gold Throne" ||
		throne.Description != "A shiny throne" || throne.Width != 2 || throne.Length != 2 || !throne.CanSitOn {
		t.Fatalf("throne definition = %#v", throne)
	}
	poster := definitions[2]
	if poster.Classname != "poster500" || poster.Kind != KindWall {
		t.Fatalf("poster definition = %#v", poster)
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
