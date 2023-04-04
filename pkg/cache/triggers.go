package cache

import (
	"context"
	"sort"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/core"
)

var TriggersCacheChannel = make(chan TriggerCache) //nolint:gochecknoglobals

type TriggerCacheOperationType uint
type workflowVersionId int

const (
	readTimeTriggerSettings TriggerCacheOperationType = iota
	writeTimeTrigger
	readEventTriggerSettings
	writeEventTrigger
	popScheduledTrigger
	writeScheduledTrigger
)

type TriggerCache struct {
	Type                            TriggerCacheOperationType
	WorkflowVersionId               int
	TriggeringWorkflowVersionNodeId int
	TriggeringNodeType              core.WorkflowNodeType
	TriggeringEvent                 any
	Reference                       string
	CancelFunction                  context.CancelFunc
	BootTime                        *time.Time
	VerificationTime                *time.Time
	Status                          core.Status
	PreviousState                   core.Status
	TriggerSettingsOut              chan<- TriggerSettingsCache
}

type TriggerSettingsCache struct {
	WorkflowVersionId               int
	TriggeringWorkflowVersionNodeId int
	TriggeringNodeType              core.WorkflowNodeType
	Reference                       string
	CancelFunction                  context.CancelFunc
	BootTime                        *time.Time
	VerificationTime                *time.Time
	SchedulingTime                  *time.Time
	TriggeringEventQueue            []any
	Status                          core.Status
	PreviousState                   core.Status
}

func TriggersCacheHandler(ch <-chan TriggerCache, ctx context.Context) {
	timeTriggerCache := make(map[workflowVersionId]TriggerSettingsCache)
	eventTriggerCache := make(map[workflowVersionId]map[int]map[int]TriggerSettingsCache)
	var scheduledTriggerCache []TriggerSettingsCache
	for {
		select {
		case <-ctx.Done():
			return
		case triggerCache := <-ch:
			scheduledTriggerCache = handelTriggerOperation(triggerCache, timeTriggerCache, eventTriggerCache, scheduledTriggerCache)
		}
	}
}

