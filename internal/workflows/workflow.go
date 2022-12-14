package workflows

import (
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type Workflow struct {
	WorkflowId int                  `json:"workflowId" db:"workflow_id"`
	Name       string               `json:"name" db:"name"`
	Type       commons.WorkflowType `json:"type" db:"type"`
	Status     commons.Status       `json:"status" db:"status"`
	CreatedOn  time.Time            `json:"createdOn" db:"created_on"`
	UpdateOn   time.Time            `json:"updatedOn" db:"updated_on"`
}

type WorkflowVersion struct {
	WorkflowVersionId int            `json:"workflowVersionId" db:"workflow_version_id"`
	Version           int            `json:"version" db:"version"`
	Status            commons.Status `json:"status" db:"status"`
	WorkflowId        int            `json:"workflowId" db:"workflow_id"`
	CreatedOn         time.Time      `json:"createdOn" db:"created_on"`
	UpdateOn          time.Time      `json:"updatedOn" db:"updated_on"`
}

type WorkflowVersionNode struct {
	WorkflowVersionNodeId int                      `json:"workflowVersionNodeId" db:"workflow_version_node_id"`
	Name                  string                   `json:"name" db:"name"`
	Status                commons.Status           `json:"status" db:"status"`
	Type                  commons.WorkflowNodeType `json:"type" db:"type"`
	Parameters            string                   `json:"parameters" db:"parameters"`
	VisibilitySettings    string                   `json:"visibilitySettings" db:"visibility_settings"`
	WorkflowVersionId     int                      `json:"workflowVersionId" db:"workflow_version_id"`
	CreatedOn             time.Time                `json:"createdOn" db:"created_on"`
	UpdateOn              time.Time                `json:"updatedOn" db:"updated_on"`
}

func (wfn WorkflowVersionNode) GetWorkflowNodeStructured() WorkflowNode {
	return WorkflowNode{
		WorkflowVersionNodeId: wfn.WorkflowVersionNodeId,
		Name:                  wfn.Name,
		Status:                wfn.Status,
		Type:                  wfn.Type,
		Parameters:            wfn.Parameters,
		VisibilitySettings:    wfn.VisibilitySettings,
		UpdateOn:              wfn.UpdateOn,
		WorkflowVersionId:     wfn.WorkflowVersionId,
	}
}

type WorkflowVersionNodeLink struct {
	WorkflowVersionNodeLinkId   int       `json:"workflowVersionNodeLinkId" db:"workflow_version_node_link_id"`
	VisibilitySettings          string    `json:"visibilitySettings" db:"visibility_settings"`
	ParentParameterIndex        int       `json:"parentParameterIndex" db:"parent_parameter_index"`
	ParentWorkflowVersionNodeId int       `json:"parentWorkflowVersionNodeId" db:"parent_workflow_version_node_id"`
	ChildParameterIndex         int       `json:"childParameterIndex" db:"child_parameter_index"`
	ChildWorkflowVersionNodeId  int       `json:"childWorkflowVersionNodeId" db:"child_workflow_version_node_id"`
	CreatedOn                   time.Time `json:"createdOn" db:"created_on"`
	UpdateOn                    time.Time `json:"updatedOn" db:"updated_on"`
}

type WorkflowVersionNodeLog struct {
	NodeId                         int       `json:"nodeId" db:"node_id"`
	TriggerReference               string    `json:"triggerReference" db:"trigger_reference"`
	InputData                      string    `json:"input_data" db:"input_data"`
	OutputData                     string    `json:"output_data" db:"output_data"`
	DebugData                      string    `json:"debug_data" db:"debug_data"`
	ErrorData                      string    `json:"error_data" db:"error_data"`
	WorkflowVersionNodeId          int       `json:"workflowVersionNodeId" db:"workflow_version_node_id"`
	TriggeredWorkflowVersionNodeId int       `json:"triggeredWorkflowVersionNodeId" db:"triggered_workflow_version_node_id"`
	CreatedOn                      time.Time `json:"createdOn" db:"created_on"`
}

type WorkflowNode struct {
	WorkflowVersionNodeId int                      `json:"workflowVersionNodeId"`
	Name                  string                   `json:"name"`
	Status                commons.Status           `json:"status"`
	Type                  commons.WorkflowNodeType `json:"type"`
	Parameters            string                   `json:"parameters"`
	VisibilitySettings    string                   `json:"visibilitySettings"`
	UpdateOn              time.Time                `json:"updatedOn"`
	ParentNodes           map[int][]*WorkflowNode  `json:"parentNodes"`
	ChildNodes            map[int][]*WorkflowNode  `json:"childNodes"`
	WorkflowVersionId     int                      `json:"workflowVersionId"`
}

type WorkflowTree struct {
	RootNodes []*WorkflowNode `json:"rootNodes"`
}

type WorkflowNodeParameter struct {
	Type        commons.WorkflowParameterType `json:"type"`
	ValueNumber int                           `json:"valueNumber"`
	ValueString string                        `json:"valueString"`
}

type WorkflowNodeParameters struct {
	Parameters []WorkflowNodeParameter `json:"parameters"`
}
