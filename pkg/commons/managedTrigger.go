package commons

import (
	"context"
	"time"
)

var ManagedTriggerChannel = make(chan ManagedTrigger) //nolint:gochecknoglobals

type ManagedTriggerCacheOperationType uint

const (
	READ_TRIGGER_SETTINGS ManagedTriggerCacheOperationType = iota
	WRITE_TRIGGER
	WRITE_TRIGGER_VERIFICATIONTIME
	WRITE_TRIGGER_BOOTED
)

type ManagedTrigger struct {
	Type                           ManagedTriggerCacheOperationType
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
	TriggeredWorkflowVersionNodeId int
	Reference                      string
	CancelFunction                 context.CancelFunc
	PreviousState                  string
	BootTime                       *time.Time
	VerificationTime               *time.Time
	Status                         Status
}

type ManagedTriggerChannelBalanceBounds struct {
	ChannelId  int
	UpperBound int64
	Balance    int64
	LowerBound int64
}

func ManagedTriggerCache(ch chan ManagedTrigger, ctx context.Context) {
	triggerCache := make(map[int]ManagedTriggerSettings, 0)
	for {
		select {
		case <-ctx.Done():
			return
		case managedTrigger := <-ch:
			processManagedTrigger(managedTrigger, triggerCache)
		}
	}
}

func processManagedTrigger(managedTrigger ManagedTrigger, triggerCache map[int]ManagedTriggerSettings) {
	switch managedTrigger.Type {
	case READ_TRIGGER_SETTINGS:
		SendToManagedTriggerSettingsChannel(managedTrigger.TriggerSettingsOut, triggerCache[managedTrigger.WorkflowVersionId])
	case WRITE_TRIGGER:
		triggerCache[managedTrigger.WorkflowVersionId] = ManagedTriggerSettings{
			WorkflowVersionId:              managedTrigger.WorkflowVersionId,
			TriggeredWorkflowVersionNodeId: managedTrigger.TriggeredWorkflowVersionNodeId,
			Reference:                      managedTrigger.Reference,
			Status:                         managedTrigger.Status,
			CancelFunction:                 managedTrigger.CancelFunction,
			BootTime:                       managedTrigger.BootTime,
			PreviousState:                  managedTrigger.PreviousState,
		}
	case WRITE_TRIGGER_VERIFICATIONTIME:
		triggerSettings := triggerCache[managedTrigger.WorkflowVersionId]
		triggerSettings.VerificationTime = managedTrigger.VerificationTime
		triggerCache[managedTrigger.WorkflowVersionId] = triggerSettings
	}
}

func GetTriggerSettingsByWorkflowVersionId(workflowVersionId int) ManagedTriggerSettings {
	triggerSettingsChannel := make(chan ManagedTriggerSettings)
	managedTrigger := ManagedTrigger{
		WorkflowVersionId:  workflowVersionId,
		Type:               READ_TRIGGER_SETTINGS,
		TriggerSettingsOut: triggerSettingsChannel,
	}
	ManagedTriggerChannel <- managedTrigger
	return <-triggerSettingsChannel
}

func SetTriggerVerificationTime(workflowVersionId int, verificationTime time.Time) {
	ManagedTriggerChannel <- ManagedTrigger{
		WorkflowVersionId: workflowVersionId,
		VerificationTime:  &verificationTime,
		Type:              WRITE_TRIGGER_VERIFICATIONTIME,
	}
}

func SetTrigger(reference string, workflowVersionId int, triggeredWorkflowVersionNodeId int, status Status, cancel context.CancelFunc) {
	var bootTime *time.Time
	if status == Active {
		now := time.Now()
		bootTime = &now
	}
	ManagedTriggerChannel <- ManagedTrigger{
		WorkflowVersionId:              workflowVersionId,
		TriggeredWorkflowVersionNodeId: triggeredWorkflowVersionNodeId,
		Status:                         status,
		BootTime:                       bootTime,
		Reference:                      reference,
		CancelFunction:                 cancel,
		Type:                           WRITE_TRIGGER,
	}
}
