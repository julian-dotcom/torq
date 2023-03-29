package services

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/commons"
)

func Start(ctx context.Context, db *sqlx.DB) {

	var wg sync.WaitGroup

	active := commons.ServiceActive

	// Interval Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in IntervalTriggerMonitor %v with stack: %v", panicError, string(debug.Stack()))
				commons.RunningServices[commons.AutomationService].Cancel(commons.TorqDummyNodeId, &active, true)
			}
		}()
		automation.IntervalTriggerMonitor(ctx, db)
	})()

	// Event Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in EventTriggerMonitor %v with stack: %v", panicError, string(debug.Stack()))
				commons.RunningServices[commons.AutomationService].Cancel(commons.TorqDummyNodeId, &active, true)
			}
		}()
		automation.EventTriggerMonitor(ctx, db)
	})()

	// Scheduled Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in ScheduledTriggerMonitor %v with stack: %v", panicError, string(debug.Stack()))
				commons.RunningServices[commons.AutomationService].Cancel(commons.TorqDummyNodeId, &active, true)
			}
		}()
		automation.ScheduledTriggerMonitor(ctx, db)
	})()

	wg.Wait()
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	active := commons.ServiceActive

	defer func() {
		if panicError := recover(); panicError != nil {
			log.Error().Msgf("Panic occurred in RebalanceServiceStart %v with stack: %v", panicError, string(debug.Stack()))
			commons.RunningServices[commons.RebalanceService].Cancel(nodeId, &active, true)
		}
	}()

	workflows.RebalanceServiceStart(ctx, conn, db, nodeId)
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB) {

	active := commons.ServiceActive

	defer func() {
		if panicError := recover(); panicError != nil {
			log.Error().Msgf("Panic occurred in MaintenanceService %v with stack: %v", panicError, string(debug.Stack()))
			commons.RunningServices[commons.MaintenanceService].Cancel(commons.TorqDummyNodeId, &active, true)
		}
	}()

	commons.MaintenanceServiceStart(ctx, db)
}

func StartCronService(ctx context.Context, db *sqlx.DB) {

	active := commons.ServiceActive

	// Cron Trigger Monitor
	defer func() {
		if panicError := recover(); panicError != nil {
			log.Error().Msgf("Panic occurred in CronTriggerMonitor %v with stack: %v", panicError, string(debug.Stack()))
			commons.RunningServices[commons.CronService].Cancel(commons.TorqDummyNodeId, &active, true)
		}
	}()

	automation.CronTriggerMonitor(ctx, db)
}
