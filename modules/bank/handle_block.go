package bank

import (
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(
	block *tmctypes.ResultBlock, _ *tmctypes.ResultBlockResults, _ *tmctypes.ResultValidators,
) error {
	err := m.updateSupply(block.Block.Height)
	if err != nil {
		log.Error().Str("module", "bank").Int64("height", block.Block.Height).
			Err(err).Msg("error while updating supply")
	}

	return nil
}

// updateSupply updates the supply for a given height
func (m *Module) updateSupply(height int64) error {
	log.Debug().Str("module", "bank").Int64("height", height).
		Msg("updating supply")

	supply, err := m.source.Supply()
	if err != nil {
		return err
	}

	return m.db.SaveSupply(supply, height)
}
