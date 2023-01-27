package services

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"
)

func Start(ctx context.Context, db *sqlx.DB,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest,
	broadcaster broadcast.BroadcastServer) error {

	var wg sync.WaitGroup

	active := commons.Active

	// Time Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in TimeTriggerMonitor %v", panicError)
				commons.RunningServices[commons.AutomationService].Cancel(commons.TorqDummyNodeId, &active, true)
			}
		}()
		automation.TimeTriggerMonitor(ctx, db)
	})()

	// Event Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in EventTriggerMonitor %v", panicError)
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
				log.Error().Msgf("Panic occurred in ScheduledTriggerMonitor %v", panicError)
				commons.RunningServices[commons.AutomationService].Cancel(commons.TorqDummyNodeId, &active, true)
			}
		}()
		automation.ScheduledTriggerMonitor(ctx, db, lightningRequestChannel, rebalanceRequestChannel)
	})()

	wg.Wait()

	return nil
}

func StartLightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer) error {

	var wg sync.WaitGroup

	active := commons.Active

	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in LightningCommunicationService %v", panicError)
				commons.RunningServices[commons.LightningCommunicationService].Cancel(nodeId, &active, true)
			}
		}()
		lnd.LightningCommunicationService(ctx, conn, db, nodeId, broadcaster)
	})()

	wg.Wait()

	return nil
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer) error {

	var wg sync.WaitGroup

	active := commons.Active

	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in RebalanceServiceStart %v", panicError)
				commons.RunningServices[commons.RebalanceService].Cancel(nodeId, &active, true)
			}
		}()
		automation.RebalanceServiceStart(ctx, conn, db, nodeId, broadcaster)
	})()

	wg.Wait()

	return nil
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB, vectorUrl string, nodeId int, lightningRequestChannel chan interface{}) error {
	var wg sync.WaitGroup

	active := commons.Active

	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in MaintenanceService %v", panicError)
				commons.RunningServices[commons.MaintenanceService].Cancel(nodeId, &active, true)
			}
		}()
		commons.MaintenanceServiceStart(ctx, db, vectorUrl, nodeId, lightningRequestChannel)
	})()

	wg.Wait()

	return nil
}
