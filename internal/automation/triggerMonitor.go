package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lnd"
	"github.com/lncapital/torq/internal/workflows"
)

const workflowTickerSeconds = 10

func IntervalTriggerMonitor(ctx context.Context, db *sqlx.DB) {

	ticker := time.NewTicker(workflowTickerSeconds * time.Second)
	defer ticker.Stop()
	bootstrapping := true

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if bootstrapping {
				var allChannelIds []int
				torqNodeIds := cache.GetAllTorqNodeIds()
				for _, torqNodeId := range torqNodeIds {
					// Force Response because we don't care about balance accuracy
					channelIdsByNode := cache.GetChannelStateNotSharedChannelIds(torqNodeId, true)
					allChannelIds = append(allChannelIds, channelIdsByNode...)
				}
				if len(allChannelIds) == 0 {
					continue
				}
			}
			var bootedWorkflowVersionIds []int
			workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, core.WorkflowNodeIntervalTrigger)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain root nodes (interval trigger nodes)")
				continue
			}
			for _, workflowTriggerNode := range workflowTriggerNodes {
				reference := fmt.Sprintf("%v_%v", workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
				triggerSettings := cache.GetTimeTriggerSettingsByWorkflowVersionId(workflowTriggerNode.WorkflowVersionId)
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
				cache.ScheduleTrigger(reference, workflowTriggerNode.WorkflowVersionId,
					workflowTriggerNode.Type, workflowTriggerNode.WorkflowVersionNodeId, workflowTriggerNode)
			}
			// When bootstrapping the automations run ALL workflows because we might have missed some events.
			if bootstrapping {
				workflowTriggerNodes, err = workflows.GetActiveEventTriggerNodes(db, core.WorkflowNodeChannelBalanceEventTrigger)
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
					cache.ScheduleTrigger(reference,
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

	serviceType := core.CronService

	ticker := time.NewTicker(workflowTickerSeconds * time.Second)
	defer ticker.Stop()

bootstrappingLoop:
	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveCoreServiceState(serviceType)
			return
		case <-ticker.C:
			torqNodeIds := cache.GetAllTorqNodeIds()
			for _, torqNodeId := range torqNodeIds {
				// Force Response because we don't care about balance accuracy
				if len(cache.GetChannelStateNotSharedChannelIds(torqNodeId, true)) != 0 {
					break bootstrappingLoop
				}
			}
		}
	}

	workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, core.WorkflowNodeCronTrigger)
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain root nodes (cron trigger nodes)")
		log.Error().Msg("Cron trigger monitor failed to start")
		cache.SetFailedCoreServiceState(serviceType)
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
			cache.ScheduleTrigger(reference, workflowVersionId, core.WorkflowNodeCronTrigger, workflowVersionNodeId, triggeringEvent)
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
	cache.SetInactiveCoreServiceState(serviceType)
}

