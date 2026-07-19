package pets

import "context"

// StandardBundleNames maps Nitro pet type IDs 0 through 35 to their asset names.
// Duplicate names are intentional where two protocol types share one bundle.
var StandardBundleNames = []string{
	"dog", "cat", "croco", "terrier", "bear", "pig", "lion", "rhino", "spider", "turtle",
	"chicken", "frog", "dragon", "monster", "monkey", "horse", "monsterplant", "bunnyeaster",
	"bunnyevil", "bunnydepressed", "bunnylove", "pigeongood", "pigeonevil", "demonmonkey",
	"bearbaby", "terrierbaby", "gnome", "gnome", "kittenbaby", "puppybaby", "pigletbaby",
	"haloompa", "fools", "pterosaur", "velociraptor", "cow",
}

type clientCatalog struct{}

// NewClientCatalog creates a catalog containing Nitro's standard pet assets.
func NewClientCatalog() Catalog { return &clientCatalog{} }

func (*clientCatalog) ListNames(context.Context) ([]string, error) {
	return append([]string(nil), StandardBundleNames...), nil
}
