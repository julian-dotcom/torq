package commons

import (
	"context"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

var ManagedTriggerChannel = make(chan ManagedTrigger) //nolint:gochecknoglobals

type ManagedTriggerCacheOperationType uint

const (
	READ_TIME_TRIGGER_SETTINGS ManagedTriggerCacheOperationType = iota
	WRITE_TIME_TRIGGER
	READ_EVENT_TRIGGER_SETTINGS
	WRITE_EVENT_TRIGGER
	POP_SCHEDULED_TRIGGER
	WRITE_SCHEDULED_TRIGGER
)

type ManagedTrigger struct {
	Type                            ManagedTriggerCacheOperationType
	WorkflowVersionId               int
	TriggeringWorkflowVersionNodeId int
	TriggeringNodeType              WorkflowNodeType
	TriggeringEvent                 any
	Reference                       string
	CancelFunction                  context.CancelFunc
	BootTime                        *time.Time
	VerificationTime                *time.Time
	Status                          Status
	PreviousState                   Status
	TriggerSettingsOut              chan<- ManagedTriggerSettings
}

type ManagedTriggerSettings struct {
	WorkflowVersionId               int
	TriggeringWorkflowVersionNodeId int
	TriggeringNodeType              WorkflowNodeType
	Reference                       string
	CancelFunction                  context.CancelFunc
	BootTime                        *time.Time
	VerificationTime                *time.Time
	SchedulingTime                  *time.Time
	TriggeringEventQueue            []any
	Status                          Status
	PreviousState                   Status
}

func ManagedTriggerCache(ch <-chan ManagedTrigger, ctx context.Context) {
	timeTriggerCache := make(map[int]ManagedTriggerSettings)
	eventTriggerCache := make(map[int]map[int]map[int]ManagedTriggerSettings)
	var scheduledTriggerCache []ManagedTriggerSettings
	for {
		select {
		case <-ctx.Done():
			return
		case managedTrigger := <-ch:
			scheduledTriggerCache = processManagedTrigger(managedTrigger, timeTriggerCache, eventTriggerCache, scheduledTriggerCache)
		}
	}
}

func processManagedTrigger(managedTrigger ManagedTrigger,
	timeTriggerCache map[int]ManagedTriggerSettings,
	eventTriggerCache map[int]map[int]map[int]ManagedTriggerSettings,
	scheduledTriggerCache []ManagedTriggerSettings) []ManagedTriggerSettings {

	switch managedTrigger.Type {
	case READ_TIME_TRIGGER_SETTINGS:
		if managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", managedTrigger.WorkflowVersionId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return scheduledTriggerCache
		}
		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut,
			timeTriggerCache[managedTrigger.WorkflowVersionId])
	case WRITE_TIME_TRIGGER:
		if managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", managedTrigger.WorkflowVersionId)
			return scheduledTriggerCache
		}
		timeTriggerSettings, exists := timeTriggerCache[managedTrigger.WorkflowVersionId]
		if !exists {
			timeTriggerCache[managedTrigger.WorkflowVersionId] = ManagedTriggerSettings{
				WorkflowVersionId:               managedTrigger.WorkflowVersionId,
				TriggeringWorkflowVersionNodeId: managedTrigger.TriggeringWorkflowVersionNodeId,
				TriggeringNodeType:              managedTrigger.TriggeringNodeType,
				TriggeringEventQueue:            []any{managedTrigger.TriggeringEvent},
				Reference:                       managedTrigger.Reference,
				Status:                          managedTrigger.Status,
				CancelFunction:                  managedTrigger.CancelFunction,
				BootTime:                        managedTrigger.BootTime,
			}
			return scheduledTriggerCache
		}

		if timeTriggerSettings.Status != managedTrigger.Status {
			timeTriggerSettings.PreviousState = timeTriggerSettings.Status
		}
		if managedTrigger.BootTime != nil {
			timeTriggerSettings.BootTime = managedTrigger.BootTime
		}
		timeTriggerSettings.TriggeringWorkflowVersionNodeId = managedTrigger.TriggeringWorkflowVersionNodeId
		timeTriggerSettings.TriggeringNodeType = managedTrigger.TriggeringNodeType
		timeTriggerSettings.TriggeringEventQueue = append(timeTriggerSettings.TriggeringEventQueue, managedTrigger.TriggeringEvent)
		timeTriggerSettings.Reference = managedTrigger.Reference
		timeTriggerSettings.Status = managedTrigger.Status
		timeTriggerSettings.CancelFunction = managedTrigger.CancelFunction
		timeTriggerCache[managedTrigger.WorkflowVersionId] = timeTriggerSettings

	case READ_EVENT_TRIGGER_SETTINGS:
		if managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", managedTrigger.WorkflowVersionId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return scheduledTriggerCache
		}

		initializeEventTriggerCache(eventTriggerCache, managedTrigger.WorkflowVersionId, managedTrigger.TriggeringWorkflowVersionNodeId)

		triggerReferenceId := getTriggerReferenceId(managedTrigger.TriggeringEvent)

		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut,
			eventTriggerCache[managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId])
	case WRITE_EVENT_TRIGGER:
		if managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", managedTrigger.WorkflowVersionId)
			return scheduledTriggerCache
		}
		initializeEventTriggerCache(eventTriggerCache, managedTrigger.WorkflowVersionId, managedTrigger.TriggeringWorkflowVersionNodeId)

		triggerReferenceId := getTriggerReferenceId(managedTrigger.TriggeringEvent)

		triggerSettings, exists := eventTriggerCache[managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId]
		if !exists {
			eventTriggerCache[managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId] = ManagedTriggerSettings{
				WorkflowVersionId:               managedTrigger.WorkflowVersionId,
				TriggeringWorkflowVersionNodeId: managedTrigger.TriggeringWorkflowVersionNodeId,
				TriggeringNodeType:              managedTrigger.TriggeringNodeType,
				TriggeringEventQueue:            []any{managedTrigger.TriggeringEvent},
				Reference:                       managedTrigger.Reference,
				Status:                          managedTrigger.Status,
				CancelFunction:                  managedTrigger.CancelFunction,
				BootTime:                        managedTrigger.BootTime,
			}
			return scheduledTriggerCache
		}

		if triggerSettings.Status != managedTrigger.Status {
			triggerSettings.PreviousState = triggerSettings.Status
		}
		if managedTrigger.BootTime != nil {
			triggerSettings.BootTime = managedTrigger.BootTime
		}
		triggerSettings.TriggeringWorkflowVersionNodeId = managedTrigger.TriggeringWorkflowVersionNodeId
		triggerSettings.TriggeringNodeType = managedTrigger.TriggeringNodeType
		triggerSettings.TriggeringEventQueue = append(triggerSettings.TriggeringEventQueue, managedTrigger.TriggeringEvent)
		triggerSettings.Reference = managedTrigger.Reference
		triggerSettings.Status = managedTrigger.Status
		triggerSettings.CancelFunction = managedTrigger.CancelFunction
		eventTriggerCache[managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId] = triggerSettings

	case POP_SCHEDULED_TRIGGER:
		if len(scheduledTriggerCache) > 0 {
			sort.Slice(scheduledTriggerCache, func(i, j int) bool {
				return (*scheduledTriggerCache[i].SchedulingTime).Before(*scheduledTriggerCache[j].SchedulingTime)
			})
			next := scheduledTriggerCache[0]
			scheduledTriggerCache = scheduledTriggerCache[1:]
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, next)
			return scheduledTriggerCache
		}

		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
	case WRITE_SCHEDULED_TRIGGER:
		if managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty WorkflowVersionId (%v) allowed", managedTrigger.WorkflowVersionId)
			return scheduledTriggerCache
		}

		triggerReferenceId := getTriggerReferenceId(managedTrigger.TriggeringEvent)

		for _, scheduledItem := range scheduledTriggerCache {
			if scheduledItem.WorkflowVersionId == managedTrigger.WorkflowVersionId &&
				scheduledItem.TriggeringNodeType == managedTrigger.TriggeringNodeType &&
				getTriggerReferenceId(scheduledItem.TriggeringEventQueue[0]) == triggerReferenceId {
				scheduledItem.TriggeringEventQueue = append(scheduledItem.TriggeringEventQueue, managedTrigger.TriggeringEvent)

				//if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
				// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
				//}

				return scheduledTriggerCache
			}
		}
		now := time.Now()
		scheduledTriggerCache = append(scheduledTriggerCache, ManagedTriggerSettings{
			SchedulingTime:                  &now,
			WorkflowVersionId:               managedTrigger.WorkflowVersionId,
			TriggeringNodeType:              managedTrigger.TriggeringNodeType,
			TriggeringWorkflowVersionNodeId: managedTrigger.TriggeringWorkflowVersionNodeId,
			TriggeringEventQueue:            []any{managedTrigger.TriggeringEvent},
			Reference:                       managedTrigger.Reference,
		})
		log.Debug().Msgf("Amount of triggers currently scheduled: %v", len(scheduledTriggerCache))
	}
	return scheduledTriggerCache
}

