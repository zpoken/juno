package pricefeed

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// RunAdditionalOperations implements modules.AdditionalOperationsModule
func (m *Module) RunAdditionalOperations() error {
	err := m.checkConfig()
	if err != nil {
		return err
	}

	return m.storeTokens()
}

// checkConfig checks if the module config is valid
func (m *Module) checkConfig() error {
	if m.cfg == nil {
		return fmt.Errorf("pricefeed config is not set but module is enabled")
	}

	return nil
}

// storeTokens stores the tokens defined inside the given configuration into the database
func (m *Module) storeTokens() error {
	log.Debug().Str("module", "pricefeed").Msg("storing tokens")

	for _, coin := range m.cfg.Tokens {
		// Save the coin as a token with its units
		err := m.db.SaveToken(coin)
		if err != nil {
			return fmt.Errorf("error while saving token: %s", err)
		}
	}

	return nil
}
