package automation

import (
	"context"
	"fmt"
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

func TimeTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings, eventChannel chan interface{}) {
	ticker := clock.New().Tick(commons.WORKFLOW_TICKER_SECONDS * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
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
					commons.SetTrigger(nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)
					if workflow.Type == commons.WorkFlowDeferredLink {
						go func(
							ctx context.Context,
							db *sqlx.DB,
							workflowTriggerNode workflows.WorkflowNode,
							workflowTriggerNodes []workflows.WorkflowNode,
							reference string,
							eventChannel chan interface{}) {

							outputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode, 0, reference, make(map[string]string), eventChannel, 0)
							workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId, nil, outputs, err)
							if err == nil {
								for _, workflowDeferredLinkNode := range workflowTriggerNodes {
									if workflowDeferredLinkNode.Type == commons.WorkflowNodeDeferredLink &&
										workflowDeferredLinkNode.WorkflowVersionId == workflowTriggerNode.WorkflowVersionId {
										deferredOutputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowDeferredLinkNode, 0, reference, outputs, eventChannel, 0)
										workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowDeferredLinkNode.WorkflowVersionNodeId, outputs, deferredOutputs, err)
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

							outputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode, 0, reference, make(map[string]string), eventChannel, 0)
							workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId, nil, outputs, err)
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

func EventTriggerMonitor(ctx context.Context, db *sqlx.DB, nodeSettings commons.ManagedNodeSettings, broadcaster broadcast.BroadcastServer, eventChannel chan interface{}) {
	listener := broadcaster.Subscribe()
	for event := range listener {
		select {
		case <-ctx.Done():
			broadcaster.CancelSubscription(listener)
			return
		default:
		}

		if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
			// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
		}

		if channelEvent, ok := event.(commons.ChannelEvent); ok {
			if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
				return
			}

			if channelEvent.Type == lnrpc.ChannelEventUpdate_CLOSED_CHANNEL {
				// TODO FIXME kill some automations?
			}
		} else if forwardEvent, ok := event.(commons.ForwardEvent); ok {
			if forwardEvent.NodeId == 0 {
				return
			}
			if forwardEvent.IncomingChannelId != nil {
				// TODO FIXME check channel
			}
			if forwardEvent.OutgoingChannelId != nil {
				// TODO FIXME check channel
			}
		} else if invoiceEvent, ok := event.(commons.InvoiceEvent); ok {
			if invoiceEvent.NodeId == 0 || invoiceEvent.State != lnrpc.Invoice_SETTLED {
				return
			}
			// TODO FIXME check channel
		} else if paymentEvent, ok := event.(commons.PaymentEvent); ok {
			if paymentEvent.NodeId == 0 || paymentEvent.OutgoingChannelId == nil || *paymentEvent.OutgoingChannelId == 0 || paymentEvent.PaymentStatus != lnrpc.Payment_SUCCEEDED {
				return
			}
			// TODO FIXME check channel
		} else if closeChannelEvent, ok := event.(commons.CloseChannelResponse); ok {
			if closeChannelEvent.Request.NodeId == 0 {
				return
			}
			// TODO FIXME kill some automations?
		}

		// LOOP OVER THE CACHED WORKFLOW BALANCE
		// WHEN BALANCE FALLS OUT OF BOUNDS RUN THE ASSOCIATED WORKFLOW

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
				commons.SetTrigger(nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId, commons.Active, triggerCancel)
				if workflow.Type == commons.WorkFlowDeferredLink {
					go func(
						ctx context.Context,
						db *sqlx.DB,
						workflowTriggerNode workflows.WorkflowNode,
						workflowTriggerNodes []workflows.WorkflowNode,
						reference string,
						eventChannel chan interface{}) {

						outputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode, 0, reference, make(map[string]string), eventChannel, 0)
						workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId, nil, outputs, err)
						if err == nil {
							for _, workflowDeferredLinkNode := range workflowTriggerNodes {
								if workflowDeferredLinkNode.Type == commons.WorkflowNodeDeferredLink &&
									workflowDeferredLinkNode.WorkflowVersionId == workflowTriggerNode.WorkflowVersionId {
									deferredOutputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowDeferredLinkNode, 0, reference, outputs, eventChannel, 0)
									workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowDeferredLinkNode.WorkflowVersionNodeId, outputs, deferredOutputs, err)
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

						outputs, err := workflows.ProcessWorkflowNode(ctx, db, nodeSettings, workflowTriggerNode, 0, reference, make(map[string]string), eventChannel, 0)
						workflows.AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference, workflowTriggerNode.WorkflowVersionNodeId, nil, outputs, err)
						if err != nil {
							log.Error().Err(err).Msgf("Failed to trigger root nodes (trigger nodes) for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
						}
					}(triggerCtx, db, workflowTriggerNode, reference, eventChannel)
				}
			}
		}
	}
}
