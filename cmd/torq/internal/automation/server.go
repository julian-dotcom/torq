package automation

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

func Start(ctx context.Context, db *sqlx.DB, nodeId int, broadcaster broadcast.BroadcastServer) error {
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)

	var wg sync.WaitGroup

	active := commons.Active

	// Time Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in TimeTriggerMonitor %v", panicError)
				commons.RunningServices[commons.AutomationService].Cancel(nodeId, &active, true)
			}
		}()
		automation.TimeTriggerMonitor(ctx, db, nodeSettings)
	})()

	// Event Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in EventTriggerMonitor %v", panicError)
				commons.RunningServices[commons.AutomationService].Cancel(nodeId, &active, true)
			}
		}()
		automation.EventTriggerMonitor(ctx, db, nodeSettings, broadcaster)
	})()

	wg.Wait()

	return nil
}

func StartLightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	lightningCommunicationChannel chan interface{}) error {

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
		lnd.LightningCommunicationService(ctx, conn, db, nodeId, lightningCommunicationChannel)
	})()

	wg.Wait()

	return nil
}

func StartRebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	rebalanceRequestChannel chan commons.RebalanceRequest) error {

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
		commons.RebalanceServiceStart(ctx, conn, db, nodeId, rebalanceRequestChannel)
	})()

	wg.Wait()

	return nil
}

func StartMaintenanceService(ctx context.Context, db *sqlx.DB, nodeId int, lightningCommunicationChannel chan interface{}) error {
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
		commons.MaintenanceServiceStart(ctx, db, nodeId, lightningCommunicationChannel)
	})()

	wg.Wait()

	return nil
}
