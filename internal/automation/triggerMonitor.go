package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

const workflowTickerSeconds = 10

func IntervalTriggerMonitor(ctx context.Context, db *sqlx.DB) {

	defer log.Info().Msgf("IntervalTriggerMonitor terminated")

	ticker := clock.New().Tick(workflowTickerSeconds * time.Second)
	bootstrapping := true

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			if bootstrapping {
				var allChannelIds []int
				torqNodeIds := commons.GetAllTorqNodeIds()
				for _, torqNodeId := range torqNodeIds {
					// Force Response because we don't care about balance accuracy
					channelIdsByNode := commons.GetChannelStateNotSharedChannelIds(torqNodeId, true)
					allChannelIds = append(allChannelIds, channelIdsByNode...)
				}
				if len(allChannelIds) == 0 {
					continue
				}
			}
			var bootedWorkflowVersionIds []int
			workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, commons.WorkflowNodeIntervalTrigger)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain root nodes (interval trigger nodes)")
				continue
			}
			for _, workflowTriggerNode := range workflowTriggerNodes {
				reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
				triggerSettings := commons.GetTimeTriggerSettingsByWorkflowVersionId(workflowTriggerNode.WorkflowVersionId)
				var param workflows.IntervalTriggerParameters
				err := json.Unmarshal([]byte(workflowTriggerNode.Parameters), &param)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to parse parameters for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					continue
				}
				if triggerSettings.BootTime != nil && int32(time.Since(*triggerSettings.BootTime).Seconds()) < param.Seconds {
					continue
				}
				if bootstrapping {
					bootedWorkflowVersionIds = append(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId)
				}
				commons.ScheduleTrigger(reference, workflowTriggerNode.WorkflowVersionId,
					workflowTriggerNode.Type, workflowTriggerNode.WorkflowVersionNodeId, workflowTriggerNode)
			}
			// When bootstrapping the automations run ALL workflows because we might have missed some events.
			if bootstrapping {
				workflowTriggerNodes, err = workflows.GetActiveEventTriggerNodes(db, commons.WorkflowNodeChannelBalanceEventTrigger)
				if err != nil {
					log.Error().Err(err).Msg("Failed to obtain root nodes (channel balance trigger nodes)")
					continue
				}
			triggerLoop:
				for _, workflowTriggerNode := range workflowTriggerNodes {
					for _, bootedWorkflowVersionId := range bootedWorkflowVersionIds {
						if bootedWorkflowVersionId == workflowTriggerNode.WorkflowVersionId {
							continue triggerLoop
						}
					}
					reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
					commons.ScheduleTrigger(reference,
						workflowTriggerNode.WorkflowVersionId,
						workflowTriggerNode.Type, workflowTriggerNode.WorkflowVersionNodeId, workflowTriggerNode)
				}
				bootstrapping = false
			}
		}
	}
}

type CronTriggerParams struct {
	CronValue string `json:"cronValue"`
}

func CronTriggerMonitor(ctx context.Context, db *sqlx.DB) {
	defer log.Info().Msgf("Cron trigger monitor terminated")

	ticker := clock.New().Tick(workflowTickerSeconds * time.Second)

bootstrappingLoop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			torqNodeIds := commons.GetAllTorqNodeIds()
			for _, torqNodeId := range torqNodeIds {
				// Force Response because we don't care about balance accuracy
				if len(commons.GetChannelStateNotSharedChannelIds(torqNodeId, true)) != 0 {
					break bootstrappingLoop
				}
			}
		}
	}

	workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, commons.WorkflowNodeCronTrigger)
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain root nodes (cron trigger nodes)")
		log.Error().Msg("Cron trigger monitor failed to start")
		return
	}

	var crons []*cron.Cron
	for _, trigger := range workflowTriggerNodes {
		var params CronTriggerParams
		if err = json.Unmarshal([]byte(trigger.Parameters), &params); err != nil {
			log.Error().Msgf("Can't unmarshal parameters for workflow version node id: %v", trigger.WorkflowVersionNodeId)
			continue
		}
		log.Debug().Msgf("Scheduling cron (%v) for workflow version node id: %v", params.CronValue, trigger.WorkflowVersionNodeId)
		c := cron.New()
		workflowVersionNodeId := trigger.WorkflowVersionNodeId
		workflowVersionId := trigger.WorkflowVersionId
		triggeringEvent := trigger
		_, err = c.AddFunc(params.CronValue, func() {
			log.Debug().Msgf("Scheduling for immediate execution cron trigger for workflow version node id %v", workflowVersionNodeId)
			reference := fmt.Sprintf("%v_%v", workflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
			commons.ScheduleTrigger(reference, workflowVersionId, commons.WorkflowNodeCronTrigger, workflowVersionNodeId, triggeringEvent)
		})
		if err != nil {
			log.Error().Msgf("Unable to add cron func for workflow version node id: %v", trigger.WorkflowVersionNodeId)
			continue
		}
		c.Start()
		crons = append(crons, c)
	}

	defer func() {
		for _, c := range crons {
			c.Stop()
		}
	}()

	log.Info().Msgf("Cron trigger monitor started")

	<-ctx.Done()
}

