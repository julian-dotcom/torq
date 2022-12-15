package commons

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

var ManagedTriggerChannel = make(chan ManagedTrigger) //nolint:gochecknoglobals

type ManagedTriggerCacheOperationType uint

const (
	READ_TRIGGER_SETTINGS ManagedTriggerCacheOperationType = iota
	WRITE_TRIGGER
	WRITE_TRIGGER_VERIFICATIONTIME
	WRITE_TRIGGER_CHANNEL_BALANCE_BOUNDS
	INVALIDATE_TRIGGER_CHANNEL_BALANCE_BOUNDS
)

type ManagedTrigger struct {
	Type                           ManagedTriggerCacheOperationType
	NodeId                         int
	WorkflowVersionId              int
	TriggeredWorkflowVersionNodeId int
	Reference                      string
	CancelFunction                 context.CancelFunc
	PreviousState                  string
	BootTime                       *time.Time
	VerificationTime               *time.Time
	Status                         Status
	ChannelBalanceBounds           ManagedTriggerChannelBalanceBounds
	Out                            chan ManagedTrigger
	TriggerSettingsOut             chan ManagedTriggerSettings
}

type ManagedTriggerSettings struct {
	WorkflowVersionId              int
	NodeId                         int
	TriggeredWorkflowVersionNodeId int
	Reference                      string
	CancelFunction                 context.CancelFunc
	PreviousState                  string
	BootTime                       *time.Time
	VerificationTime               *time.Time
	Status                         Status
}

func ManagedTriggerCache(ch chan ManagedTrigger, ctx context.Context) {
	triggerCache := make(map[int]map[int]ManagedTriggerSettings, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedTrigger := <-ch:
			processManagedTrigger(managedTrigger, triggerCache, channelBalanceBoundsCache)
		}
	}
}

func processManagedTrigger(managedTrigger ManagedTrigger,
	triggerCache map[int]map[int]ManagedTriggerSettings,
	channelBalanceBoundsCache map[int]map[int]map[int]ManagedTriggerChannelBalanceBounds) {

	switch managedTrigger.Type {
	case READ_TRIGGER_SETTINGS:
		if managedTrigger.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}
		initializeTriggerCache(triggerCache, channelBalanceBoundsCache, managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, triggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId])
	case WRITE_TRIGGER:
		if managedTrigger.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}
		initializeTriggerCache(triggerCache, channelBalanceBoundsCache, managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
		triggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId] = ManagedTriggerSettings{
			WorkflowVersionId:              managedTrigger.WorkflowVersionId,
			TriggeredWorkflowVersionNodeId: managedTrigger.TriggeredWorkflowVersionNodeId,
			Reference:                      managedTrigger.Reference,
			Status:                         managedTrigger.Status,
			CancelFunction:                 managedTrigger.CancelFunction,
			BootTime:                       managedTrigger.BootTime,
			PreviousState:                  managedTrigger.PreviousState,
		}
	case WRITE_TRIGGER_VERIFICATIONTIME:
		if managedTrigger.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}
		initializeTriggerCache(triggerCache, channelBalanceBoundsCache, managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
		triggerSettings := triggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId]
		triggerSettings.VerificationTime = managedTrigger.VerificationTime
		triggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId] = triggerSettings
	case WRITE_TRIGGER_CHANNEL_BALANCE_BOUNDS:
		if managedTrigger.NodeId == 0 || managedTrigger.WorkflowVersionId == 0 || managedTrigger.ChannelBalanceBounds.ChannelId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) or ChannelId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId, managedTrigger.ChannelBalanceBounds.ChannelId)
			return
		}
		initializeTriggerCache(triggerCache, channelBalanceBoundsCache, managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
		existingBounds := channelBalanceBoundsCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId][managedTrigger.ChannelBalanceBounds.ChannelId]
		if existingBounds.ChannelId == 0 {
			channelBalanceBoundsCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId][managedTrigger.ChannelBalanceBounds.ChannelId] = managedTrigger.ChannelBalanceBounds
		} else {
			if existingBounds.LowerBound < managedTrigger.ChannelBalanceBounds.LowerBound {
				existingBounds.LowerBound = managedTrigger.ChannelBalanceBounds.LowerBound
			}
			if existingBounds.UpperBound > managedTrigger.ChannelBalanceBounds.UpperBound {
				existingBounds.UpperBound = managedTrigger.ChannelBalanceBounds.UpperBound
			}
			if existingBounds.Balance != managedTrigger.ChannelBalanceBounds.Balance {
				existingBounds.Balance = managedTrigger.ChannelBalanceBounds.Balance
			}
			channelBalanceBoundsCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId][managedTrigger.ChannelBalanceBounds.ChannelId] = existingBounds
		}
	case INVALIDATE_TRIGGER_CHANNEL_BALANCE_BOUNDS:
		if managedTrigger.NodeId == 0 || managedTrigger.WorkflowVersionId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId, managedTrigger.ChannelBalanceBounds.ChannelId)
			return
		}
		initializeTriggerCache(triggerCache, channelBalanceBoundsCache, managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
		channelBalanceBoundsCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId] = make(map[int]ManagedTriggerChannelBalanceBounds, 0)
	}
}

