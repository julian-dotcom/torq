package services

import (
	"context"
	"runtime/debug"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/commons"
)

func StartIntervalService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationIntervalTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetActiveTorqServiceState(serviceType)

	automation.IntervalTriggerMonitor(ctx, db)

	commons.SetInactiveTorqServiceState(serviceType)
}

func StartChannelBalanceEventService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationChannelBalanceEventTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetActiveTorqServiceState(serviceType)

	automation.ChannelBalanceEventTriggerMonitor(ctx, db)

	commons.SetInactiveTorqServiceState(serviceType)
}

func StartChannelEventService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationChannelEventTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetActiveTorqServiceState(serviceType)

	automation.ChannelEventTriggerMonitor(ctx, db)

	commons.SetInactiveTorqServiceState(serviceType)
}

func StartScheduledService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationScheduledTriggerService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetActiveTorqServiceState(serviceType)

	automation.ScheduledTriggerMonitor(ctx, db)

	commons.SetInactiveTorqServiceState(serviceType)
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.RebalanceService

	defer log.Info().Msgf("%v terminated for nodeId: %v", serviceType.String(), nodeId)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking (nodeId: %v) %v", serviceType.String(), nodeId, string(debug.Stack()))
			commons.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	}()

	commons.SetActiveLndServiceState(serviceType, nodeId)

	workflows.RebalanceServiceStart(ctx, conn, db, nodeId)

	commons.SetInactiveLndServiceState(serviceType, nodeId)
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.MaintenanceService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetActiveTorqServiceState(serviceType)

	commons.MaintenanceServiceStart(ctx, db)

	commons.SetInactiveTorqServiceState(serviceType)
}

func StartCronService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.CronService

	defer log.Info().Msgf("%v terminated", serviceType.String())

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("%v is panicking %v", serviceType.String(), string(debug.Stack()))
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
	}()

	commons.SetActiveTorqServiceState(serviceType)

	automation.CronTriggerMonitor(ctx, db)
}