func EventTriggerMonitor(ctx context.Context, db *sqlx.DB,
	broadcaster broadcast.BroadcastServer) {

	defer log.Info().Msgf("EventTriggerMonitor terminated")

	var wg sync.WaitGroup

	wg.Add(1)
	go (func() {
		defer wg.Done()
		channelBalanceEventTriggerMonitor(ctx, db, broadcaster)
	})()

	wg.Add(1)
	go (func() {
		defer wg.Done()
		channelEventTriggerMonitor(ctx, db, broadcaster)
	})()

	wg.Wait()
}

func ScheduledTriggerMonitor(ctx context.Context, db *sqlx.DB,
	lightningRequestChannel chan<- interface{},
	rebalanceRequestChannel chan<- commons.RebalanceRequests) {

	defer log.Info().Msgf("ScheduledTriggerMonitor terminated")

	var delay bool

	for {
		if delay {
			ticker := clock.New().Tick(1 * time.Second)
			select {
			case <-ctx.Done():
				return
			case <-ticker:
			}
		} else {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		scheduledTrigger := commons.GetScheduledTrigger()
		if scheduledTrigger.SchedulingTime == nil {
			delay = true
			continue
		}
		events := scheduledTrigger.TriggeringEventQueue
		if len(events) > 0 {
			log.Debug().Msgf("ScheduledTriggerMonitor initiated for %v events", len(events))
			firstEvent := events[0]
			//lastEvent := events[len(events)-1]

			workflowTriggerNode, err := workflows.GetWorkflowNode(db, scheduledTrigger.TriggeringWorkflowVersionNodeId)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to obtain the Triggering WorkflowNode for WorkflowVersionNodeId: %v", scheduledTrigger.TriggeringWorkflowVersionNodeId)
				continue
			}
			triggerGroupWorkflowVersionNodeId, err := workflows.GetTriggerGroupWorkflowVersionNodeId(db, scheduledTrigger.TriggeringWorkflowVersionNodeId)
			if err != nil || triggerGroupWorkflowVersionNodeId == 0 {
				log.Error().Err(err).Msgf("Failed to obtain the group node id for WorkflowVersionNodeId: %v", scheduledTrigger.TriggeringWorkflowVersionNodeId)
				continue
			}
			groupWorkflowVersionNode, err := workflows.GetWorkflowNode(db, triggerGroupWorkflowVersionNodeId)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to obtain the group WorkflowNode for triggerGroupWorkflowVersionNodeId: %v", triggerGroupWorkflowVersionNodeId)
				continue
			}
			workflowTriggerNode.ChildNodes = groupWorkflowVersionNode.ChildNodes
			workflowTriggerNode.ParentNodes = make(map[int]*workflows.WorkflowNode)
			workflowTriggerNode.LinkDetails = groupWorkflowVersionNode.LinkDetails

			// Override the default WorkflowTrigger since you should never directly run the group node
			if scheduledTrigger.TriggeringNodeType == commons.WorkflowNodeManualTrigger && workflowTriggerNode.Type == commons.WorkflowTrigger {
				workflowTriggerNode.Type = commons.WorkflowNodeManualTrigger
			}

			//if scheduledTrigger.TriggeringNodeType == commons.WorkflowNodeChannelBalanceEventTrigger {
			//	// If the event is a WorkflowNode then this is the bootstrapping event that simulates a channel update
			//	// since we don't know what happened while Torq was offline.
			//	switch firstEvent.(type) {
			//	case workflows.WorkflowNode:
			//		eventTime := time.Now()
			//		dummyChannelBalanceEvents := make(map[int]map[int]*commons.ChannelBalanceEvent)
			//		torqNodeIds := commons.GetAllTorqNodeIds()
			//		for _, torqNodeId := range torqNodeIds {
			//			for _, channelId := range commons.GetChannelIdsByNodeId(torqNodeId) {
			//				channelSettings := commons.GetChannelSettingByChannelId(channelId)
			//				capacity := channelSettings.Capacity
			//				remoteNodeId := channelSettings.FirstNodeId
			//				if remoteNodeId == torqNodeId {
			//					remoteNodeId = channelSettings.SecondNodeId
			//				}
			//				channelState := commons.GetChannelState(torqNodeId, channelId, true)
			//				if channelState != nil {
			//					if dummyChannelBalanceEvents[remoteNodeId] == nil {
			//						dummyChannelBalanceEvents[remoteNodeId] = make(map[int]*commons.ChannelBalanceEvent)
			//					}
			//					dummyChannelBalanceEvents[remoteNodeId][channelId] = &commons.ChannelBalanceEvent{
			//						EventData: commons.EventData{
			//							EventTime: eventTime,
			//							NodeId:    torqNodeId,
			//						},
			//						ChannelId:            channelId,
			//						BalanceDelta:         0,
			//						BalanceDeltaAbsolute: 0,
			//						ChannelBalanceEventData: commons.ChannelBalanceEventData{
			//							Capacity:                  capacity,
			//							LocalBalance:              channelState.LocalBalance,
			//							RemoteBalance:             channelState.RemoteBalance,
			//							LocalBalancePerMilleRatio: int(channelState.LocalBalance / capacity * 1000),
			//						},
			//					}
			//				}
			//			}
			//		}
			//
			//		aggregateCapacity := make(map[int]int64)
			//		aggregateCount := make(map[int]int)
			//		aggregateLocalBalance := make(map[int]int64)
			//		aggregateLocalBalancePerMilleRatio := make(map[int]int)
			//		for remoteNodeId, dummyChannelBalanceEventByRemote := range dummyChannelBalanceEvents {
			//			var localBalanceAggregate int64
			//			var capacityAggregate int64
			//			for _, dummyChannelBalanceEvent := range dummyChannelBalanceEventByRemote {
			//				localBalanceAggregate += dummyChannelBalanceEvent.LocalBalance
			//				capacityAggregate += dummyChannelBalanceEvent.Capacity
			//			}
			//			if capacityAggregate == 0 {
			//				continue
			//			}
			//			aggregateCapacity[remoteNodeId] = capacityAggregate
			//			aggregateCount[remoteNodeId] = len(dummyChannelBalanceEventByRemote)
			//			aggregateLocalBalance[remoteNodeId] = localBalanceAggregate
			//			aggregateLocalBalancePerMilleRatio[remoteNodeId] = int(localBalanceAggregate / capacityAggregate * 1000)
			//		}
			//		for remoteNodeId, dummyChannelBalanceEventByRemote := range dummyChannelBalanceEvents {
			//			for _, dummyChannelBalanceEvent := range dummyChannelBalanceEventByRemote {
			//				dummyChannelBalanceEvent.PeerChannelCapacity = aggregateCapacity[remoteNodeId]
			//				dummyChannelBalanceEvent.PeerChannelCount = aggregateCount[remoteNodeId]
			//				dummyChannelBalanceEvent.PeerLocalBalance = aggregateLocalBalance[remoteNodeId]
			//				dummyChannelBalanceEvent.PeerLocalBalancePerMilleRatio = aggregateLocalBalancePerMilleRatio[remoteNodeId]
			//
			//				triggerCtx, triggerCancel := context.WithCancel(ctx)
			//
			//				commons.ActivateEventTrigger(scheduledTrigger.Reference,
			//					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId,
			//					scheduledTrigger.TriggeringNodeType, firstEvent, triggerCancel)
			//
			//				processWorkflowNode(triggerCtx, db, workflowTriggerNode,
			//					scheduledTrigger.Reference, events, lightningRequestChannel,
			//					rebalanceRequestChannel)
			//
			//				triggerCancel()
			//
			//				commons.DeactivateEventTrigger(workflowTriggerNode.WorkflowVersionId,
			//					workflowTriggerNode.WorkflowVersionNodeId, scheduledTrigger.TriggeringNodeType, lastEvent)
			//
			//			}
			//		}
			//		continue
			//	}
			//}

			triggerCtx, triggerCancel := context.WithCancel(ctx)

			switch workflowTriggerNode.Type {
			case commons.WorkflowNodeIntervalTrigger, commons.WorkflowNodeCronTrigger, commons.WorkflowNodeManualTrigger:
				commons.ActivateWorkflowTrigger(scheduledTrigger.Reference,
					workflowTriggerNode.WorkflowVersionId, triggerCancel)
			default:
				commons.ActivateEventTrigger(scheduledTrigger.Reference,
					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId,
					scheduledTrigger.TriggeringNodeType, firstEvent, triggerCancel)
			}

			processWorkflowNode(triggerCtx, db, workflowTriggerNode,
				scheduledTrigger.Reference, events, lightningRequestChannel,
				rebalanceRequestChannel)

			triggerCancel()

			switch workflowTriggerNode.Type {
			case commons.WorkflowNodeIntervalTrigger, commons.WorkflowNodeCronTrigger, commons.WorkflowNodeManualTrigger:
				commons.DeactivateWorkflowTrigger(workflowTriggerNode.WorkflowVersionId)
			default:
				commons.DeactivateEventTrigger(workflowTriggerNode.WorkflowVersionId,
					workflowTriggerNode.WorkflowVersionNodeId, scheduledTrigger.TriggeringNodeType, firstEvent)
			}
		}
	}
}

