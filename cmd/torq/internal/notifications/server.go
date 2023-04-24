package notifications

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/communications"
	"github.com/lncapital/torq/internal/services_helpers"
)

func StartNotifier(ctx context.Context, db *sqlx.DB) {

	serviceType := services_helpers.NotifierService

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

	serviceType := services_helpers.SlackService

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

	serviceType := services_helpers.TelegramHighService
	if !highPriority {
		serviceType = services_helpers.TelegramLowService
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
