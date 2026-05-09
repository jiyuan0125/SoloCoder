package plugin

import "context"

const CurrentAPIVersion = "1.0.0"

type Extension interface {
	Name() string
	Execute(ctx context.Context, input []byte) ([]byte, error)
}

type PluginMetadata struct {
	Name         string
	APIVersion   string
	Capabilities []string
}