func channelBalanceEventTriggerMonitor(ctx context.Context, db *sqlx.DB,
	broadcaster broadcast.BroadcastServer) {

	defer log.Info().Msgf("EventTriggerMonitor terminated")

	listener := broadcaster.SubscribeChannelBalanceEvent()
	defer broadcaster.CancelSubscriptionChannelBalanceEvent(listener)

	for {
		select {
		case <-ctx.Done():
			return
		case channelBalanceEvent := <-listener:
			if channelBalanceEvent.NodeId == 0 || channelBalanceEvent.ChannelId == 0 {
				continue
			}
			processEventTrigger(db, channelBalanceEvent, commons.WorkflowNodeChannelBalanceEventTrigger)
		}
	}
}

func channelEventTriggerMonitor(ctx context.Context, db *sqlx.DB, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeChannelEvent()
	defer broadcaster.CancelSubscriptionChannelEvent(listener)

	for {
		select {
		case <-ctx.Done():
			return
		case channelEvent := <-listener:
			if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
				continue
			}
			switch channelEvent.Type {
			case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
				processEventTrigger(db, channelEvent, commons.WorkflowNodeChannelOpenEventTrigger)
			case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
				processEventTrigger(db, channelEvent, commons.WorkflowNodeChannelCloseEventTrigger)
			}
		}
	}
}

func processEventTrigger(db *sqlx.DB, triggeringEvent any, workflowNodeType commons.WorkflowNodeType) {

	//if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
	// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
	//}

	workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, workflowNodeType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain root nodes (trigger nodes)")
		return
	}
	for _, workflowTriggerNode := range workflowTriggerNodes {
		reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
		commons.ScheduleTrigger(reference, workflowTriggerNode.WorkflowVersionId, workflowNodeType, workflowTriggerNode.WorkflowVersionNodeId, triggeringEvent)
	}
}

func processWorkflowNode(ctx context.Context, db *sqlx.DB,
	workflowTriggerNode workflows.WorkflowNode,
	reference string,
	events []any,
	lightningRequestChannel chan<- interface{},
	rebalanceRequestChannel chan<- commons.RebalanceRequests) {

	err := workflows.ProcessWorkflow(ctx, db, workflowTriggerNode, reference, events, lightningRequestChannel, rebalanceRequestChannel)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to trigger nodes for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}
}
