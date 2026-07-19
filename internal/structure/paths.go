package structure

// placeholderSuffix names the empty marker object created under a missing path
// so object storage consoles render it as a folder.
const placeholderSuffix = ".keep"

// ExpectedPaths are the canonical bucket key prefixes asset-cli expects to exist.
// See docs/wiki/STRUCTURE.md for what each path holds and why it is shaped this way.
var ExpectedPaths = []string{
	"avatar/clothing/",
	"avatar/effects/",
	"furniture/bundles/",
	"furniture/icons/",
	"pets/",
	"engine/",
	"media/badges/",
	"media/catalog-pages/",
	"media/campaigns/",
	"media/guilds/",
	"media/quests/",
	"media/talent/",
	"media/stories/",
	"media/flags/",
	"media/client-ui/",
	"sounds/ui/",
	"sounds/machine-samples/",
	"branding/",
	"gamedata/",
}
