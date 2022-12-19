package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

func TimeTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}) {
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
					triggerParameters, err := workflows.GetWorkflowNodeParameters(workflowTriggerNode)
					if err != nil {
						log.Error().Err(err).Msgf("Obtaining trigger parameters for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
						continue
					}
					foundIt := false
					for _, triggerParameter := range triggerParameters.Parameters {
						if triggerParameter.Type == commons.WorkflowParameterTimeInSeconds && triggerParameter.ValueNumber != 0 {
							if triggerSettings.BootTime != nil && int(time.Since(*triggerSettings.BootTime).Seconds()) < triggerParameter.ValueNumber {
								continue
							}
							foundIt = true
						}
					}
					if !foundIt {
						log.Error().Err(err).Msgf("Trigger parameter could not be found for WorkflowVersionNodeId: %v.", workflowTriggerNode.WorkflowVersionNodeId)
						continue
					}
				}

				workflow, err := workflows.GetWorkflowByWorkflowVersionId(db, workflowTriggerNode.WorkflowVersionId)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to obtain workflow for workflowVersionId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					continue
				}
				bootedWorkflowVersionIds = append(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId)
				triggerCtx, triggerCancel := context.WithCancel(context.Background())
				reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
				commons.SetTrigger(nodeSettings.NodeId, reference,
					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId,
					commons.Active, triggerCancel)
				switch workflowTriggerNode.Type {
				case commons.WorkflowNodeTimeTrigger:
					inputs := make(map[string]string)
					marshalledTimerEvent, err := json.Marshal(workflowTriggerNode)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to marshal WorkflowNodeTimeTrigger for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
						continue
					}
					inputs["commons.WorkflowNodeTimeTrigger"] = string(marshalledTimerEvent)
					if workflow.Type == commons.WorkFlowDeferredLink {
						go processWorkflowNodeDeferredLinkRoutine(triggerCtx, db, nodeSettings,
							workflowTriggerNode, workflowTriggerNodes, reference, inputs, eventChannel)
					} else {
						go processWorkflowNodeRoutine(triggerCtx, db, nodeSettings,
							workflowTriggerNode, reference, inputs, eventChannel)
					}
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
							inputs := make(map[string]string)
							marshalledChannelBalanceEvent, err := json.Marshal(dummyChannelBalanceEvent)
							if err != nil {
								log.Error().Err(err).Msgf("Failed to marshal ChannelBalanceEvent for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
								continue
							}
							inputs["commons.ChannelBalanceEvent"] = string(marshalledChannelBalanceEvent)
							if workflow.Type == commons.WorkFlowDeferredLink {
								go processWorkflowNodeDeferredLinkRoutine(triggerCtx, db, nodeSettings,
									workflowTriggerNode, workflowTriggerNodes, reference, inputs, eventChannel)
							} else {
								go processWorkflowNodeRoutine(triggerCtx, db, nodeSettings,
									workflowTriggerNode, reference, inputs, eventChannel)
							}
						}
					}
				}
			}
			bootstrapping = false
		}
	}
}

func EventTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) {

	listener := broadcaster.Subscribe()
	for event := range listener {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscription(listener)
			return
		default:
		}

		if channelEvent, ok := event.(commons.ChannelBalanceEvent); ok {
			if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
				return
			}

			//if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
			// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
			//}

			var bootedWorkflowVersionIds []int
			workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, commons.WorkflowNodeChannelBalanceEventTrigger)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain root nodes (trigger nodes)")
				continue
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
				workflow, err := workflows.GetWorkflowByWorkflowVersionId(db, workflowTriggerNode.WorkflowVersionId)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to obtain workflow for workflowVersionId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					continue
				}
				bootedWorkflowVersionIds = append(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId)
				inputs := make(map[string]string)
				marshalledChannelBalanceEvent, err := json.Marshal(channelEvent)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to marshal channelEvent for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					continue
				}
				inputs["commons.ChannelBalanceEvent"] = string(marshalledChannelBalanceEvent)
				triggerCtx, triggerCancel := context.WithCancel(context.Background())
				reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
				commons.SetTrigger(nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionId,
					workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)
				if workflow.Type == commons.WorkFlowDeferredLink {
					go processWorkflowNodeDeferredLinkRoutine(triggerCtx, db, nodeSettings,
						workflowTriggerNode, workflowTriggerNodes, reference, inputs, eventChannel)
				} else {
					go processWorkflowNodeRoutine(triggerCtx, db, nodeSettings,
						workflowTriggerNode, reference, inputs, eventChannel)
				}
			}
		}
	}
}

func processWorkflowNodeRoutine(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	workflowTriggerNode workflows.WorkflowNode, reference string, inputs map[string]string,
	eventChannel chan interface{}) {

	workflowNodeCache := make(map[int]workflows.WorkflowNode)
	workflowNodeStatus := make(map[int]commons.Status)
	workflowNodeStagingParametersCache := make(map[int]map[string]string)
	outputs, _, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode,
		0, workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache, reference, inputs, eventChannel, 0)
	workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId,
		0, inputs, outputs, err)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		return
	}
}

func processWorkflowNodeDeferredLinkRoutine(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	workflowTriggerNode workflows.WorkflowNode, workflowTriggerNodes []workflows.WorkflowNode, reference string,
	inputs map[string]string, eventChannel chan interface{}) {

	workflowNodeCache := make(map[int]workflows.WorkflowNode)
	workflowNodeStatus := make(map[int]commons.Status)
	workflowNodeStagingParametersCache := make(map[int]map[string]string)
	outputs, _, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode,
		0, workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache, reference, inputs, eventChannel, 0)
	workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId,
		0, inputs, outputs, err)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		return
	}
	for _, workflowDeferredLinkNode := range workflowTriggerNodes {
		if workflowDeferredLinkNode.Type == commons.WorkflowNodeDeferredLink &&
			workflowDeferredLinkNode.WorkflowVersionId == workflowTriggerNode.WorkflowVersionId {
			deferredOutputs, _, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowDeferredLinkNode,
				0, workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache, reference, outputs, eventChannel, 0)
			workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowDeferredLinkNode.WorkflowVersionNodeId,
				0, outputs, deferredOutputs, err)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to trigger deferred link node for WorkflowVersionNodeId: %v", workflowDeferredLinkNode.WorkflowVersionNodeId)
			}
		}
	}
}