func getTriggerReferenceId(triggeringEvent any) int {
	var triggerReferenceId int
	switch event := triggeringEvent.(type) {
	case ChannelBalanceEvent:
		triggerReferenceId = event.ChannelId
	case ChannelEvent:
		triggerReferenceId = event.ChannelId
	default:
		triggerReferenceId = 0
	}
	return triggerReferenceId
}

func initializeEventTriggerCache(
	eventTriggerCache map[int]map[int]map[int]ManagedTriggerSettings,
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int) {

	if eventTriggerCache[workflowVersionId] == nil {
		eventTriggerCache[workflowVersionId] = make(map[int]map[int]ManagedTriggerSettings)
	}
	if eventTriggerCache[workflowVersionId][triggeringWorkflowVersionNodeId] == nil {
		eventTriggerCache[workflowVersionId][triggeringWorkflowVersionNodeId] = make(map[int]ManagedTriggerSettings)
	}
}

func GetTimeTriggerSettingsByWorkflowVersionId(workflowVersionId int) ManagedTriggerSettings {
	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	managedTrigger := ManagedTrigger{
		WorkflowVersionId:  workflowVersionId,
		Type:               READ_TIME_TRIGGER_SETTINGS,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	ManagedTriggerChannel <- managedTrigger
	return <-triggerSettingsChannel
}

func GetEventTriggerSettingsByWorkflowVersionId(
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType WorkflowNodeType,
	triggeringEvent any) ManagedTriggerSettings {

	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	managedTrigger := ManagedTrigger{
		WorkflowVersionId:               workflowVersionId,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringEvent:                 triggeringEvent,
		Type:                            READ_EVENT_TRIGGER_SETTINGS,
		TriggerSettingsOut:              triggerSettingsChannel,
	}
	ManagedTriggerChannel <- managedTrigger
	return <-triggerSettingsChannel
}

func ActivateWorkflowTrigger(
	reference string,
	workflowVersionId int,
	cancel context.CancelFunc) {

	now := time.Now()
	bootTime := &now
	ManagedTriggerChannel <- ManagedTrigger{
		WorkflowVersionId: workflowVersionId,
		Status:            Active,
		BootTime:          bootTime,
		Reference:         reference,
		CancelFunction:    cancel,
		Type:              WRITE_TIME_TRIGGER,
	}
}

func DeactivateWorkflowTrigger(workflowVersionId int) {
	ManagedTriggerChannel <- ManagedTrigger{
		WorkflowVersionId: workflowVersionId,
		Status:            Inactive,
		Type:              WRITE_TIME_TRIGGER,
	}
}

func ActivateEventTrigger(
	reference string,
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType WorkflowNodeType,
	triggeringEvent any,
	cancel context.CancelFunc) {

	now := time.Now()
	bootTime := &now
	ManagedTriggerChannel <- ManagedTrigger{
		WorkflowVersionId:               workflowVersionId,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringEvent:                 triggeringEvent,
		Status:                          Active,
		BootTime:                        bootTime,
		Reference:                       reference,
		CancelFunction:                  cancel,
		Type:                            WRITE_EVENT_TRIGGER,
	}
}

func DeactivateEventTrigger(
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType WorkflowNodeType,
	triggeringEvent any) {

	ManagedTriggerChannel <- ManagedTrigger{
		WorkflowVersionId:               workflowVersionId,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringEvent:                 triggeringEvent,
		Status:                          Inactive,
		Type:                            WRITE_EVENT_TRIGGER,
	}
}

func ScheduleTrigger(
	reference string,
	workflowVersionId int,
	triggeringNodeType WorkflowNodeType,
	triggeringWorkflowVersionNodeId int,
	triggeringEvent any) {

	ManagedTriggerChannel <- ManagedTrigger{
		Reference:                       reference,
		WorkflowVersionId:               workflowVersionId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringEvent:                 triggeringEvent,
		Type:                            WRITE_SCHEDULED_TRIGGER,
	}
}

func GetScheduledTrigger() ManagedTriggerSettings {
	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	ManagedTriggerChannel <- ManagedTrigger{
		Type:               POP_SCHEDULED_TRIGGER,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	return <-triggerSettingsChannel
}
