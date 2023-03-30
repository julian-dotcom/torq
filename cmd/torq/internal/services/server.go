package services

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/commons"
)

func Start(ctx context.Context, db *sqlx.DB) {

	commons.SetActiveTorqServiceState(commons.AutomationService)
	var wg sync.WaitGroup

	// Interval Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msg("Panic occurred in AutomationService: Interval Trigger Monitor")
				commons.SetFailedTorqServiceState(commons.AutomationService)
				return
			}
			commons.SetInactiveTorqServiceState(commons.AutomationService)
		}()

		automation.IntervalTriggerMonitor(ctx, db)
	})()

	// Event Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msg("Panic occurred in AutomationService: Event Trigger Monitor")
				commons.SetFailedTorqServiceState(commons.AutomationService)
				return
			}
			commons.SetInactiveTorqServiceState(commons.AutomationService)
		}()

		automation.EventTriggerMonitor(ctx, db)
	})()

	// Scheduled Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msg("Panic occurred in AutomationService: Scheduled Trigger Monitor")
				commons.SetFailedTorqServiceState(commons.AutomationService)
				return
			}
			commons.SetInactiveTorqServiceState(commons.AutomationService)
		}()

		automation.ScheduledTriggerMonitor(ctx, db)
	})()

	wg.Wait()
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Panic occurred in RebalanceService (nodeId: %v)", nodeId)
			commons.SetFailedLndServiceState(commons.RebalanceService, nodeId)
			return
		}
		commons.SetInactiveLndServiceState(commons.RebalanceService, nodeId)
	}()
	commons.SetActiveLndServiceState(commons.RebalanceService, nodeId)

	workflows.RebalanceServiceStart(ctx, conn, db, nodeId)
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msg("Panic occurred in MaintenanceService")
			commons.SetFailedTorqServiceState(commons.MaintenanceService)
			return
		}
		commons.SetInactiveTorqServiceState(commons.MaintenanceService)
	}()
	commons.SetActiveTorqServiceState(commons.MaintenanceService)

	commons.MaintenanceServiceStart(ctx, db)
}

func StartCronService(ctx context.Context, db *sqlx.DB) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msg("Panic occurred in CronService")
			commons.SetFailedTorqServiceState(commons.CronService)
			return
		}
		commons.SetInactiveTorqServiceState(commons.CronService)
	}()
	commons.SetActiveTorqServiceState(commons.CronService)

	automation.CronTriggerMonitor(ctx, db)
}
