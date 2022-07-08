package remote

import (
	"context"

	"github.com/forbole/juno/v3/node"
)

var (
	_ node.Source = &Source{}
)

// Source implements the keeper.Source interface relying on a GRPC connection
type Source struct {
	Ctx context.Context
}

// NewSource returns a new Source instance
func NewSource() (*Source, error) {
	return &Source{
		Ctx: context.Background(),
	}, nil
}

// Type implements keeper.Type
func (k Source) Type() string {
	return node.RemoteKeeper
}
