package workflows

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/pkg/commons"
)

type Workflow struct {
	WorkflowId int            `json:"workflowId" db:"workflow_id"`
	Name       string         `json:"name" db:"name"`
	Status     commons.Status `json:"status" db:"status"`
	CreatedOn  time.Time      `json:"createdOn" db:"created_on"`
	UpdateOn   time.Time      `json:"updatedOn" db:"updated_on"`
}

type UpdateWorkflow struct {
	WorkflowId int             `json:"workflowId" db:"workflow_id"`
	Name       *string         `json:"name" db:"name"`
	Status     *commons.Status `json:"status" db:"status"`
}

type DeleteStageRequest struct {
	WorkflowVersionId int `json:"workflowVersionId"`
	Stage             int `json:"stage"`
}

type WorkflowTableRow struct {
	WorkflowId              int     `json:"workflowId" db:"workflow_id"`
	WorkflowName            string  `json:"workflowName" db:"workflow_name"`
	WorkflowStatus          int     `json:"workflowStatus" db:"workflow_status"`
	LatestVersionName       *string `json:"latestVersionName" db:"latest_version_name"`
	LatestVersion           *int    `json:"latestVersion" db:"latest_version"`
	LatestWorkflowVersionId *int    `json:"latestWorkflowVersionId" db:"latest_workflow_version_id"`
	LatestVersionStatus     *int    `json:"latestVersionStatus" db:"latest_version_status"`
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

type TimeTriggerParameters struct {
	Seconds  int32 `json:"seconds" db:"seconds"`
	TimeUnit int   `json:"timeUnit" db:"time_unit"` // Time Unit is just used in the frontend
}

type ChannelPolicyConfigurationParameters struct {
	FeeRate       int64 `json:"feeRate"`
	BaseFee       int64 `json:"baseFee"`
	MaxHtlcAmount int64 `json:"maxHtlcAmount"`
	MinHtlcAmount int64 `json:"minHtlcAmount"`
}

type TagAction int

const (
	TagActionAdd = TagAction(iota)
	TagActionRemove
	TagActionToggle
)

type ModifyTagsParameters struct {
	TagNames  []string  `json:"tagNames"`
	TagAction TagAction `json:"tagAction"`
}

type ReBalanceParameters struct {
	OutgoingChannelId  *int   `json:"outgoingChannelId"`
	IncomingChannelId  *int   `json:"incomingChannelId"`
	ChannelIds         []int  `json:"channelIds"`
	AmountMsat         uint64 `json:"amountMsat"`
	MaximumCostMsat    uint64 `json:"maximumCostMsat"`
	MaximumConcurrency *int   `json:"maximumConcurrency"`
}

type TagParameters struct {
	AddedTags   []TagInfo `json:"addedTags"`
	RemovedTags []TagInfo `json:"removedTags"`
}

type TagInfo struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type WorkflowVersionNode struct {
	WorkflowId            int                            `json:"workflowId" db:"workflow_id"`
	Version               int                            `json:"version" db:"version"`
	WorkflowVersionId     int                            `json:"workflowVersionId" db:"workflow_version_id"`
	WorkflowVersionNodeId int                            `json:"workflowVersionNodeId" db:"workflow_version_node_id"`
	Name                  string                         `json:"name" db:"name"`
	Stage                 int                            `json:"stage" db:"stage"`
	Status                commons.Status                 `json:"status" db:"status"`
	Type                  commons.WorkflowNodeType       `json:"type" db:"type"`
	Parameters            interface{}                    `json:"parameters" db:"parameters"`
	VisibilitySettings    WorkflowNodeVisibilitySettings `json:"visibilitySettings" db:"visibility_settings"`
	CreatedOn             time.Time                      `json:"createdOn" db:"created_on"`
	UpdateOn              time.Time                      `json:"updatedOn" db:"updated_on"`
}

type CreateNodeRequest struct {
	WorkflowVersionId  int                            `json:"workflowVersionId" db:"workflow_version_id"`
	Name               string                         `json:"name" db:"name"`
	Type               commons.WorkflowNodeType       `json:"type" db:"type"`
	Stage              int                            `json:"stage" db:"stage"`
	VisibilitySettings WorkflowNodeVisibilitySettings `json:"visibilitySettings" db:"visibility_settings"`
	Parameters         *interface{}                   `json:"parameters" db:"parameters"`
}

type UpdateNodeRequest struct {
	WorkflowVersionNodeId int                             `json:"workflowVersionNodeId" db:"workflow_version_node_id"`
	Name                  *string                         `json:"name" db:"name"`
	Status                *commons.Status                 `json:"status" db:"status"`
	Parameters            *interface{}                    `json:"parameters" db:"parameters"`
	VisibilitySettings    *WorkflowNodeVisibilitySettings `json:"visibilitySettings" db:"visibility_settings"`
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

type LinkVisibilitySettings struct {
	XPosition int `json:"xPosition"`
	YPosition int `json:"yPosition"`
}

type WorkflowVersionNodeLinkVisibilitySettings struct {
	Child  LinkVisibilitySettings `json:"child"`
	Parent LinkVisibilitySettings `json:"parent"`
}

func (nvls *WorkflowVersionNodeLinkVisibilitySettings) Scan(val interface{}) (err error) {
	switch v := val.(type) {
	case []byte:
		err = json.Unmarshal(v, &nvls)
	}
	if err != nil {
		return errors.Wrapf(err, "Incompatible type for WorkflowNodeLinkVisibilitySettings")
	}
	return nil
}

type WorkflowVersionNodeLink struct {
	WorkflowVersionNodeLinkId   int                                       `json:"workflowVersionNodeLinkId" db:"workflow_version_node_link_id"`
	Name                        string                                    `json:"name" db:"name"`
	VisibilitySettings          WorkflowVersionNodeLinkVisibilitySettings `json:"visibilitySettings" db:"visibility_settings"`
	ParentOutput                commons.WorkflowParameterLabel            `json:"parentOutput" db:"parent_output"`
	ParentWorkflowVersionNodeId int                                       `json:"parentWorkflowVersionNodeId" db:"parent_workflow_version_node_id"`
	ChildInput                  commons.WorkflowParameterLabel            `json:"childInput" db:"child_input"`
	ChildWorkflowVersionNodeId  int                                       `json:"childWorkflowVersionNodeId" db:"child_workflow_version_node_id"`
	WorkflowVersionId           int                                       `json:"workflowVersionId" db:"workflow_version_id"`
	CreatedOn                   time.Time                                 `json:"createdOn" db:"created_on"`
	UpdateOn                    time.Time                                 `json:"updatedOn" db:"updated_on"`
	Stage                       int                                       `json:"stage" db:"stage"`
}

type CreateWorkflowVersionNodeLinkRequest struct {
	WorkflowVersionId           int                            `json:"workflowVersionId" db:"workflow_version_id"`
	ParentOutput                commons.WorkflowParameterLabel `json:"parentOutput" db:"parent_output"`
	ParentWorkflowVersionNodeId int                            `json:"parentWorkflowVersionNodeId" db:"parent_workflow_version_node_id"`
	ChildInput                  commons.WorkflowParameterLabel `json:"childInput" db:"child_input"`
	ChildWorkflowVersionNodeId  int                            `json:"childWorkflowVersionNodeId" db:"child_workflow_version_node_id"`
}

type WorkflowVersionNodeLog struct {
	NodeId                          int       `json:"nodeId" db:"node_id"`
	TriggerReference                string    `json:"triggerReference" db:"trigger_reference"`
	InputData                       string    `json:"input_data" db:"input_data"`
	OutputData                      string    `json:"output_data" db:"output_data"`
	DebugData                       string    `json:"debug_data" db:"debug_data"`
	ErrorData                       string    `json:"error_data" db:"error_data"`
	WorkflowVersionNodeId           int       `json:"workflowVersionNodeId" db:"workflow_version_node_id"`
	TriggeringWorkflowVersionNodeId *int      `json:"triggeringWorkflowVersionNodeId" db:"triggering_workflow_version_node_id"`
	CreatedOn                       time.Time `json:"createdOn" db:"created_on"`
}

type WorkflowNode struct {
	WorkflowVersionNodeId int                             `json:"workflowVersionNodeId"`
	Name                  string                          `json:"name"`
	Status                commons.Status                  `json:"status"`
	Type                  commons.WorkflowNodeType        `json:"type"`
	Parameters            interface{}                     `json:"parameters"`
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
	Workflow Workflow                  `json:"workflow"`
	Version  WorkflowVersion           `json:"version"`
	Nodes    []WorkflowVersionNode     `json:"nodes"`
	Links    []WorkflowVersionNodeLink `json:"links"`
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

func getWorkflowNodeInputsStatus(workflowNode WorkflowNode, inputs map[commons.WorkflowParameterLabel]string,
	stagingParameters map[commons.WorkflowParameterLabel]string) (commons.Status, map[commons.WorkflowParameterLabel]string) {

	// Get the required inputs for the workflow node's type
	requiredInputs := commons.GetWorkflowNodes()[workflowNode.Type].RequiredInputs
	// Iterate over the required inputs
	for label := range requiredInputs {
		// If the parameter is already a key in the inputs map, skip the rest of the loop body
		if _, exists := inputs[label]; exists {
			continue
		}
		// If the parameter is not a key in the inputs map, check if it is a key in the stagingParameters map
		inputData, exists := stagingParameters[label]
		if !exists {
			// If it is not, return commons.Pending and the inputs map
			return commons.Pending, inputs
		}
		// If it is, add an entry to the inputs map with the matching label and value from the stagingParameters map
		inputs[label] = inputData
	}

	optionalInputs := commons.GetWorkflowNodes()[workflowNode.Type].OptionalInputs
	for label := range optionalInputs {
		_, exists := inputs[label]
		if exists {
			continue
		}
		inputData, exists := stagingParameters[label]
		if !exists {
			continue
		}
		inputs[label] = inputData
	}

	// If all required inputs have been processed, return commons.Active and the inputs map
	return commons.Active, inputs
}
