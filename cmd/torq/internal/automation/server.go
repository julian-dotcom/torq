package automation

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/pkg/broadcast"
)

func Start(ctx context.Context, db *sqlx.DB, broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) error {
	var wg sync.WaitGroup

	// Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in TriggerMonitor %v", panicError)
				automation.TriggerMonitor(ctx, db, broadcaster, eventChannel)
			}
		}()
		automation.TriggerMonitor(ctx, db, broadcaster, eventChannel)
	})()

	wg.Wait()

	return nil
}
