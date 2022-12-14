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

func TriggerMonitor(ctx context.Context, db *sqlx.DB, broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) error {
	ticker := clock.New().Tick(commons.WORKFLOW_TICKER_SECONDS * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker:
			var bootedWorkflowVersionIds []int
			workflowTriggerNodes, err := workflows.GetActiveTriggerNodes(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain root nodes (trigger nodes)")
				continue
			}
			for _, workflowTriggerNode := range workflowTriggerNodes {
				if slices.Contains(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId) {
					continue
				}
				shouldTrigger, err := workflows.ShouldTrigger(db, workflowTriggerNode)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to verify trigger activation for workflowVersionId: %v", workflowTriggerNode.WorkflowVersionNodeId)
					continue
				}
				if shouldTrigger {
					workflow, err := workflows.GetWorkflowByWorkflowVersionId(db, workflowTriggerNode.WorkflowVersionId)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to obtain workflow for workflowVersionId: %v", workflowTriggerNode.WorkflowVersionNodeId)
						continue
					}
					bootedWorkflowVersionIds = append(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId)
					triggerCtx, triggerCancel := context.WithCancel(context.Background())
					reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
					commons.SetTrigger(reference, workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)
					if workflow.Type == commons.WorkFlowDeferredLink {
						go func(
							ctx context.Context,
							db *sqlx.DB,
							workflowTriggerNode workflows.WorkflowNode,
							workflowTriggerNodes []workflows.WorkflowNode,
							reference string,
							eventChannel chan interface{}) {

							outputs, err := workflows.ProcessWorkflowNode(ctx, db, workflowTriggerNode, 0, reference, make(map[string]string), eventChannel, 0)
							workflows.AddWorkflowVersionNodeLog(db, reference, workflowTriggerNode.WorkflowVersionNodeId, nil, outputs, err)
							if err == nil {
								for _, workflowDeferredLinkNode := range workflowTriggerNodes {
									if workflowDeferredLinkNode.Type == commons.WorkflowNodeDeferredLink &&
										workflowDeferredLinkNode.WorkflowVersionId == workflowTriggerNode.WorkflowVersionId {
										deferredOutputs, err := workflows.ProcessWorkflowNode(ctx, db, workflowDeferredLinkNode, 0, reference, outputs, eventChannel, 0)
										workflows.AddWorkflowVersionNodeLog(db, reference, workflowDeferredLinkNode.WorkflowVersionNodeId, outputs, deferredOutputs, err)
										if err != nil {
											log.Error().Err(err).Msgf("Failed to trigger deferred link node for WorkflowVersionNodeId: %v", workflowDeferredLinkNode.WorkflowVersionNodeId)
										}
									}
								}
							} else {
								log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
							}
						}(triggerCtx, db, workflowTriggerNode, workflowTriggerNodes, reference, eventChannel)
					} else {
						go func(
							ctx context.Context,
							db *sqlx.DB,
							workflowTriggerNode workflows.WorkflowNode,
							reference string,
							eventChannel chan interface{}) {

							outputs, err := workflows.ProcessWorkflowNode(ctx, db, workflowTriggerNode, 0, reference, make(map[string]string), eventChannel, 0)
							workflows.AddWorkflowVersionNodeLog(db, reference, workflowTriggerNode.WorkflowVersionNodeId, nil, outputs, err)
							if err != nil {
								log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
							}
						}(triggerCtx, db, workflowTriggerNode, reference, eventChannel)
					}
				}
			}
		}
	}
}
