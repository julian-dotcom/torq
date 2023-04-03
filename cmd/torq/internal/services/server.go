package services

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/commons"
)

func StartIntervalService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationIntervalTriggerService

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in %v", serviceType.String())
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
		commons.SetInactiveTorqServiceState(serviceType)
	}()
	commons.SetActiveTorqServiceState(serviceType)

	automation.IntervalTriggerMonitor(ctx, db)
}

func StartChannelBalanceEventService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationChannelBalanceEventTriggerService

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in %v", serviceType.String())
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
		commons.SetInactiveTorqServiceState(serviceType)
	}()
	commons.SetActiveTorqServiceState(serviceType)

	automation.ChannelBalanceEventTriggerMonitor(ctx, db)
}

func StartChannelEventService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationChannelEventTriggerService

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in %v", serviceType.String())
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
		commons.SetInactiveTorqServiceState(serviceType)
	}()
	commons.SetActiveTorqServiceState(serviceType)

	automation.ChannelEventTriggerMonitor(ctx, db)
}

func StartScheduledService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.AutomationScheduledTriggerService

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in %v", serviceType.String())
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
		commons.SetInactiveTorqServiceState(serviceType)
	}()
	commons.SetActiveTorqServiceState(serviceType)

	automation.ScheduledTriggerMonitor(ctx, db)
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	serviceType := commons.MaintenanceService

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in %v (nodeId: %v)", serviceType.String(), nodeId)
			commons.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(serviceType, nodeId)
	}()
	commons.SetActiveLndServiceState(serviceType, nodeId)

	workflows.RebalanceServiceStart(ctx, conn, db, nodeId)
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.MaintenanceService

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in %v", serviceType.String())
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
		commons.SetInactiveTorqServiceState(serviceType)
	}()
	commons.SetActiveTorqServiceState(serviceType)

	commons.MaintenanceServiceStart(ctx, db)
}

func StartCronService(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.CronService

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in %v", serviceType.String())
			commons.SetFailedTorqServiceState(serviceType)
			return
		}
		commons.SetInactiveTorqServiceState(serviceType)
	}()
	commons.SetActiveTorqServiceState(serviceType)

	automation.CronTriggerMonitor(ctx, db)
}
