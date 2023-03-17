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
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"
)

func Start(ctx context.Context, db *sqlx.DB,
	lightningRequestChannel chan<- interface{},
	rebalanceRequestChannel chan<- commons.RebalanceRequests,
	broadcaster broadcast.BroadcastServer,
	serviceEventChannel chan<- commons.ServiceEvent) {

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
		automation.EventTriggerMonitor(ctx, db, broadcaster)
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
		automation.ScheduledTriggerMonitor(ctx, db, lightningRequestChannel, rebalanceRequestChannel)
	})()

	commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannel, commons.ServicePending, commons.ServiceActive, commons.AutomationService, nil)

	wg.Wait()
}

func StartLightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer, serviceEventChannel chan<- commons.ServiceEvent) {

	active := commons.ServiceActive

	defer func() {
		if panicError := recover(); panicError != nil {
			log.Error().Msgf("Panic occurred in LightningCommunicationService %v with stack: %v", panicError, string(debug.Stack()))
			commons.RunningServices[commons.LightningCommunicationService].Cancel(nodeId, &active, true)
		}
	}()

	lnd.LightningCommunicationService(ctx, conn, db, nodeId, broadcaster, serviceEventChannel)
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer, serviceEventChannel chan<- commons.ServiceEvent) {

	active := commons.ServiceActive

	defer func() {
		if panicError := recover(); panicError != nil {
			log.Error().Msgf("Panic occurred in RebalanceServiceStart %v with stack: %v", panicError, string(debug.Stack()))
			commons.RunningServices[commons.RebalanceService].Cancel(nodeId, &active, true)
		}
	}()

	workflows.RebalanceServiceStart(ctx, conn, db, nodeId, broadcaster, serviceEventChannel)
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB, serviceEventChannel chan<- commons.ServiceEvent) {

	active := commons.ServiceActive

	defer func() {
		if panicError := recover(); panicError != nil {
			log.Error().Msgf("Panic occurred in MaintenanceService %v with stack: %v", panicError, string(debug.Stack()))
			commons.RunningServices[commons.MaintenanceService].Cancel(commons.TorqDummyNodeId, &active, true)
		}
	}()

	commons.MaintenanceServiceStart(ctx, db)

	commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannel, commons.ServicePending, commons.ServiceActive, commons.MaintenanceService, nil)
}

func StartCronService(ctx context.Context, db *sqlx.DB, serviceEventChannel chan<- commons.ServiceEvent) {

	active := commons.ServiceActive

	defer func() {
		if panicError := recover(); panicError != nil {
			log.Error().Msgf("Panic occurred in CronTriggerMonitor %v with stack: %v", panicError, string(debug.Stack()))
			commons.RunningServices[commons.CronService].Cancel(commons.TorqDummyNodeId, &active, true)
		}
	}()

	automation.CronTriggerMonitor(ctx, db)

	commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannel, commons.ServicePending, commons.ServiceActive, commons.CronService, nil)
}
