package ibc

import (
	"fmt"

	"github.com/forbole/juno/v3/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "ibc").Msg("setting up periodic tasks")

	// Setup a cron job to run every midnight
	if _, err := scheduler.Every(1).Day().At("00:00").Do(func() {
		m.updateIBCParams()
	}); err != nil {
		return err
	}

	return nil
}

// updateIBCParams gets the updated ibc params
// and stores them inside the database
func (m *Module) updateIBCParams() error {
	height, err := m.db.GetLastBlockHeight()
	if err != nil {
		return err
	}

	log.Debug().Str("module", "ibc").Int64("height", height).
		Msg("updating ibc params")

	params, err := m.source.IBCParams()
	if err != nil {
		return fmt.Errorf("error while getting params: %s", err)
	}

	return m.db.SaveIBCParams(types.NewIBCParams(params, height))

}
