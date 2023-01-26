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
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/robfig/cron/v3"
)

func TimeTriggerMonitor(ctx context.Context, db *sqlx.DB) {

	defer log.Info().Msgf("TimeTriggerMonitor terminated")

	ticker := clock.New().Tick(commons.WORKFLOW_TICKER_SECONDS * time.Second)
	bootstrapping := true

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			var bootedWorkflowVersionIds []int
			workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, commons.WorkflowNodeTimeTrigger)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain root nodes (time trigger nodes)")
				continue
			}
			for _, workflowTriggerNode := range workflowTriggerNodes {
				reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
				triggerSettings := commons.GetTimeTriggerSettingsByWorkflowVersionId(workflowTriggerNode.WorkflowVersionId)
				var param workflows.TimeTriggerParameters
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
					workflowTriggerNode.Type, workflowTriggerNode)
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
						workflowTriggerNode.Type, workflowTriggerNode)
				}
				bootstrapping = false
			}
		}
	}
}

type CronTriggerParams struct {
	CronValue string `json:"cronValue"`
}

func CronTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings) {
	defer log.Info().Msgf("Cron trigger monitor terminated for nodeId: %v", nodeSettings.NodeId)

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
			commons.ScheduleTrigger(nodeSettings.NodeId, reference, trigger.WorkflowVersionId,
				commons.WorkflowNodeCronTrigger, trigger)
		})
		if err != nil {
			log.Error().Msgf("Unable to add cron func for workflow version node id: %v", trigger.WorkflowVersionNodeId)
			continue
		}
	}

	c.Start()

	log.Info().Msgf("Cron trigger monitor started for nodeId: %v", nodeSettings.NodeId)

	<-ctx.Done()
	c.Stop()
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

			var workflowVersionNodeId int
			switch scheduledTrigger.TriggeringNodeType {
			case commons.WorkflowNodeTimeTrigger:
				workflowVersionNodeId = firstEvent.(workflows.WorkflowNode).WorkflowVersionNodeId
			case commons.WorkflowNodeManualTrigger:
				workflowVersionNodeId = firstEvent.(commons.ManualTriggerEvent).WorkflowVersionNodeId
			default:
				workflowVersionNodeId = scheduledTrigger.TriggeringWorkflowVersionNodeId
				if workflowVersionNodeId == 0 {
					// Initial run without event when Torq boots
					workflowVersionNodeId = firstEvent.(workflows.WorkflowNode).WorkflowVersionNodeId
				}
			}

			workflowTriggerNode, err := workflows.GetWorkflowNode(db, workflowVersionNodeId)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to obtain the Triggering WorkflowNode for WorkflowVersionNodeId: %v", workflowVersionNodeId)
				continue
			}
			triggerGroupWorkflowVersionNodeId, err := workflows.GetTriggerGroupWorkflowVersionNodeId(db, workflowVersionNodeId)
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

			if scheduledTrigger.TriggeringNodeType != commons.WorkflowNodeTimeTrigger &&
				scheduledTrigger.TriggeringNodeType != commons.WorkflowNodeCronTrigger {
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
							marshalledChannelBalanceEvent, err := json.Marshal(dummyChannelBalanceEvent)
							if err != nil {
								log.Error().Err(err).Msgf("Failed to marshal ChannelBalanceEvent for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
								continue
							}
							inputs[commons.WorkflowParameterLabelChannelEventTriggered] = string(marshalledChannelBalanceEvent)
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

			marshalledEvents, err := json.Marshal(events)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to marshal events for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
				continue
			}
			if workflowTriggerNode.Type == commons.WorkflowNodeTimeTrigger {
				inputs[commons.WorkflowParameterLabelTimeTriggered] = string(marshalledEvents)
			} else {
				inputs[commons.WorkflowParameterLabelChannelEventTriggered] = string(marshalledEvents)
			}
			marshalledChannelIdsFromEvents, err := json.Marshal(eventChannelIds)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to marshal eventChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
				continue
			}
			inputs[commons.WorkflowParameterLabelChannels] = string(marshalledChannelIdsFromEvents)

			triggerCtx, triggerCancel := context.WithCancel(ctx)

			switch workflowTriggerNode.Type {
			case commons.WorkflowNodeTimeTrigger:
				commons.ActivateTimeTrigger(scheduledTrigger.Reference,
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
			case commons.WorkflowNodeTimeTrigger:
				commons.DeactivateTimeTrigger(workflowTriggerNode.WorkflowVersionId)
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
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionChannelBalanceEvent(listener)
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
	channelEventListener := broadcaster.SubscribeChannelEvent()
	for {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscriptionChannelEvent(channelEventListener)
			return
		case channelEvent := <-channelEventListener:
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
		commons.ScheduleTrigger(reference, workflowTriggerNode.WorkflowVersionId, workflowNodeType, triggeringEvent)
	}
}

func processWorkflowNode(ctx context.Context, db *sqlx.DB,
	triggerGroupWorkflowVersionNodeId int, workflowTriggerNode workflows.WorkflowNode, reference string,
	inputs map[commons.WorkflowParameterLabel]string,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest) {

	workflowNodeStatus := make(map[int]commons.Status)
	workflowNodeStagingParametersCache := make(map[int]map[commons.WorkflowParameterLabel]string)
	workflowStageExitConfigurationCache := make(map[int]map[commons.WorkflowParameterLabel]string)
	workflowNodeCache := make(map[int]workflows.WorkflowNode)

	// Flag the trigger group node as processed
	workflowNodeStatus[triggerGroupWorkflowVersionNodeId] = commons.Active

	outputs, _, err := workflows.ProcessWorkflowNode(ctx, db, workflowTriggerNode,
		0, workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache,
		reference, inputs, 0, workflowStageExitConfigurationCache,
		lightningRequestChannel, rebalanceRequestChannel)
	workflows.AddWorkflowVersionNodeLog(db, reference, workflowTriggerNode.WorkflowVersionNodeId,
		0, inputs, outputs, err)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		return
	}

	workflowChannelBalanceTriggerNodes, err := workflows.GetActiveSortedStageTriggerNodeForWorkflowVersionId(db,
		workflowTriggerNode.WorkflowVersionId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain stage trigger nodes for WorkflowVersionId: %v", workflowTriggerNode.WorkflowVersionId)
		return
	}

	for _, workflowDeferredLinkNode := range workflowChannelBalanceTriggerNodes {
		inputs = commons.CopyParameters(outputs)
		outputs, _, err = workflows.ProcessWorkflowNode(ctx, db, workflowDeferredLinkNode,
			0, workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache,
			reference, inputs, 0, workflowStageExitConfigurationCache,
			lightningRequestChannel, rebalanceRequestChannel)
		workflows.AddWorkflowVersionNodeLog(db, reference, workflowDeferredLinkNode.WorkflowVersionNodeId,
			0, inputs, outputs, err)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to trigger deferred link node for WorkflowVersionNodeId: %v", workflowDeferredLinkNode.WorkflowVersionNodeId)
		}
	}
}
