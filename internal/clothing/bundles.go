package clothing

import (
	"context"
	"strings"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

// bundlesPrefix is the bucket prefix holding avatar clothing bundles.
const bundlesPrefix = "avatar/clothing/"

// bundleExtension is the extension used by clothing bundles.
const bundleExtension = ".nitro"

type bundleStorage struct{ client *minio.Client }

func (storage *bundleStorage) ListNames(ctx context.Context) ([]string, error) {
	keys, err := storage.client.ListKeys(ctx, bundlesPrefix)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		if !strings.HasSuffix(key, bundleExtension) {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(key, bundlesPrefix), bundleExtension)
		if name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}