func handelTriggerOperation(triggerCache TriggerCache,
	timeTriggerCache map[workflowVersionId]TriggerSettingsCache,
	eventTriggerCache map[workflowVersionId]map[int]map[int]TriggerSettingsCache,
	scheduledTriggerCache []TriggerSettingsCache) []TriggerSettingsCache {

	switch triggerCache.Type {
	case readTimeTriggerSettings:
		if triggerCache.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", triggerCache.WorkflowVersionId)
			triggerCache.TriggerSettingsOut <- TriggerSettingsCache{}
			return scheduledTriggerCache
		}
		triggerCache.TriggerSettingsOut <- timeTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)]
	case writeTimeTrigger:
		if triggerCache.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", triggerCache.WorkflowVersionId)
			return scheduledTriggerCache
		}
		timeTriggerSettings, exists := timeTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)]
		if !exists {
			timeTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)] = TriggerSettingsCache{
				WorkflowVersionId:               triggerCache.WorkflowVersionId,
				TriggeringWorkflowVersionNodeId: triggerCache.TriggeringWorkflowVersionNodeId,
				TriggeringNodeType:              triggerCache.TriggeringNodeType,
				TriggeringEventQueue:            []any{triggerCache.TriggeringEvent},
				Reference:                       triggerCache.Reference,
				Status:                          triggerCache.Status,
				CancelFunction:                  triggerCache.CancelFunction,
				BootTime:                        triggerCache.BootTime,
			}
			return scheduledTriggerCache
		}

		if timeTriggerSettings.Status != triggerCache.Status {
			timeTriggerSettings.PreviousState = timeTriggerSettings.Status
		}
		if triggerCache.BootTime != nil {
			timeTriggerSettings.BootTime = triggerCache.BootTime
		}
		timeTriggerSettings.TriggeringWorkflowVersionNodeId = triggerCache.TriggeringWorkflowVersionNodeId
		timeTriggerSettings.TriggeringNodeType = triggerCache.TriggeringNodeType
		timeTriggerSettings.TriggeringEventQueue = append(timeTriggerSettings.TriggeringEventQueue, triggerCache.TriggeringEvent)
		timeTriggerSettings.Reference = triggerCache.Reference
		timeTriggerSettings.Status = triggerCache.Status
		timeTriggerSettings.CancelFunction = triggerCache.CancelFunction
		timeTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)] = timeTriggerSettings

	case readEventTriggerSettings:
		if triggerCache.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", triggerCache.WorkflowVersionId)
			triggerCache.TriggerSettingsOut <- TriggerSettingsCache{}
			return scheduledTriggerCache
		}

		initializeEventTriggerCache(eventTriggerCache, triggerCache.WorkflowVersionId, triggerCache.TriggeringWorkflowVersionNodeId)

		triggerReferenceId := getTriggerReferenceId(triggerCache.TriggeringEvent)

		triggerCache.TriggerSettingsOut <-
			eventTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)][triggerCache.TriggeringWorkflowVersionNodeId][triggerReferenceId]
	case writeEventTrigger:
		if triggerCache.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", triggerCache.WorkflowVersionId)
			return scheduledTriggerCache
		}
		initializeEventTriggerCache(eventTriggerCache, triggerCache.WorkflowVersionId, triggerCache.TriggeringWorkflowVersionNodeId)

		triggerReferenceId := getTriggerReferenceId(triggerCache.TriggeringEvent)

		triggerSettings, exists := eventTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)][triggerCache.TriggeringWorkflowVersionNodeId][triggerReferenceId]
		if !exists {
			eventTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)][triggerCache.TriggeringWorkflowVersionNodeId][triggerReferenceId] = TriggerSettingsCache{
				WorkflowVersionId:               triggerCache.WorkflowVersionId,
				TriggeringWorkflowVersionNodeId: triggerCache.TriggeringWorkflowVersionNodeId,
				TriggeringNodeType:              triggerCache.TriggeringNodeType,
				TriggeringEventQueue:            []any{triggerCache.TriggeringEvent},
				Reference:                       triggerCache.Reference,
				Status:                          triggerCache.Status,
				CancelFunction:                  triggerCache.CancelFunction,
				BootTime:                        triggerCache.BootTime,
			}
			return scheduledTriggerCache
		}

		if triggerSettings.Status != triggerCache.Status {
			triggerSettings.PreviousState = triggerSettings.Status
		}
		if triggerCache.BootTime != nil {
			triggerSettings.BootTime = triggerCache.BootTime
		}
		triggerSettings.TriggeringWorkflowVersionNodeId = triggerCache.TriggeringWorkflowVersionNodeId
		triggerSettings.TriggeringNodeType = triggerCache.TriggeringNodeType
		triggerSettings.TriggeringEventQueue = append(triggerSettings.TriggeringEventQueue, triggerCache.TriggeringEvent)
		triggerSettings.Reference = triggerCache.Reference
		triggerSettings.Status = triggerCache.Status
		triggerSettings.CancelFunction = triggerCache.CancelFunction
		eventTriggerCache[workflowVersionId(triggerCache.WorkflowVersionId)][triggerCache.TriggeringWorkflowVersionNodeId][triggerReferenceId] = triggerSettings

	case popScheduledTrigger:
		if len(scheduledTriggerCache) > 0 {
			sort.Slice(scheduledTriggerCache, func(i, j int) bool {
				return (*scheduledTriggerCache[i].SchedulingTime).Before(*scheduledTriggerCache[j].SchedulingTime)
			})
			next := scheduledTriggerCache[0]
			scheduledTriggerCache = scheduledTriggerCache[1:]
			triggerCache.TriggerSettingsOut <- next
			return scheduledTriggerCache
		}

		triggerCache.TriggerSettingsOut <- TriggerSettingsCache{}
	case writeScheduledTrigger:
		if triggerCache.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", triggerCache.WorkflowVersionId)
			return scheduledTriggerCache
		}

		triggerReferenceId := getTriggerReferenceId(triggerCache.TriggeringEvent)

		for _, scheduledItem := range scheduledTriggerCache {
			if scheduledItem.WorkflowVersionId == triggerCache.WorkflowVersionId &&
				scheduledItem.TriggeringNodeType == triggerCache.TriggeringNodeType &&
				getTriggerReferenceId(scheduledItem.TriggeringEventQueue[0]) == triggerReferenceId {
				scheduledItem.TriggeringEventQueue = append(scheduledItem.TriggeringEventQueue, triggerCache.TriggeringEvent)

				//if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
				// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
				//}

				log.Debug().Msgf("Trigger got scheduled while there is a pending version with triggerReferenceId: %v, events: %v, queue: %v",
					triggerReferenceId, len(scheduledItem.TriggeringEventQueue), len(scheduledTriggerCache))

				return scheduledTriggerCache
			}
		}
		now := time.Now()
		scheduledTriggerCache = append(scheduledTriggerCache, TriggerSettingsCache{
			SchedulingTime:                  &now,
			WorkflowVersionId:               triggerCache.WorkflowVersionId,
			TriggeringNodeType:              triggerCache.TriggeringNodeType,
			TriggeringWorkflowVersionNodeId: triggerCache.TriggeringWorkflowVersionNodeId,
			TriggeringEventQueue:            []any{triggerCache.TriggeringEvent},
			Reference:                       triggerCache.Reference,
		})
		log.Debug().Msgf("Amount of triggers currently scheduled: %v", len(scheduledTriggerCache))
	}
	return scheduledTriggerCache
}

