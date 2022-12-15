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
			processManagedTrigger(managedTrigger, triggerCache)
		}
	}
}

func processManagedTrigger(managedTrigger ManagedTrigger,
	triggerCache map[int]map[int]ManagedTriggerSettings) {

	switch managedTrigger.Type {
	case READ_TRIGGER_SETTINGS:
		if managedTrigger.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}
		initializeTriggerCache(triggerCache, managedTrigger.NodeId)
		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, triggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId])
	case WRITE_TRIGGER:
		if managedTrigger.NodeId == 0 {
			log.Error().Msgf("No empty NodeId (%v) or WorkflowVersionId (%v) allowed",
				managedTrigger.NodeId, managedTrigger.WorkflowVersionId)
			SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, ManagedTriggerSettings{})
			return
		}
		initializeTriggerCache(triggerCache, managedTrigger.NodeId)
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
		initializeTriggerCache(triggerCache, managedTrigger.NodeId)
		triggerSettings := triggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId]
		triggerSettings.VerificationTime = managedTrigger.VerificationTime
		triggerCache[managedTrigger.NodeId][managedTrigger.WorkflowVersionId] = triggerSettings
	}
}

func initializeTriggerCache(triggerCache map[int]map[int]ManagedTriggerSettings, nodeId int) {
	if triggerCache[nodeId] == nil {
		triggerCache[nodeId] = make(map[int]ManagedTriggerSettings, 0)
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
