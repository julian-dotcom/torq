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

	// Time Trigger Monitor
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in TimeTriggerMonitor %v", panicError)
				automation.TimeTriggerMonitor(ctx, db, nodeSettings)
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
				automation.EventTriggerMonitor(ctx, db, nodeSettings, broadcaster)
			}
		}()
		automation.EventTriggerMonitor(ctx, db, nodeSettings, broadcaster)
	})()

	wg.Wait()

	return nil
}

func StartLightningCommunicationService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer) error {

	var wg sync.WaitGroup

	// Fee Service
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in TimeTriggerMonitor %v", panicError)
				lnd.LightningCommunicationService(ctx, conn, db, nodeId, broadcaster)
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

	// Rebalance Service
	wg.Add(1)
	go (func() {
		defer wg.Done()
		defer func() {
			if panicError := recover(); panicError != nil {
				log.Error().Msgf("Panic occurred in TimeTriggerMonitor %v", panicError)
				lnd.RebalanceService(ctx, conn, db, nodeId, broadcaster)
			}
		}()
		lnd.RebalanceService(ctx, conn, db, nodeId, broadcaster)
	})()

	wg.Wait()

	return nil
}
