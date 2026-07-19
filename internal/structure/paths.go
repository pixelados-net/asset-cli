package structure

// placeholderSuffix names the empty marker object created under a missing path
// so object storage consoles render it as a folder.
const placeholderSuffix = ".keep"

// ExpectedPaths are the canonical bucket keys asset-cli expects to exist. Entries
// ending in "/" are folder prefixes: Create fabricates an empty placeholder object
// for any of these that are missing. Entries without a trailing "/" are exact file
// keys — real content that must be uploaded, so Create reports them as missing
// but never fabricates a placeholder in their place.
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
	"gamedata/FurnitureData.json",
	"gamedata/ProductData.json",
	"gamedata/FigureData.json",
	"gamedata/FigureMap.json",
	"gamedata/ExternalTexts.json",
	"gamedata/HabboAvatarActions.json",
	"gamedata/UITexts.json",
	"gamedata/EffectMap.json",
}

// FlatPaths are expected paths that must hold one bundle or icon file per item
// directly under the prefix, with no nested subfolders.
var FlatPaths = []string{
	"avatar/clothing/",
	"avatar/effects/",
	"furniture/bundles/",
	"furniture/icons/",
	"engine/",
	"pets/",
}
