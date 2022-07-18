package pricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/logging"
	"github.com/forbole/juno/v3/modules"
	source "github.com/forbole/juno/v3/node"
	"github.com/forbole/juno/v3/types/config"
)

var (
	_ modules.Module                     = &Module{}
	_ modules.AdditionalOperationsModule = &Module{}
)

// Module represents the pricefeed module
type Module struct {
	cfg    *Config
	cdc    codec.Marshaler
	db     database.Database
	logger logging.Logger
	source source.Node
}

func NewModule(cfg config.Config, cdc codec.Marshaler, db database.Database, logger logging.Logger, source source.Node) *Module {
	bz, err := cfg.GetBytes()
	if err != nil {
		panic(err)
	}

	pricefeedCfg, err := ParseConfig(bz)
	if err != nil {
		panic(err)
	}

	return &Module{
		cfg:    pricefeedCfg,
		cdc:    cdc,
		db:     db,
		logger: logger,
		source: source,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "pricefeed"
}