func initializeTriggerCache(triggerCache map[int]map[int]ManagedTriggerSettings,
	channelBalanceBoundsCache map[int]map[int]map[int]ManagedTriggerChannelBalanceBounds,
	nodeId int,
	workflowVersionId int) {

	if triggerCache[nodeId] == nil {
		triggerCache[nodeId] = make(map[int]ManagedTriggerSettings, 0)
	}
	if channelBalanceBoundsCache[nodeId] == nil {
		channelBalanceBoundsCache[nodeId] = make(map[int]map[int]ManagedTriggerChannelBalanceBounds, 0)
	}
	if workflowVersionId != 0 {
		channelBalanceBoundsCache[nodeId][workflowVersionId] = make(map[int]ManagedTriggerChannelBalanceBounds, 0)
	}
}

func GetTriggerSettingsByWorkflowVersionId(nodeId int, workflowVersionId int) ManagedTriggerSettings {
	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	managedTrigger := ManagedTrigger{
		NodeId:             nodeId,
		WorkflowVersionId:  workflowVersionId,
		Type:               READ_TRIGGER_SETTINGS,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	ManagedTriggerChannel <- managedTrigger
	return <-triggerSettingsChannel
}

func SetTriggerVerificationTime(nodeId int, workflowVersionId int, verificationTime time.Time) {
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:            nodeId,
		WorkflowVersionId: workflowVersionId,
		VerificationTime:  &verificationTime,
		Type:              WRITE_TRIGGER_VERIFICATIONTIME,
	}
}

func SetTrigger(nodeId int, reference string, workflowVersionId int, triggeredWorkflowVersionNodeId int, status Status, cancel context.CancelFunc) {
	var bootTime *time.Time
	if status == Active {
		now := time.Now()
		bootTime = &now
	}
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:                         nodeId,
		WorkflowVersionId:              workflowVersionId,
		TriggeredWorkflowVersionNodeId: triggeredWorkflowVersionNodeId,
		Status:                         status,
		BootTime:                       bootTime,
		Reference:                      reference,
		CancelFunction:                 cancel,
		Type:                           WRITE_TRIGGER,
	}
}

func SetTriggerChannelBalanceBound(nodeId int, workflowVersionId int, channelBalanceBounds ManagedTriggerChannelBalanceBounds) {
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:               nodeId,
		WorkflowVersionId:    workflowVersionId,
		ChannelBalanceBounds: channelBalanceBounds,
		Type:                 WRITE_TRIGGER_CHANNEL_BALANCE_BOUNDS,
	}
}

// ClearTriggerChannelBalanceBounds Use this to clear out previous versions of a workflow.
// That way they don't keep using system resources.
func ClearTriggerChannelBalanceBounds(nodeId int, workflowVersionId int) {
	ManagedTriggerChannel <- ManagedTrigger{
		NodeId:            nodeId,
		WorkflowVersionId: workflowVersionId,
		Type:              INVALIDATE_TRIGGER_CHANNEL_BALANCE_BOUNDS,
	}
}
