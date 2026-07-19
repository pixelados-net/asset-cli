package stats

// nitroExtension is the file extension counted by the nitro command.
const nitroExtension = ".nitro"

// NitroCategory names one bucket prefix whose .nitro bundles are counted.
type NitroCategory struct {
	// Name is the human-readable category label shown in output.
	Name string
	// Path is the bucket prefix holding this category's .nitro bundles.
	Path string
}

// NitroCategories are the content categories counted by the nitro command. Engine
// bundles and furniture icons are intentionally excluded: they are internal client
// assets, not purchasable content.
var NitroCategories = []NitroCategory{
	{Name: "clothing", Path: "avatar/clothing/"},
	{Name: "effects", Path: "avatar/effects/"},
	{Name: "furniture", Path: "furniture/bundles/"},
	{Name: "pets", Path: "pets/"},
}
