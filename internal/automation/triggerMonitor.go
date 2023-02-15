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

func IntervalTriggerMonitor(ctx context.Context, db *sqlx.DB) {

	defer log.Info().Msgf("IntervalTriggerMonitor terminated")

	ticker := clock.New().Tick(commons.WORKFLOW_TICKER_SECONDS * time.Second)
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
					channelIdsByNode := commons.GetChannelStateChannelIds(torqNodeId, true)
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
				err := json.Unmarshal([]byte(workflowTriggerNode.Parameters.([]uint8)), &param)
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

	ticker := clock.New().Tick(commons.WORKFLOW_TICKER_SECONDS * time.Second)

bootstrappingLoop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			var allChannelIds []int
			torqNodeIds := commons.GetAllTorqNodeIds()
			for _, torqNodeId := range torqNodeIds {
				// Force Response because we don't care about balance accuracy
				channelIdsByNode := commons.GetChannelStateChannelIds(torqNodeId, true)
				allChannelIds = append(allChannelIds, channelIdsByNode...)
			}
			if len(allChannelIds) != 0 {
				break bootstrappingLoop
			}
		}
	}

	workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, commons.WorkflowNodeCronTrigger)
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain root nodes (cron trigger nodes)")
		log.Error().Msg("Cron trigger monitor failed to start")
		return
	}

	c := cron.New()

	for _, trigger := range workflowTriggerNodes {
		var params CronTriggerParams
		if err = json.Unmarshal(trigger.Parameters.([]byte), &params); err != nil {
			log.Error().Msgf("Can't unmarshal parameters for workflow version node id: %v", trigger.WorkflowVersionNodeId)
			continue
		}
		log.Debug().Msgf("Scheduling cron (%v) for workflow version node id: %v", params.CronValue, trigger.WorkflowVersionNodeId)
		_, err = c.AddFunc(params.CronValue, func() {
			log.Debug().Msgf("Scheduling for immediate execution cron trigger for workflow version node id %v", trigger.WorkflowVersionNodeId)
			reference := fmt.Sprintf("%v_%v", trigger.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
			commons.ScheduleTrigger(reference, trigger.WorkflowVersionId,
				commons.WorkflowNodeCronTrigger, trigger.WorkflowVersionNodeId, trigger)
		})
		if err != nil {
			log.Error().Msgf("Unable to add cron func for workflow version node id: %v", trigger.WorkflowVersionNodeId)
			continue
		}
	}

	c.Start()
	defer c.Stop()

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
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest) {

	defer log.Info().Msgf("ScheduledTriggerMonitor terminated")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		scheduledTrigger := commons.GetScheduledTrigger()
		if scheduledTrigger.SchedulingTime == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		events := scheduledTrigger.TriggeringEventQueue
		if len(events) > 0 {
			firstEvent := events[0]
			lastEvent := events[len(events)-1]

			var eventChannelIds []int
			for _, event := range events {
				channelBalanceEvent, ok := event.(commons.ChannelBalanceEvent)
				if ok {
					eventChannelIds = append(eventChannelIds, channelBalanceEvent.ChannelId)
				}
				channelEvent, ok := event.(commons.ChannelEvent)
				if ok {
					eventChannelIds = append(eventChannelIds, channelEvent.ChannelId)
				}
			}

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

			if scheduledTrigger.TriggeringNodeType == commons.WorkflowNodeChannelBalanceEventTrigger {
				// If the event is a WorkflowNode then this is the bootstrapping event that simulates a channel update
				// since we don't know what happened while Torq was offline.
				switch firstEvent.(type) {
				case workflows.WorkflowNode:
					eventTime := time.Now()
					dummyChannelBalanceEvents := make(map[int]map[int]*commons.ChannelBalanceEvent)
					torqNodeIds := commons.GetAllTorqNodeIds()
					for _, torqNodeId := range torqNodeIds {
						for _, channelId := range commons.GetChannelIdsByNodeId(torqNodeId) {
							channelSettings := commons.GetChannelSettingByChannelId(channelId)
							capacity := channelSettings.Capacity
							remoteNodeId := channelSettings.FirstNodeId
							if remoteNodeId == torqNodeId {
								remoteNodeId = channelSettings.SecondNodeId
							}
							channelState := commons.GetChannelState(torqNodeId, channelId, true)
							if channelState != nil {
								if dummyChannelBalanceEvents[remoteNodeId] == nil {
									dummyChannelBalanceEvents[remoteNodeId] = make(map[int]*commons.ChannelBalanceEvent)
								}
								dummyChannelBalanceEvents[remoteNodeId][channelId] = &commons.ChannelBalanceEvent{
									EventData: commons.EventData{
										EventTime: eventTime,
										NodeId:    torqNodeId,
									},
									ChannelId: channelId,
									ChannelBalanceEventData: commons.ChannelBalanceEventData{
										Capacity:                  capacity,
										LocalBalance:              channelState.LocalBalance,
										RemoteBalance:             channelState.RemoteBalance,
										LocalBalancePerMilleRatio: int(channelState.LocalBalance / capacity * 1000),
									},
								}
							}
						}
					}

					aggregateLocalBalance := make(map[int]int64)
					aggregateLocalBalancePerMilleRatio := make(map[int]int)
					for remoteNodeId, dummyChannelBalanceEventByRemote := range dummyChannelBalanceEvents {
						var localBalanceAggregate int64
						var capacityAggregate int64
						for _, dummyChannelBalanceEvent := range dummyChannelBalanceEventByRemote {
							localBalanceAggregate += dummyChannelBalanceEvent.LocalBalance
							capacityAggregate += dummyChannelBalanceEvent.Capacity
						}
						if capacityAggregate == 0 {
							continue
						}
						aggregateLocalBalance[remoteNodeId] = localBalanceAggregate
						aggregateLocalBalancePerMilleRatio[remoteNodeId] = int(localBalanceAggregate / capacityAggregate * 1000)
					}
					for remoteNodeId, dummyChannelBalanceEventByRemote := range dummyChannelBalanceEvents {
						for _, dummyChannelBalanceEvent := range dummyChannelBalanceEventByRemote {
							dummyChannelBalanceEvent.AggregatedLocalBalance = aggregateLocalBalance[remoteNodeId]
							dummyChannelBalanceEvent.AggregatedLocalBalancePerMilleRatio = aggregateLocalBalancePerMilleRatio[remoteNodeId]
							inputs := make(map[commons.WorkflowParameterLabel]string)
							marshalledChannelIdsFromEvents, err := json.Marshal(eventChannelIds)
							if err != nil {
								log.Error().Err(err).Msgf("Failed to marshal eventChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
								continue
							}
							inputs[commons.WorkflowParameterLabelChannels] = string(marshalledChannelIdsFromEvents)

							triggerCtx, triggerCancel := context.WithCancel(ctx)

							commons.ActivateEventTrigger(scheduledTrigger.Reference,
								workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId,
								scheduledTrigger.TriggeringNodeType, firstEvent, triggerCancel)

							processWorkflowNode(triggerCtx, db, triggerGroupWorkflowVersionNodeId,
								workflowTriggerNode, scheduledTrigger.Reference, inputs, lightningRequestChannel,
								rebalanceRequestChannel)

							triggerCancel()

							commons.DeactivateEventTrigger(workflowTriggerNode.WorkflowVersionId,
								workflowTriggerNode.WorkflowVersionNodeId, scheduledTrigger.TriggeringNodeType, lastEvent)

						}
					}
					continue
				}
			}

			inputs := make(map[commons.WorkflowParameterLabel]string)
			if len(eventChannelIds) != 0 {
				marshalledChannelIdsFromEvents, err := json.Marshal(eventChannelIds)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to marshal eventChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					continue
				}
				inputs[commons.WorkflowParameterLabelChannels] = string(marshalledChannelIdsFromEvents)
			} else {
				var allChannelIds []int
				torqNodeIds := commons.GetAllTorqNodeIds()
				for _, torqNodeId := range torqNodeIds {
					// Force Response because we don't care about balance accuracy
					channelIdsByNode := commons.GetChannelStateChannelIds(torqNodeId, true)
					allChannelIds = append(allChannelIds, channelIdsByNode...)
				}

				ba, err := json.Marshal(allChannelIds)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to marshal allChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					continue
				}
				inputs[commons.WorkflowParameterLabelChannels] = string(ba)
			}

			triggerCtx, triggerCancel := context.WithCancel(ctx)

			switch workflowTriggerNode.Type {
			case commons.WorkflowNodeIntervalTrigger:
				fallthrough
			case commons.WorkflowNodeCronTrigger:
				fallthrough
			case commons.WorkflowNodeManualTrigger:
				commons.ActivateWorkflowTrigger(scheduledTrigger.Reference,
					workflowTriggerNode.WorkflowVersionId, triggerCancel)
			default:
				commons.ActivateEventTrigger(scheduledTrigger.Reference,
					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId,
					scheduledTrigger.TriggeringNodeType, firstEvent, triggerCancel)
			}

			processWorkflowNode(triggerCtx, db, triggerGroupWorkflowVersionNodeId,
				workflowTriggerNode, scheduledTrigger.Reference, inputs, lightningRequestChannel,
				rebalanceRequestChannel)

			triggerCancel()

			switch workflowTriggerNode.Type {
			case commons.WorkflowNodeIntervalTrigger:
				fallthrough
			case commons.WorkflowNodeCronTrigger:
				fallthrough
			case commons.WorkflowNodeManualTrigger:
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
	go func() {
		for range ctx.Done() {
			broadcaster.CancelSubscriptionChannelBalanceEvent(listener)
			return
		}
	}()
	go func() {
		for channelBalanceEvent := range listener {
			if channelBalanceEvent.NodeId == 0 || channelBalanceEvent.ChannelId == 0 {
				continue
			}
			processEventTrigger(db, channelBalanceEvent, commons.WorkflowNodeChannelBalanceEventTrigger)
		}
	}()
}

func channelEventTriggerMonitor(ctx context.Context, db *sqlx.DB, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeChannelEvent()
	go func() {
		for range ctx.Done() {
			broadcaster.CancelSubscriptionChannelEvent(listener)
			return
		}
	}()
	go func() {
		for channelEvent := range listener {
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
	}()
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
	triggerGroupWorkflowVersionNodeId int, workflowTriggerNode workflows.WorkflowNode, reference string,
	inputs map[commons.WorkflowParameterLabel]string,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest) {

	workflowNodeStatus := make(map[int]commons.Status)
	workflowStageExitConfigurationCache := make(map[int]map[commons.WorkflowParameterLabel]string)
	workflowNodeCache := make(map[int]workflows.WorkflowNode)

	if workflowTriggerNode.Type != commons.WorkflowNodeManualTrigger {
		// Flag the trigger group node as processed
		workflowNodeStatus[triggerGroupWorkflowVersionNodeId] = commons.Active
	}

	workflowChannelBalanceTriggerNodes, err := workflows.GetActiveSortedStageTriggerNodeForWorkflowVersionId(db,
		workflowTriggerNode.WorkflowVersionId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain stage trigger nodes for WorkflowVersionId: %v", workflowTriggerNode.WorkflowVersionId)
		return
	}

	if len(workflowChannelBalanceTriggerNodes) == 0 {
		outputs, _, err := workflows.ProcessWorkflowNode(ctx, db, workflowTriggerNode,
			0, workflowNodeCache, workflowNodeStatus,
			reference, inputs, 0, workflowTriggerNode.Type, workflowStageExitConfigurationCache,
			lightningRequestChannel, rebalanceRequestChannel)
		workflows.AddWorkflowVersionNodeLog(db, reference, workflowTriggerNode.WorkflowVersionNodeId,
			0, inputs, outputs, err)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		}
		return
	}
	var channelIds []int
	err = json.Unmarshal([]byte(inputs[commons.WorkflowParameterLabelChannels]), &channelIds)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to unmarshal eventChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		return
	}
	for _, channelId := range channelIds {
		marshalledChannelIdsFromEvents, err := json.Marshal([]int{channelId})
		if err != nil {
			log.Error().Err(err).Msgf("Failed to marshal eventChannelId for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
			continue
		}
		inputs[commons.WorkflowParameterLabelChannels] = string(marshalledChannelIdsFromEvents)

		outputs, _, err := workflows.ProcessWorkflowNode(ctx, db, workflowTriggerNode,
			0, workflowNodeCache, workflowNodeStatus,
			reference, inputs, 0, workflowTriggerNode.Type, workflowStageExitConfigurationCache,
			lightningRequestChannel, rebalanceRequestChannel)
		workflows.AddWorkflowVersionNodeLog(db, reference, workflowTriggerNode.WorkflowVersionNodeId,
			0, inputs, outputs, err)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		}

		for _, workflowDeferredLinkNode := range workflowChannelBalanceTriggerNodes {
			inputs = commons.CopyParameters(outputs)
			outputs, _, err = workflows.ProcessWorkflowNode(ctx, db, workflowDeferredLinkNode,
				0, workflowNodeCache, workflowNodeStatus,
				reference, inputs, 0, workflowTriggerNode.Type, workflowStageExitConfigurationCache,
				lightningRequestChannel, rebalanceRequestChannel)
			workflows.AddWorkflowVersionNodeLog(db, reference, workflowDeferredLinkNode.WorkflowVersionNodeId,
				0, inputs, outputs, err)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to trigger deferred link node for WorkflowVersionNodeId: %v", workflowDeferredLinkNode.WorkflowVersionNodeId)
			}
		}
	}
}
