package staking

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/logging"
	"github.com/forbole/juno/v3/modules"
	source "github.com/forbole/juno/v3/node"
)

var (
	_ modules.Module      = &Module{}
	_ modules.BlockModule = &Module{}
)

// Module represents the staking module
type Module struct {
	cdc    codec.Marshaler
	db     database.Database
	logger logging.Logger
	source source.Node
}

func NewModule(cdc codec.Marshaler, db database.Database, logger logging.Logger, source source.Node) *Module {
	return &Module{
		cdc:    cdc,
		db:     db,
		logger: logger,
		source: source,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "staking"
}
