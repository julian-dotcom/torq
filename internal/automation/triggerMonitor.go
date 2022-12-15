package automation

import (
	"context"
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
					if workflow.Type == commons.WorkFlowDeferredLink {
						go processWorkflowNodeDeferredLinkRoutine(triggerCtx, db, nodeSettings,
							workflowTriggerNode, workflowTriggerNodes, reference, 0, eventChannel)
					} else {
						go processWorkflowNodeRoutine(triggerCtx, db, nodeSettings,
							workflowTriggerNode, reference, 0, eventChannel)
					}
				case commons.WorkflowNodeChannelBalanceEventTrigger:
					for _, channelId := range commons.GetChannelIdsByNodeId(nodeSettings.NodeId) {
						if workflow.Type == commons.WorkFlowDeferredLink {
							go processWorkflowNodeDeferredLinkRoutine(triggerCtx, db, nodeSettings,
								workflowTriggerNode, workflowTriggerNodes, reference, channelId, eventChannel)
						} else {
							go processWorkflowNodeRoutine(triggerCtx, db, nodeSettings,
								workflowTriggerNode, reference, channelId, eventChannel)
						}
					}
				}
			}
			bootstrapping = false
		}
	}
}

func EventTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings, broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) {
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
				triggerCtx, triggerCancel := context.WithCancel(context.Background())
				reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
				commons.SetTrigger(nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)
				if workflow.Type == commons.WorkFlowDeferredLink {
					go processWorkflowNodeDeferredLinkRoutine(triggerCtx, db, nodeSettings,
						workflowTriggerNode, workflowTriggerNodes, reference, channelEvent.ChannelId, eventChannel)
				} else {
					go processWorkflowNodeRoutine(triggerCtx, db, nodeSettings,
						workflowTriggerNode, reference, channelEvent.ChannelId, eventChannel)
				}
			}
		}
	}
}

func processWorkflowNodeRoutine(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	workflowTriggerNode workflows.WorkflowNode, reference string, channelId int, eventChannel chan interface{}) {

	outputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode,
		0, reference, make(map[string]string), eventChannel, 0)
	workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId,
		0, nil, outputs, err)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		return
	}
}

func processWorkflowNodeDeferredLinkRoutine(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings,
	workflowTriggerNode workflows.WorkflowNode, workflowTriggerNodes []workflows.WorkflowNode, reference string,
	channelId int, eventChannel chan interface{}) {

	outputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode,
		0, reference, make(map[string]string), eventChannel, 0)
	workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId,
		0, nil, outputs, err)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		return
	}
	for _, workflowDeferredLinkNode := range workflowTriggerNodes {
		if workflowDeferredLinkNode.Type == commons.WorkflowNodeDeferredLink &&
			workflowDeferredLinkNode.WorkflowVersionId == workflowTriggerNode.WorkflowVersionId {
			deferredOutputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowDeferredLinkNode,
				0, reference, outputs, eventChannel, 0)
			workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowDeferredLinkNode.WorkflowVersionNodeId,
				0, outputs, deferredOutputs, err)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to trigger deferred link node for WorkflowVersionNodeId: %v", workflowDeferredLinkNode.WorkflowVersionNodeId)
			}
		}
	}
}