func ScheduledTriggerMonitor(ctx context.Context, db *sqlx.DB) {

	var delay bool

	for {
		if delay {
			core.Sleep(ctx, 1*time.Second)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		scheduledTrigger := cache.GetScheduledTrigger()
		if scheduledTrigger.SchedulingTime == nil {
			log.Trace().Msg("ScheduledTriggerMonitor couldn't find any pending triggers")
			delay = true
			continue
		}
		delay = false
		events := scheduledTrigger.TriggeringEventQueue
		if len(events) == 0 {
			log.Error().Msgf("ScheduledTriggerMonitor initiated but no event found for WorkflowVersionId: %v",
				scheduledTrigger.WorkflowVersionId)
			continue
		}

		log.Debug().Msgf("ScheduledTriggerMonitor initiated for %v events", len(events))

		workflowTriggerNode, err := workflows.GetWorkflowNode(db, scheduledTrigger.TriggeringWorkflowVersionNodeId)
		if err != nil {
			log.Error().Err(err).Msgf(
				"ScheduledTriggerMonitor could not obtain the Triggering WorkflowNode for WorkflowVersionNodeId: %v",
				scheduledTrigger.TriggeringWorkflowVersionNodeId)
			continue
		}
		triggerGroupWorkflowVersionNodeId, err := workflows.GetTriggerGroupWorkflowVersionNodeId(db,
			scheduledTrigger.TriggeringWorkflowVersionNodeId)
		if err != nil || triggerGroupWorkflowVersionNodeId == 0 {
			log.Error().Err(err).Msgf(
				"ScheduledTriggerMonitor could not obtain the group node id for WorkflowVersionNodeId: %v",
				scheduledTrigger.TriggeringWorkflowVersionNodeId)
			continue
		}
		groupWorkflowVersionNode, err := workflows.GetWorkflowNode(db, triggerGroupWorkflowVersionNodeId)
		if err != nil {
			log.Error().Err(err).Msgf(
				"ScheduledTriggerMonitor could not obtain the group WorkflowNode for "+
					"triggerGroupWorkflowVersionNodeId: %v", triggerGroupWorkflowVersionNodeId)
			continue
		}
		workflowTriggerNode.ChildNodes = groupWorkflowVersionNode.ChildNodes
		workflowTriggerNode.ParentNodes = make(map[int]*workflows.WorkflowNode)
		workflowTriggerNode.LinkDetails = groupWorkflowVersionNode.LinkDetails

		// Override the default WorkflowTrigger since you should never directly run the group node
		if scheduledTrigger.TriggeringNodeType == core.WorkflowNodeManualTrigger &&
			workflowTriggerNode.Type == core.WorkflowTrigger {

			workflowTriggerNode.Type = core.WorkflowNodeManualTrigger
		}

		triggerCtx, triggerCancel := context.WithCancel(ctx)

		switch workflowTriggerNode.Type {
		case core.WorkflowNodeIntervalTrigger, core.WorkflowNodeCronTrigger, core.WorkflowNodeManualTrigger:
			log.Trace().Msgf("ScheduledTriggerMonitor activating trigger with WorkflowVersionId: %v",
				workflowTriggerNode.WorkflowVersionId)
			cache.ActivateWorkflowTrigger(scheduledTrigger.Reference,
				workflowTriggerNode.WorkflowVersionId, triggerCancel)
		default:
			log.Trace().Msgf("ScheduledTriggerMonitor activating event trigger with WorkflowVersionId: %v",
				workflowTriggerNode.WorkflowVersionId)
			cache.ActivateEventTrigger(scheduledTrigger.Reference,
				workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.WorkflowVersionNodeId,
				scheduledTrigger.TriggeringNodeType, events[0], triggerCancel)
		}

		processWorkflowNode(triggerCtx, db, workflowTriggerNode, scheduledTrigger.Reference, events)

		triggerCancel()

		switch workflowTriggerNode.Type {
		case core.WorkflowNodeIntervalTrigger, core.WorkflowNodeCronTrigger, core.WorkflowNodeManualTrigger:
			log.Trace().Msgf("ScheduledTriggerMonitor deactivating trigger with WorkflowVersionId: %v",
				workflowTriggerNode.WorkflowVersionId)
			cache.DeactivateWorkflowTrigger(workflowTriggerNode.WorkflowVersionId)
		default:
			log.Trace().Msgf("ScheduledTriggerMonitor deactivating event trigger with WorkflowVersionId: %v",
				workflowTriggerNode.WorkflowVersionId)
			cache.DeactivateEventTrigger(workflowTriggerNode.WorkflowVersionId,
				workflowTriggerNode.WorkflowVersionNodeId, scheduledTrigger.TriggeringNodeType, events[0])
		}
	}
}

func ChannelBalanceEventTriggerMonitor(ctx context.Context, db *sqlx.DB) {
	for {
		select {
		case <-ctx.Done():
			return
		case channelBalanceEvent := <-cache.ChannelBalanceChanges:
			if channelBalanceEvent.NodeId == 0 || channelBalanceEvent.ChannelId == 0 {
				continue
			}
			processEventTrigger(db, channelBalanceEvent, core.WorkflowNodeChannelBalanceEventTrigger)
		}
	}
}

func ChannelEventTriggerMonitor(ctx context.Context, db *sqlx.DB) {
	for {
		select {
		case <-ctx.Done():
			return
		case channelEvent := <-lnd.ChannelChanges:
			if channelEvent.NodeId == 0 || channelEvent.ChannelId == 0 {
				continue
			}
			switch channelEvent.Type {
			case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
				processEventTrigger(db, channelEvent, core.WorkflowNodeChannelOpenEventTrigger)
			case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
				processEventTrigger(db, channelEvent, core.WorkflowNodeChannelCloseEventTrigger)
			}
		}
	}
}

func processEventTrigger(db *sqlx.DB, triggeringEvent any, workflowNodeType core.WorkflowNodeType) {

	//if LndService.GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
	// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
	//}

	workflowTriggerNodes, err := workflows.GetActiveEventTriggerNodes(db, workflowNodeType)
	if err != nil {
		log.Error().Err(err).Msg("ScheduledTriggerMonitor failed to obtain root nodes (trigger nodes)")
		return
	}
	for _, workflowTriggerNode := range workflowTriggerNodes {
		reference := fmt.Sprintf("%v_%v",
			workflowTriggerNode.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
		cache.ScheduleTrigger(reference, workflowTriggerNode.WorkflowVersionId, workflowNodeType,
			workflowTriggerNode.WorkflowVersionNodeId, triggeringEvent)
	}
}

func processWorkflowNode(ctx context.Context, db *sqlx.DB,
	workflowTriggerNode workflows.WorkflowNode,
	reference string,
	events []any) {

	err := workflows.ProcessWorkflow(ctx, db, workflowTriggerNode, reference, events)
	if err != nil {
		log.Error().Err(err).Msgf(
			"ScheduledTriggerMonitor failed to trigger nodes for WorkflowVersionNodeId: %v",
			workflowTriggerNode.WorkflowVersionNodeId)
	}
}
