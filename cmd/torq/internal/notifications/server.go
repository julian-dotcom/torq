package notifications

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/communications"
	"github.com/lncapital/torq/pkg/cache"
	"github.com/lncapital/torq/pkg/core"
)

func StartNotifier(ctx context.Context, db *sqlx.DB) {

	serviceType := core.NotifierService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	communications.Notify(ctx, db)
}

func StartSlackListener(ctx context.Context, db *sqlx.DB) {

	serviceType := core.SlackService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	communications.SubscribeSlack(ctx, db)
}

func StartTelegramListeners(ctx context.Context, db *sqlx.DB, highPriority bool) {

	serviceType := core.TelegramHighService
	if !highPriority {
		serviceType = core.TelegramLowService
	}

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	communications.SubscribeTelegram(ctx, db, highPriority)
}