func getTriggerReferenceId(triggeringEvent any) int {
	var triggerReferenceId int
	switch event := triggeringEvent.(type) {
	case core.ChannelBalanceEvent:
		triggerReferenceId = event.ChannelId
	case core.ChannelEvent:
		triggerReferenceId = event.ChannelId
	default:
		triggerReferenceId = 0
	}
	return triggerReferenceId
}

func initializeEventTriggerCache(
	eventTriggerCache map[workflowVersionId]map[int]map[int]TriggerSettingsCache,
	wfVersionId int,
	triggeringWorkflowVersionNodeId int) {

	if eventTriggerCache[workflowVersionId(wfVersionId)] == nil {
		eventTriggerCache[workflowVersionId(wfVersionId)] = make(map[int]map[int]TriggerSettingsCache)
	}
	if eventTriggerCache[workflowVersionId(wfVersionId)][triggeringWorkflowVersionNodeId] == nil {
		eventTriggerCache[workflowVersionId(wfVersionId)][triggeringWorkflowVersionNodeId] = make(map[int]TriggerSettingsCache)
	}
}

func GetTimeTriggerSettingsByWorkflowVersionId(workflowVersionId int) TriggerSettingsCache {
	triggerSettingsChannel := make(chan TriggerSettingsCache)
	triggerCache := TriggerCache{
		WorkflowVersionId:  workflowVersionId,
		Type:               readTimeTriggerSettings,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	TriggersCacheChannel <- triggerCache
	return <-triggerSettingsChannel
}

func GetEventTriggerSettingsByWorkflowVersionId(
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType core.WorkflowNodeType,
	triggeringEvent any) TriggerSettingsCache {

	triggerSettingsChannel := make(chan TriggerSettingsCache)
	triggerCache := TriggerCache{
		WorkflowVersionId:               workflowVersionId,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringEvent:                 triggeringEvent,
		Type:                            readEventTriggerSettings,
		TriggerSettingsOut:              triggerSettingsChannel,
	}
	TriggersCacheChannel <- triggerCache
	return <-triggerSettingsChannel
}

func ActivateWorkflowTrigger(
	reference string,
	workflowVersionId int,
	cancel context.CancelFunc) {

	now := time.Now()
	bootTime := &now
	TriggersCacheChannel <- TriggerCache{
		WorkflowVersionId: workflowVersionId,
		Status:            core.Active,
		BootTime:          bootTime,
		Reference:         reference,
		CancelFunction:    cancel,
		Type:              writeTimeTrigger,
	}
}

func DeactivateWorkflowTrigger(workflowVersionId int) {
	TriggersCacheChannel <- TriggerCache{
		WorkflowVersionId: workflowVersionId,
		Status:            core.Inactive,
		Type:              writeTimeTrigger,
	}
}

func ActivateEventTrigger(
	reference string,
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType core.WorkflowNodeType,
	triggeringEvent any,
	cancel context.CancelFunc) {

	now := time.Now()
	bootTime := &now
	TriggersCacheChannel <- TriggerCache{
		WorkflowVersionId:               workflowVersionId,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringEvent:                 triggeringEvent,
		Status:                          core.Active,
		BootTime:                        bootTime,
		Reference:                       reference,
		CancelFunction:                  cancel,
		Type:                            writeEventTrigger,
	}
}

func DeactivateEventTrigger(
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType core.WorkflowNodeType,
	triggeringEvent any) {

	TriggersCacheChannel <- TriggerCache{
		WorkflowVersionId:               workflowVersionId,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringEvent:                 triggeringEvent,
		Status:                          core.Inactive,
		Type:                            writeEventTrigger,
	}
}

func ScheduleTrigger(
	reference string,
	workflowVersionId int,
	triggeringNodeType core.WorkflowNodeType,
	triggeringWorkflowVersionNodeId int,
	triggeringEvent any) {

	TriggersCacheChannel <- TriggerCache{
		Reference:                       reference,
		WorkflowVersionId:               workflowVersionId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringEvent:                 triggeringEvent,
		Type:                            writeScheduledTrigger,
	}
}

func GetScheduledTrigger() TriggerSettingsCache {
	triggerSettingsChannel := make(chan TriggerSettingsCache)
	TriggersCacheChannel <- TriggerCache{
		Type:               popScheduledTrigger,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	return <-triggerSettingsChannel
}
