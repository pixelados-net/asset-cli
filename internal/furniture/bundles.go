package furniture

import (
	"context"
	"strings"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

// bundlesPrefix is the bucket prefix holding furniture .nitro bundles.
const bundlesPrefix = "furniture/bundles/"

// bundleExtension is the file extension every furniture bundle uses.
const bundleExtension = ".nitro"

type bundleStorage struct {
	client *minio.Client
}

func (storage *bundleStorage) ListClassnames(ctx context.Context) ([]string, error) {
	keys, err := storage.client.ListKeys(ctx, bundlesPrefix)
	if err != nil {
		return nil, err
	}
	classnames := make([]string, 0, len(keys))
	for _, key := range keys {
		name := strings.TrimSuffix(strings.TrimPrefix(key, bundlesPrefix), bundleExtension)
		if name != "" {
			classnames = append(classnames, name)
		}
	}
	return classnames, nil
}
