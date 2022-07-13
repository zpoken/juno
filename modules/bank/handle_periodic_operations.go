package bank

import (
	"fmt"

	"github.com/forbole/juno/v3/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "bank").Msg("setting up periodic tasks")

	// Setup a cron job to run every 10 mins
	if _, err := scheduler.Every(10).Minutes().Do(func() {
		m.updateAcccountBalance()
	}); err != nil {
		return fmt.Errorf("error while updating account balances: %s", err)
	}

	return nil
}

// updateAcccountBalance gets the latest accounts balances
// and stores them inside the database
func (m *Module) updateAcccountBalance() error {
	height, err := m.db.GetLastBlockHeight()
	if err != nil {
		return fmt.Errorf("error while getting latest height: %s", err)
	}

	log.Debug().Str("module", "bank").Int64("height", height).
		Msg("updating accounts balance")

	address := "nomic133g3xwupa82xkef95dnfruvgus5m44mcrhtqu6"
	balance, err := m.source.AccountBalance(address)
	if err != nil {
		return fmt.Errorf("error while getting account balances: %s", err)
	}

	return m.db.SaveAccountBalance(types.NewAccountBalance(
		address,
		balance,
		height,
	))

}
