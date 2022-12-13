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
					bootedWorkflowVersionIds = append(bootedWorkflowVersionIds, workflowTriggerNode.WorkflowVersionId)
					triggerCtx, triggerCancel := context.WithCancel(context.Background())
					reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
					commons.SetTrigger(reference, workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)
					go func(ctx context.Context, db *sqlx.DB, workflowTriggerNode workflows.WorkflowNode, reference string, eventChannel chan interface{}) {
						err := workflows.Trigger(ctx, db, workflowTriggerNode, reference, eventChannel)
						if err != nil {
							log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) %v", workflowTriggerNode.WorkflowVersionNodeId)
						}
						workflowVersionNodeLog := workflows.WorkflowVersionNodeLog{
							WorkflowVersionNodeId: workflowTriggerNode.WorkflowVersionNodeId,
							TriggerReference:      reference,
						}
						if err != nil {
							workflowVersionNodeLog.ErrorData = err.Error()
						}
						_, err = workflows.AddWorkflowVersionNodeLog(db, workflowVersionNodeLog)
						if err != nil {
							log.Error().Err(err).Msgf("Failed to log root node execution %v", workflowTriggerNode.WorkflowVersionNodeId)
						}
					}(triggerCtx, db, workflowTriggerNode, reference, eventChannel)
				}
			}
		}
	}
}
