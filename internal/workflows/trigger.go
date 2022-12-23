package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/commons"
)

// ProcessWorkflowNode workflowNodeStagingParametersCache[WorkflowVersionNodeId][inputLabel] (i.e. inputLabel = sourceChannelIds)
func ProcessWorkflowNode(ctx context.Context, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, workflowNode WorkflowNode, triggeredWorkflowVersionNodeId int,
	workflowNodeCache map[int]WorkflowNode, workflowNodeStatus map[int]commons.Status, workflowNodeStagingParametersCache map[int]map[string]string, reference string, inputs map[string]string,
	eventChannel chan interface{}, iteration int) (map[string]string, commons.Status, error) {

	iteration++
	if iteration > 100 {
		return nil, commons.Inactive, errors.New(fmt.Sprintf("Infinite loop for WorkflowVersionId: %v", workflowNode.WorkflowVersionId))
	}
	select {
	case <-ctx.Done():
		return nil, commons.Inactive, errors.New(fmt.Sprintf("Context terminated for WorkflowVersionId: %v", workflowNode.WorkflowVersionId))
	default:
	}
	outputs := commons.CopyParameters(inputs)
	var err error

	workflowNodeCached, cached := workflowNodeCache[workflowNode.WorkflowVersionNodeId]
	if cached {
		workflowNode = workflowNodeCached
	} else {
		// Obtain workflowNode because parent and child aren't completely populated
		workflowNode, err = GetWorkflowNode(db, workflowNode.WorkflowVersionNodeId)
		if err != nil {
			// Probably doesn't make sense to wrap in recursive loop
			return nil, commons.Inactive, err
		}
		workflowNodeCache[workflowNode.WorkflowVersionNodeId] = workflowNode
	}

	if workflowNode.Status == commons.Active {
		status, exists := workflowNodeStatus[workflowNode.WorkflowVersionNodeId]
		if exists && status == commons.Active {
			// When the node is in the cache and active then it's already been processed successfully
			return nil, commons.Deleted, nil
		}
		var processingStatus commons.Status
		processingStatus, inputs = getWorkflowNodeInputsStatus(workflowNode, inputs, workflowNodeStagingParametersCache[workflowNode.WorkflowVersionNodeId])
		if processingStatus == commons.Pending {
			// When the node is pending then not all inputs are available yet
			return nil, commons.Pending, nil
		}
		parameters := workflowNode.Parameters
		activeOutputIndex := -1
		switch workflowNode.Type {
		case commons.WorkflowNodeSetVariable:
			variableName := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableName).ValueString
			stringVariableParameter := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueString)
			if stringVariableParameter.ValueString != "" {
				outputs[variableName] = stringVariableParameter.ValueString
			} else {
				outputs[variableName] = fmt.Sprintf("%d", getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
			}
		case commons.WorkflowNodeFilterOnVariable:
			variableName := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableName).ValueString
			stringVariableParameter := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueString)
			stringValue := ""
			if stringVariableParameter.ValueString != "" {
				stringValue = stringVariableParameter.ValueString
			} else {
				stringValue = fmt.Sprintf("%d", getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
			}
			if inputs[variableName] == stringValue {
				activeOutputIndex = 0
			} else {
				activeOutputIndex = 1
			}
		case commons.WorkflowNodeChannelFilter:

		case commons.WorkflowNodeStageTrigger:
			if iteration > 0 {
				// There shouldn't be any stage nodes except when it's the first node
				return outputs, commons.Deleted, nil
			}
		case commons.WorkflowNodeRoutingPolicyParameters:
		case commons.WorkflowNodeRebalanceParameters:
		case commons.WorkflowNodeRebalanceRun:
			//if eventChannel != nil {
			//	eventChannel <- commons.RebalanceRequest{
			//		NodeId: nodeSettings.NodeId,
			//		SourceChannelIds: strconv.Atoi(inputs["channelId"]),
			//		DestinationChannelIds: strconv.Atoi(inputs["channelId"]),
			//		MaxCost: ,
			//	}
			//}
		case commons.WorkflowNodeRoutingPolicyRun:
			//if eventChannel != nil {
			//	eventChannel <- commons.ChannelStatusUpdateRequest{
			//		NodeId: nodeSettings.NodeId,
			//		ChannelId: strconv.Atoi(inputs["channelId"]),
			//		ChannelStatus: ,
			//	}
			//	eventChannel <- commons.RoutingPolicyUpdateRequest{
			//		NodeId: nodeSettings.NodeId,
			//		ChannelId: strconv.Atoi(inputs["channelId"]),
			//		FeeRateMilliMsat: ,
			//		FeeBaseMsat: ,
			//		MinHtlcMsat: ,
			//		MaxHtlcMsat: ,
			//		TimeLockDelta: ,
			//	}
			//}
		}
		workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = commons.Active
		for childLinkId, childNode := range workflowNode.ChildNodes {
			if activeOutputIndex != -1 && workflowNode.LinkDetails[childLinkId].ParentOutputIndex != activeOutputIndex {
				continue
			}
			if workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] == nil {
				workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] = make(map[string]string)
			}
			childParameters := commons.GetWorkflowNodes()[childNode.Type]
			var parameterWithLabel commons.WorkflowParameterWithLabel
			if workflowNode.LinkDetails[childLinkId].ChildInputIndex >= len(childParameters.RequiredInputs) {
				parameterWithLabel = childParameters.OptionalInputs[workflowNode.LinkDetails[childLinkId].ChildInputIndex-len(childParameters.RequiredInputs)]
			} else {
				parameterWithLabel = childParameters.RequiredInputs[workflowNode.LinkDetails[childLinkId].ChildInputIndex]
			}
			for key, value := range outputs {
				if key == string(parameterWithLabel.WorkflowParameter) || parameterWithLabel.WorkflowParameter == commons.WorkflowParameterAny {
					workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId][parameterWithLabel.Label] = value
				}
			}
			childOutputs, childProcessingStatus, err := ProcessWorkflowNode(ctx, db, nodeSettings, *childNode, workflowNode.WorkflowVersionNodeId,
				workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache, reference, outputs, eventChannel, iteration)
			if childProcessingStatus != commons.Pending {
				AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference,
					workflowNode.WorkflowVersionNodeId, triggeredWorkflowVersionNodeId, inputs, childOutputs, err)
			}
			if err != nil {
				// Probably doesn't make sense to wrap in recursive loop
				return nil, commons.Inactive, err
			}
		}
	}
	return outputs, commons.Active, nil
}

