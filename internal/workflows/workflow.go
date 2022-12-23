package workflows

import (
	"encoding/json"
	"github.com/cockroachdb/errors"
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

type Workflow struct {
	WorkflowId int            `json:"workflowId" db:"workflow_id"`
	Name       string         `json:"name" db:"name"`
	Status     commons.Status `json:"status" db:"status"`
	CreatedOn  time.Time      `json:"createdOn" db:"created_on"`
	UpdateOn   time.Time      `json:"updatedOn" db:"updated_on"`
}

type WorkflowTableRow struct {
	WorkflowId              int     `json:"workflowId" db:"workflow_id"`
	WorkflowName            string  `json:"workflowName" db:"workflow_name"`
	WorkflowStatus          int     `json:"workflowStatus" db:"workflow_status"`
	LatestVersionName       string  `json:"latestVersionName" db:"latest_version_name"`
	LatestVersion           int     `json:"latestVersion" db:"latest_version"`
	LatestWorkflowVersionId int     `json:"latestWorkflowVersionId" db:"latest_workflow_version_id"`
	LatestVersionStatus     int     `json:"latestVersionStatus" db:"latest_version_status"`
	ActiveVersionName       *string `json:"activeVersionName" db:"active_version_name"`
	ActiveVersion           *int    `json:"activeVersion" db:"active_version"`
	ActiveWorkflowVersionId *int    `json:"activeWorkflowVersionId" db:"active_workflow_version_id"`
	ActiveVersionStatus     *int    `json:"activeVersionStatus" db:"active_version_status"`
}

type WorkflowVersion struct {
	WorkflowVersionId int            `json:"workflowVersionId" db:"workflow_version_id"`
	Name              string         `json:"name" db:"name"`
	Version           int            `json:"version" db:"version"`
	Status            commons.Status `json:"status" db:"status"`
	WorkflowId        int            `json:"workflowId" db:"workflow_id"`
	CreatedOn         time.Time      `json:"createdOn" db:"created_on"`
	UpdateOn          time.Time      `json:"updatedOn" db:"updated_on"`
}

type WorkflowVersionNode struct {
	WorkflowVersionNodeId int                            `json:"workflowVersionNodeId" db:"workflow_version_node_id"`
	Name                  string                         `json:"name" db:"name"`
	Stage                 int                            `json:"stage" db:"stage"`
	Status                commons.Status                 `json:"status" db:"status"`
	Type                  commons.WorkflowNodeType       `json:"type" db:"type"`
	Parameters            WorkflowNodeParameters         `json:"parameters" db:"parameters"`
	VisibilitySettings    WorkflowNodeVisibilitySettings `json:"visibilitySettings" db:"visibility_settings"`
	WorkflowVersionId     int                            `json:"workflowVersionId" db:"workflow_version_id"`
	CreatedOn             time.Time                      `json:"createdOn" db:"created_on"`
	UpdateOn              time.Time                      `json:"updatedOn" db:"updated_on"`
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
	Name                        string    `json:"name" db:"name"`
	VisibilitySettings          string    `json:"visibilitySettings" db:"visibility_settings"`
	ParentOutputIndex           int       `json:"parentOutputIndex" db:"parent_output_index"`
	ParentWorkflowVersionNodeId int       `json:"parentWorkflowVersionNodeId" db:"parent_workflow_version_node_id"`
	ChildInputIndex             int       `json:"childInputIndex" db:"child_input_index"`
	ChildWorkflowVersionNodeId  int       `json:"childWorkflowVersionNodeId" db:"child_workflow_version_node_id"`
	WorkflowVersionId           int       `json:"workflowVersionId" db:"workflow_version_id"`
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
	WorkflowVersionNodeId int                             `json:"workflowVersionNodeId"`
	Name                  string                          `json:"name"`
	Status                commons.Status                  `json:"status"`
	Type                  commons.WorkflowNodeType        `json:"type"`
	Parameters            WorkflowNodeParameters          `json:"parameters"`
	VisibilitySettings    WorkflowNodeVisibilitySettings  `json:"visibilitySettings"`
	UpdateOn              time.Time                       `json:"updatedOn"`
	ParentNodes           map[int]*WorkflowNode           `json:"parentNodes"`
	ChildNodes            map[int]*WorkflowNode           `json:"childNodes"`
	LinkDetails           map[int]WorkflowVersionNodeLink `json:"LinkDetails"`
	WorkflowVersionId     int                             `json:"workflowVersionId"`
}

type WorkflowForest struct {
	SortedStageTrees map[int][]*WorkflowNode `json:"sortedStageTrees"`
}

type WorkflowPage struct {
	Workflow       Workflow        `json:"workflow"`
	Version        WorkflowVersion `json:"version"`
	WorkflowForest WorkflowForest  `json:"workflowForest"`
}

type WorkflowNodeParameter struct {
	Type        commons.WorkflowParameterType `json:"type"`
	ValueNumber int                           `json:"valueNumber"`
	ValueString string                        `json:"valueString"`
}

func (wp *WorkflowNodeParameter) Scan(val interface{}) (err error) {
	switch v := val.(type) {
	case []byte:
		err = json.Unmarshal(v, &wp)
	}
	if err != nil {
		return errors.Wrapf(err, "Incompatible type for WorkflowNodeParameter")
	}
	return nil
}

type WorkflowNodeParameters struct {
	Parameters []WorkflowNodeParameter `json:"parameters"`
}

func (nvp *WorkflowNodeParameters) Scan(val interface{}) (err error) {
	switch v := val.(type) {
	case []byte:
		err = json.Unmarshal(v, &nvp)
	}
	if err != nil {
		return errors.Wrapf(err, "Incompatible type for WorkflowNodeParameters")
	}
	return nil
}

type WorkflowNodeVisibilitySettings struct {
	YPosition *int  `json:"yPosition" db:"yPosition"`
	XPosition *int  `json:"xPosition" db:"xPosition"`
	Collapsed *bool `json:"collapsed" db:"collapsed"`
}

func (nvs *WorkflowNodeVisibilitySettings) Scan(val interface{}) (err error) {
	switch v := val.(type) {
	case []byte:
		err = json.Unmarshal(v, &nvs)
	}
	if err != nil {
		return errors.Wrapf(err, "Incompatible type for WorkflowNodeVisibilitySettings")
	}
	return nil
}

func getWorkflowNodeInputsStatus(workflowNode WorkflowNode, inputs map[string]string,
	stagingParameters map[string]string) (commons.Status, map[string]string) {

	requiredInputs := commons.GetWorkflowNodes()[workflowNode.Type].RequiredInputs
	for _, label := range requiredInputs {
		_, exists := inputs[label.Label]
		if exists {
			continue
		}
		foundIt := false
		for stagingLabel, inputData := range stagingParameters {
			if label.Label == stagingLabel {
				foundIt = true
				inputs[label.Label] = inputData
				continue
			}
		}
		if !foundIt {
			return commons.Pending, inputs
		}
	}
	return commons.Active, inputs
}
