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
	NodeId                          int
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
	TriggerSettingsOut              chan ManagedTriggerSettings
}

type ManagedTriggerSettings struct {
	WorkflowVersionId               int
	NodeId                          int
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

func ManagedTriggerCache(ch chan ManagedTrigger, ctx context.Context) {
	timeTriggerCache := make(map[int]map[int]ManagedTriggerSettings)
	eventTriggerCache := make(map[int]map[int]map[int]map[int]ManagedTriggerSettings)
	scheduledTriggerCache := make(map[int][]ManagedTriggerSettings)
	for {
		select {
		case <-ctx.Done():
			return
		case managedTrigger := <-ch:
			processManagedTrigger(managedTrigger, timeTriggerCache, eventTriggerCache, scheduledTriggerCache)
		}
	}
}

func processManagedTrigger(managedTrigger ManagedTrigger,
	timeTriggerCache map[int]map[int]ManagedTriggerSettings,
	eventTriggerCache map[int]map[int]map[int]map[int]ManagedTriggerSettings,
	scheduledTriggerCache map[int][]ManagedTriggerSettings) {

	switch managedTrigger.Type {
	case READ_TIME_TRIGGER_SETTINGS:
		if managedTrigger.NodeId == 0 || managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId, managedTrigger.TriggeringWorkflowVersionNodeId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}
		initializeTimeTriggerCache(timeTriggerCache, managedTrigger.NodeId)
		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut,
			timeTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId])
	case WRITE_TIME_TRIGGER:
		if managedTrigger.NodeId == 0 || managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			return
		}
		initializeTimeTriggerCache(timeTriggerCache, managedTrigger.NodeId)

		timeTriggerSettings, exists := timeTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId]
		if !exists {
			timeTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId] = ManagedTriggerSettings{
				NodeId:                          managedTrigger.NodeId,
				WorkflowVersionId:               managedTrigger.WorkflowVersionId,
				TriggeringWorkflowVersionNodeId: managedTrigger.TriggeringWorkflowVersionNodeId,
				TriggeringNodeType:              managedTrigger.TriggeringNodeType,
				TriggeringEventQueue:            []any{managedTrigger.TriggeringEvent},
				Reference:                       managedTrigger.Reference,
				Status:                          managedTrigger.Status,
				CancelFunction:                  managedTrigger.CancelFunction,
				BootTime:                        managedTrigger.BootTime,
			}
			return
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
		timeTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId] = timeTriggerSettings

	case READ_EVENT_TRIGGER_SETTINGS:
		if managedTrigger.NodeId == 0 || managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId, managedTrigger.TriggeringWorkflowVersionNodeId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}

		initializeEventTriggerCache(eventTriggerCache, managedTrigger.NodeId, managedTrigger.WorkflowVersionId, managedTrigger.TriggeringWorkflowVersionNodeId)

		triggerReferenceId := getTriggerReferenceId(managedTrigger.TriggeringEvent)

		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut,
			eventTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId])
	case WRITE_EVENT_TRIGGER:
		if managedTrigger.NodeId == 0 || managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			return
		}
		initializeEventTriggerCache(eventTriggerCache, managedTrigger.NodeId, managedTrigger.WorkflowVersionId, managedTrigger.TriggeringWorkflowVersionNodeId)

		triggerReferenceId := getTriggerReferenceId(managedTrigger.TriggeringEvent)

		triggerSettings, exists := eventTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId]
		if !exists {
			eventTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId] = ManagedTriggerSettings{
				NodeId:                          managedTrigger.NodeId,
				WorkflowVersionId:               managedTrigger.WorkflowVersionId,
				TriggeringWorkflowVersionNodeId: managedTrigger.TriggeringWorkflowVersionNodeId,
				TriggeringNodeType:              managedTrigger.TriggeringNodeType,
				TriggeringEventQueue:            []any{managedTrigger.TriggeringEvent},
				Reference:                       managedTrigger.Reference,
				Status:                          managedTrigger.Status,
				CancelFunction:                  managedTrigger.CancelFunction,
				BootTime:                        managedTrigger.BootTime,
			}
			return
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
		eventTriggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId][managedTrigger.TriggeringWorkflowVersionNodeId][triggerReferenceId] = triggerSettings

	case POP_SCHEDULED_TRIGGER:
		if managedTrigger.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId, managedTrigger.TriggeringWorkflowVersionNodeId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}

		scheduledItems, exists := scheduledTriggerCache[managedTrigger.NodeId]
		if exists && len(scheduledItems) > 0 {
			sort.Slice(scheduledItems, func(i, j int) bool {
				return (*scheduledItems[i].SchedulingTime).Before(*scheduledItems[j].SchedulingTime)
			})
			scheduledTriggerCache[managedTrigger.NodeId] = scheduledItems[1:]
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, scheduledItems[0])
			return
		}

		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
	case WRITE_SCHEDULED_TRIGGER:
		if managedTrigger.NodeId == 0 || managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			return
		}

		triggerReferenceId := getTriggerReferenceId(managedTrigger.TriggeringEvent)

		scheduledItems, exists := scheduledTriggerCache[managedTrigger.NodeId]
		if exists {
			for _, scheduledItem := range scheduledItems {
				if scheduledItem.WorkflowVersionId == managedTrigger.WorkflowVersionId &&
					scheduledItem.TriggeringNodeType == managedTrigger.TriggeringNodeType &&
					getTriggerReferenceId(scheduledItem.TriggeringEventQueue[0]) == triggerReferenceId {
					scheduledItem.TriggeringEventQueue = append(scheduledItem.TriggeringEventQueue, managedTrigger.TriggeringEvent)

					//if commons.RunningServices[commons.LndService].GetChannelBalanceCacheStreamStatus(nodeSettings.NodeId) != commons.Active {
					// TODO FIXME CHECK HOW LONG IT'S BEEN DOWN FOR AND POTENTIALLY KILL AUTOMATIONS
					//}

					return
				}
			}
		}
		now := time.Now()
		scheduledTriggerCache[managedTrigger.NodeId] = []ManagedTriggerSettings{{
			SchedulingTime:       &now,
			NodeId:               managedTrigger.NodeId,
			WorkflowVersionId:    managedTrigger.WorkflowVersionId,
			TriggeringNodeType:   managedTrigger.TriggeringNodeType,
			TriggeringEventQueue: []any{managedTrigger.TriggeringEvent},
			Reference:            managedTrigger.Reference,
		}}
	}
}

