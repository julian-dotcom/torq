package services

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"
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
		defer commons.SetInactiveTorqServiceState(commons.AutomationService)

		automation.IntervalTriggerMonitor(ctx, db)
	})()

	// Event Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer commons.SetInactiveTorqServiceState(commons.AutomationService)

		automation.EventTriggerMonitor(ctx, db)
	})()

	// Scheduled Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer commons.SetInactiveTorqServiceState(commons.AutomationService)

		automation.ScheduledTriggerMonitor(ctx, db)
	})()

	wg.Wait()
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int) {

	defer commons.SetInactiveLndServiceState(commons.RebalanceService, nodeId)
	commons.SetActiveLndServiceState(commons.RebalanceService, nodeId)

	workflows.RebalanceServiceStart(ctx, conn, db, nodeId)
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB) {

	defer commons.SetInactiveTorqServiceState(commons.MaintenanceService)
	commons.SetActiveTorqServiceState(commons.MaintenanceService)

	commons.MaintenanceServiceStart(ctx, db)
}

func StartCronService(ctx context.Context, db *sqlx.DB) {

	defer commons.SetInactiveTorqServiceState(commons.CronService)
	commons.SetActiveTorqServiceState(commons.CronService)

	automation.CronTriggerMonitor(ctx, db)
}
