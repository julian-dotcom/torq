package services

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/cache"
	"github.com/lncapital/torq/pkg/core"
)

func StartIntervalService(ctx context.Context, db *sqlx.DB) {

	serviceType := core.AutomationIntervalTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	automation.IntervalTriggerMonitor(ctx, db)

	cache.SetInactiveCoreServiceState(serviceType)
}

func StartChannelBalanceEventService(ctx context.Context, db *sqlx.DB) {

	serviceType := core.AutomationChannelBalanceEventTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	automation.ChannelBalanceEventTriggerMonitor(ctx, db)

	cache.SetInactiveCoreServiceState(serviceType)
}

func StartChannelEventService(ctx context.Context, db *sqlx.DB) {

	serviceType := core.AutomationChannelEventTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	automation.ChannelEventTriggerMonitor(ctx, db)

	cache.SetInactiveCoreServiceState(serviceType)
}

func StartScheduledService(ctx context.Context, db *sqlx.DB) {

	serviceType := core.AutomationScheduledTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	automation.ScheduledTriggerMonitor(ctx, db)

	cache.SetInactiveCoreServiceState(serviceType)
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := core.RebalanceService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	cache.SetActiveLndServiceState(serviceType, nodeId)

	workflows.RebalanceServiceStart(ctx, conn, db, nodeId)

	cache.SetInactiveLndServiceState(serviceType, nodeId)
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB) {

	serviceType := core.MaintenanceService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	automation.MaintenanceServiceStart(ctx, db)

	cache.SetInactiveCoreServiceState(serviceType)
}

func StartCronService(ctx context.Context, db *sqlx.DB) {

	serviceType := core.CronService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}()

	cache.SetActiveCoreServiceState(serviceType)

	automation.CronTriggerMonitor(ctx, db)
}
