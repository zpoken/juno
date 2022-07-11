package bank

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/logging"
	"github.com/forbole/juno/v3/modules"
)

var (
	_ modules.Module      = &Module{}
	_ modules.BlockModule = &Module{}
)

// Module represents the bank module
type Module struct {
	cdc    codec.Marshaler
	db     database.Database
	logger logging.Logger
}

func NewModule(cdc codec.Marshaler, db database.Database, logger logging.Logger) *Module {
	return &Module{
		cdc:    cdc,
		db:     db,
		logger: logger,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "bank"
}