func AddWorkflowVersionNodeLog(db *sqlx.DB, nodeId int, reference string, workflowVersionNodeId int,
	triggeredWorkflowVersionNodeId int, inputs map[string]string, outputs map[string]string, workflowError error) {

	workflowVersionNodeLog := WorkflowVersionNodeLog{
		NodeId:                         nodeId,
		WorkflowVersionNodeId:          workflowVersionNodeId,
		TriggeredWorkflowVersionNodeId: triggeredWorkflowVersionNodeId,
		TriggerReference:               reference,
	}
	if len(inputs) > 0 {
		marshalledInputs, err := json.Marshal(inputs)
		if err == nil {
			workflowVersionNodeLog.InputData = string(marshalledInputs)
		} else {
			log.Error().Err(err).Msgf("Failed to marshal inputs for WorkflowVersionNodeId: %v", workflowVersionNodeId)
		}
	}
	if len(outputs) > 0 {
		marshalledOutputs, err := json.Marshal(outputs)
		if err == nil {
			workflowVersionNodeLog.OutputData = string(marshalledOutputs)
		} else {
			log.Error().Err(err).Msgf("Failed to marshal outputs for WorkflowVersionNodeId: %v", workflowVersionNodeId)
		}
	}
	if workflowError != nil {
		workflowVersionNodeLog.ErrorData = workflowError.Error()
	}
	_, err := addWorkflowVersionNodeLog(db, workflowVersionNodeLog)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to log root node execution for workflowVersionNodeId: %v", workflowVersionNodeId)
	}
}
