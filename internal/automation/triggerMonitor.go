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
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

func TimeTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings) {

	defer log.Info().Msgf("TimeTriggerMonitor terminated for nodeId: %v", nodeSettings.NodeId)

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
			// When bootstrapping the automations run ALL workflows because we might have missed some events.
			if bootstrapping {
				workflowChannelBalanceTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, commons.WorkflowNodeChannelBalanceEventTrigger)
				if err != nil {
					log.Error().Err(err).Msg("Failed to obtain root nodes (channel balance trigger nodes)")
					continue
				}
				workflowTriggerNodes = append(workflowTriggerNodes, workflowChannelBalanceTriggerNodes...)
			}
			for _, workflowTriggerNode := range workflowTriggerNodes {
				commons.SetTriggerVerificationTime(nodeSettings.NodeId, workflowTriggerNode.WorkflowVersionId, time.Now())
				if slices.Contains(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId) {
					continue
				}
				triggerSettings := commons.GetTriggerSettingsByWorkflowVersionId(nodeSettings.NodeId, workflowTriggerNode.WorkflowVersionId)
				if triggerSettings.Status == commons.Active {
					log.Error().Err(err).Msgf("Trigger is already active with reference: %v.", triggerSettings.Reference)
					continue
				}
				if workflowTriggerNode.Type == commons.WorkflowNodeTimeTrigger {
					var param workflows.TimeTriggerParameters
					err := json.Unmarshal([]byte(workflowTriggerNode.Parameters.([]uint8)), &param)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to parse parameters for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
						continue
					}
					if triggerSettings.BootTime != nil && int32(time.Since(*triggerSettings.BootTime).Seconds()) < param.Seconds {
						continue
					}
				}

				bootedWorkflowVersionIds = append(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId)
				triggerCtx, triggerCancel := context.WithCancel(ctx)
				reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))

				triggerGroupWorkflowVersionNodeId, err := workflows.GetTriggerGroupWorkflowVersionNodeId(db, workflowTriggerNode.WorkflowVersionNodeId)
				if err != nil || triggerGroupWorkflowVersionNodeId == 0 {
					log.Error().Err(err).Msgf("Failed to obtain the group node id for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					triggerCancel()
					continue
				}
				groupWorkflowVersionNode, err := workflows.GetWorkflowNode(db, triggerGroupWorkflowVersionNodeId)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to obtain the group nodes links for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					triggerCancel()
					continue
				}
				workflowTriggerNode.ChildNodes = groupWorkflowVersionNode.ChildNodes
				workflowTriggerNode.LinkDetails = groupWorkflowVersionNode.LinkDetails
				commons.SetTrigger(nodeSettings.NodeId, reference,
					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)

				switch workflowTriggerNode.Type {
				case commons.WorkflowNodeTimeTrigger:
					inputs := make(map[commons.WorkflowParameterLabel]string)
					marshalledTimerEvent, err := json.Marshal(workflowTriggerNode)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to marshal WorkflowNodeTimeTrigger for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
						triggerCancel()
						continue
					}
					inputs[commons.WorkflowParameterLabelTimeTriggered] = string(marshalledTimerEvent)
					processWorkflowNode(triggerCtx, db, nodeSettings, workflowTriggerNode, reference, inputs)
				case commons.WorkflowNodeChannelBalanceEventTrigger:
					eventTime := time.Now()
					dummyChannelBalanceEvents := make(map[int]map[int]*commons.ChannelBalanceEvent)
					for _, channelId := range commons.GetChannelIdsByNodeId(nodeSettings.NodeId) {
						channelSettings := commons.GetChannelSettingByChannelId(channelId)
						capacity := channelSettings.Capacity
						remoteNodeId := channelSettings.FirstNodeId
						if remoteNodeId == nodeSettings.NodeId {
							remoteNodeId = channelSettings.SecondNodeId
						}
						if dummyChannelBalanceEvents[remoteNodeId] == nil {
							dummyChannelBalanceEvents[remoteNodeId] = make(map[int]*commons.ChannelBalanceEvent)
						}
						channelState := commons.GetChannelState(nodeSettings.NodeId, channelId, true)
						dummyChannelBalanceEvents[remoteNodeId][channelId] = &commons.ChannelBalanceEvent{
							EventData: commons.EventData{
								EventTime: eventTime,
								NodeId:    nodeSettings.NodeId,
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
					aggregateLocalBalance := make(map[int]int64)
					aggregateLocalBalancePerMilleRatio := make(map[int]int)
					for remoteNodeId, dummyChannelBalanceEventByRemote := range dummyChannelBalanceEvents {
						var localBalanceAggregate int64
						var capacityAggregate int64
						for _, dummyChannelBalanceEvent := range dummyChannelBalanceEventByRemote {
							localBalanceAggregate += dummyChannelBalanceEvent.LocalBalance
							capacityAggregate += dummyChannelBalanceEvent.Capacity
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
								triggerCancel()
								continue
							}
							inputs[commons.WorkflowParameterLabelChannelEventTriggered] = string(marshalledChannelBalanceEvent)
							processWorkflowNode(triggerCtx, db, nodeSettings, workflowTriggerNode, reference, inputs)
						}
					}
				}
				commons.SetTrigger(nodeSettings.NodeId, reference,
					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId, commons.Inactive, triggerCancel)
				triggerCancel()
			}
			bootstrapping = false
		}
	}
}

func EventTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer) {

	defer log.Info().Msgf("EventTriggerMonitor terminated for nodeId: %v", nodeSettings.NodeId)

	var wg sync.WaitGroup

	wg.Add(1)
	go (func() {
		defer wg.Done()
		channelBalanceEventTriggerMonitor(ctx, db, nodeSettings, broadcaster)
	})()

	wg.Add(1)
	go (func() {
		defer wg.Done()
		channelEventTriggerMonitor(ctx, db, nodeSettings, broadcaster)
	})()

	wg.Wait()
}

func channelBalanceEventTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer) {

	defer log.Info().Msgf("EventTriggerMonitor terminated for nodeId: %v", nodeSettings.NodeId)

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
			processEventTrigger(ctx, db, nodeSettings, channelBalanceEvent, commons.WorkflowNodeChannelBalanceEventTrigger)
		}
	}
}

func channelEventTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer) {

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
			if channelEvent.Type == lnrpc.ChannelEventUpdate_OPEN_CHANNEL {
				processEventTrigger(ctx, db, nodeSettings, channelEvent, commons.WorkflowNodeChannelOpenEventTrigger)
			}
			if channelEvent.Type == lnrpc.ChannelEventUpdate_CLOSED_CHANNEL {
				processEventTrigger(ctx, db, nodeSettings, channelEvent, commons.WorkflowNodeChannelCloseEventTrigger)
			}
		}
	}
}

func processEventTrigger(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	triggeringEvent any, workflowNodeType commons.WorkflowNodeType) {

	//if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
	// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
	//}

	var bootedWorkflowVersionIds []int
	workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, workflowNodeType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain root nodes (trigger nodes)")
		return
	}
	for _, workflowTriggerNode := range workflowTriggerNodes {
		if slices.Contains(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId) {
			continue
		}
		triggerSettings := commons.GetTriggerSettingsByWorkflowVersionId(nodeSettings.NodeId, workflowTriggerNode.WorkflowVersionId)
		if triggerSettings.Status == commons.Active {
			log.Info().Msgf("Trigger is already active with reference: %v.", triggerSettings.Reference)
			continue
		}
		commons.SetTriggerVerificationTime(nodeSettings.NodeId, workflowTriggerNode.WorkflowVersionId, time.Now())
		bootedWorkflowVersionIds = append(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId)
		inputs := make(map[commons.WorkflowParameterLabel]string)
		marshalledEvent, err := json.Marshal(triggeringEvent)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to marshal event for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
			continue
		}
		inputs[commons.WorkflowParameterLabelChannelEventTriggered] = string(marshalledEvent)
		reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
		triggerCtx, triggerCancel := context.WithCancel(ctx)
		commons.SetTrigger(nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionId,
			workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)
		processWorkflowNode(triggerCtx, db, nodeSettings, workflowTriggerNode, reference, inputs)
		commons.SetTrigger(nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionId,
			workflowTriggerNode.WorkflowVersionNodeId, commons.Inactive, triggerCancel)
		triggerCancel()
	}
}

func processWorkflowNode(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	workflowTriggerNode workflows.WorkflowNode, reference string, inputs map[commons.WorkflowParameterLabel]string) {

	workflowNodeStatus := make(map[int]commons.Status)
	workflowNodeStagingParametersCache := make(map[int]map[commons.WorkflowParameterLabel]string)

	outputs, _, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode,
		0, workflowNodeStatus, workflowNodeStagingParametersCache, reference, inputs, 0)
	workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId,
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
		outputs, _, err = workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowDeferredLinkNode,
			0, workflowNodeStatus, workflowNodeStagingParametersCache, reference, inputs, 0)
		workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowDeferredLinkNode.WorkflowVersionNodeId,
			0, inputs, outputs, err)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to trigger deferred link node for WorkflowVersionNodeId: %v", workflowDeferredLinkNode.WorkflowVersionNodeId)
		}
	}

}
