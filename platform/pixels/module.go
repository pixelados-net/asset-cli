package pixels

import (
	"context"

	"go.uber.org/fx"
)

// Module provides the configured Pixels client and lifecycle cleanup.
var Module = fx.Module("pixels", fx.Provide(provideClient))

func provideClient(lifecycle fx.Lifecycle, config Config) (*Client, error) {
	client, err := New(config)
	if err != nil {
		return nil, err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		client.Close()
		return nil
	}})
	return client, nil
}
