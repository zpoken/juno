package staking

import (
	types "github.com/forbole/juno/v3/types"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(
	block *tmctypes.ResultBlock, _ *tmctypes.ResultBlockResults, _ *tmctypes.ResultValidators,
) error {

	// Update the staking pool
	go m.updateStakingPool(block.Block.Height)

	return nil
}

// updateStakingPool reads the current staking pool and stores its value inside the database
func (m *Module) updateStakingPool(height int64) {
	log.Debug().Str("module", "staking").Int64("height", height).
		Msg("updating staking pool")

	stakingPool, err := m.source.StakingPool()
	if err != nil {
		log.Error().Str("module", "staking").Err(err).Int64("height", height).
			Msg("error while getting staking pool")
		return
	}

	err = m.db.SaveStakingPool(types.NewStakingPool(stakingPool.BondedTokens, stakingPool.NotBondedTokens, height))
	if err != nil {
		log.Error().Str("module", "staking").Err(err).Int64("height", height).
			Msg("error while saving staking pool")
		return
	}
}
