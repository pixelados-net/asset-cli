package arcturus

import (
	"context"

	"go.uber.org/fx"
)

// Module provides the configured Arcturus client and lifecycle cleanup.
var Module = fx.Module("arcturus", fx.Provide(provideClient))

func provideClient(lifecycle fx.Lifecycle, config Config) (*Client, error) {
	client, err := New(config)
	if err != nil {
		return nil, err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		return client.Close()
	}})
	return client, nil
}
