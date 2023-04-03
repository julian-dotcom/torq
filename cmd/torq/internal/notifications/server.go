package notifications

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/communications"
	"github.com/lncapital/torq/pkg/commons"
)

func StartNotifier(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.NotifierService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetPendingTorqServiceState(serviceType)

	communications.Notify(ctx, db)
}

func StartSlackListener(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.SlackService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetPendingTorqServiceState(serviceType)

	communications.SubscribeSlack(ctx, db)
}

func StartTelegramListeners(ctx context.Context, db *sqlx.DB, highPriority bool) {
	serviceType := commons.TelegramHighService
	if !highPriority {
		serviceType = commons.TelegramLowService
	}

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetActiveTorqServiceState(serviceType)

	communications.SubscribeTelegram(ctx, db, highPriority)
}