func getTriggerReferenceId(triggeringEvent any) int {
	var triggerReferenceId int
	switch triggeringEvent.(type) {
	case ChannelBalanceEvent:
		triggerReferenceId = triggeringEvent.(ChannelBalanceEvent).ChannelId
	case ChannelEvent:
		triggerReferenceId = triggeringEvent.(ChannelEvent).ChannelId
	default:
		triggerReferenceId = 0
	}
	return triggerReferenceId
}

func initializeTimeTriggerCache(timeTriggerCache map[int]map[int]ManagedTriggerSettings, nodeId int) {
	if timeTriggerCache[nodeId] == nil {
		timeTriggerCache[nodeId] = make(map[int]ManagedTriggerSettings)
	}
}

func initializeEventTriggerCache(
	eventTriggerCache map[int]map[int]map[int]map[int]ManagedTriggerSettings,
	nodeId int,
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int) {

	if eventTriggerCache[nodeId] == nil {
		eventTriggerCache[nodeId] = make(map[int]map[int]map[int]ManagedTriggerSettings)
	}
	if eventTriggerCache[nodeId][workflowVersionId] == nil {
		eventTriggerCache[nodeId][workflowVersionId] = make(map[int]map[int]ManagedTriggerSettings)
	}
	if eventTriggerCache[nodeId][workflowVersionId][triggeringWorkflowVersionNodeId] == nil {
		eventTriggerCache[nodeId][workflowVersionId][triggeringWorkflowVersionNodeId] = make(map[int]ManagedTriggerSettings)
	}
}

func GetTimeTriggerSettingsByWorkflowVersionId(
	nodeId int,
	workflowVersionId int) ManagedTriggerSettings {

	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	managedTrigger := ManagedTrigger{
		NodeId:             nodeId,
		WorkflowVersionId:  workflowVersionId,
		Type:               READ_TIME_TRIGGER_SETTINGS,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	ManagedTriggerChannel <- managedTrigger
	return <-triggerSettingsChannel
}

func GetEventTriggerSettingsByWorkflowVersionId(
	nodeId int,
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType WorkflowNodeType,
	triggeringEvent any) ManagedTriggerSettings {

	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	managedTrigger := ManagedTrigger{
		NodeId:                          nodeId,
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

func ActivateTimeTrigger(
	nodeId int,
	reference string,
	workflowVersionId int,
	cancel context.CancelFunc) {

	now := time.Now()
	bootTime := &now
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:            nodeId,
		WorkflowVersionId: workflowVersionId,
		Status:            Active,
		BootTime:          bootTime,
		Reference:         reference,
		CancelFunction:    cancel,
		Type:              WRITE_TIME_TRIGGER,
	}
}

func DeactivateTimeTrigger(nodeId int, workflowVersionId int) {
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:            nodeId,
		WorkflowVersionId: workflowVersionId,
		Status:            Inactive,
		Type:              WRITE_TIME_TRIGGER,
	}
}

func ActivateEventTrigger(
	nodeId int,
	reference string,
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType WorkflowNodeType,
	triggeringEvent any,
	cancel context.CancelFunc) {

	now := time.Now()
	bootTime := &now
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:                          nodeId,
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
	nodeId int,
	workflowVersionId int,
	triggeringWorkflowVersionNodeId int,
	triggeringNodeType WorkflowNodeType,
	triggeringEvent any) {

	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:                          nodeId,
		WorkflowVersionId:               workflowVersionId,
		TriggeringWorkflowVersionNodeId: triggeringWorkflowVersionNodeId,
		TriggeringNodeType:              triggeringNodeType,
		TriggeringEvent:                 triggeringEvent,
		Status:                          Inactive,
		Type:                            WRITE_EVENT_TRIGGER,
	}
}

func ScheduleTrigger(
	nodeId int,
	reference string,
	workflowVersionId int,
	triggeringNodeType WorkflowNodeType,
	triggeringEvent any) {

	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:             nodeId,
		Reference:          reference,
		WorkflowVersionId:  workflowVersionId,
		TriggeringNodeType: triggeringNodeType,
		TriggeringEvent:    triggeringEvent,
		Type:               WRITE_SCHEDULED_TRIGGER,
	}
}

func GetScheduledTrigger(nodeId int) ManagedTriggerSettings {
	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:             nodeId,
		Type:               POP_SCHEDULED_TRIGGER,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	return <-triggerSettingsChannel
}
