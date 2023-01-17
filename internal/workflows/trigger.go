package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channel_groups"
	"github.com/lncapital/torq/pkg/commons"
)

// ProcessWorkflowNode workflowNodeStagingParametersCache[WorkflowVersionNodeId][inputLabel] (i.e. inputLabel = sourceChannelIds)
func ProcessWorkflowNode(ctx context.Context, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, workflowNode WorkflowNode, triggeringWorkflowVersionNodeId int,
	workflowNodeCache map[int]WorkflowNode, workflowNodeStatus map[int]commons.Status, workflowNodeStagingParametersCache map[int]map[string]string,
	reference string, inputs map[string]string, iteration int) (map[string]string, commons.Status, error) {

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
		//parameters := workflowNode.Parameters
		activeOutputIndex := -1
		switch workflowNode.Type {
		case commons.WorkflowNodeSetVariable:
			//variableName := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableName).ValueString
			//stringVariableParameter := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueString)
			//if stringVariableParameter.ValueString != "" {
			//	outputs[variableName] = stringVariableParameter.ValueString
			//} else {
			//	outputs[variableName] = fmt.Sprintf("%d", getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
			//}
		case commons.WorkflowNodeFilterOnVariable:
			//variableName := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableName).ValueString
			//stringVariableParameter := getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueString)
			//stringValue := ""
			//if stringVariableParameter.ValueString != "" {
			//	stringValue = stringVariableParameter.ValueString
			//} else {
			//	stringValue = fmt.Sprintf("%d", getWorkflowNodeParameter(parameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
			//}
			//if inputs[variableName] == stringValue {
			//	activeOutputIndex = 0
			//} else {
			//	activeOutputIndex = 1
			//}
		case commons.WorkflowNodeChannelFilter:

		case commons.WorkflowTag:
			var params TagParameters
			err := json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &params)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				return nil, commons.Inactive, err
			}
			var tagsToAdd []int
			for _, tagtoAdd := range params.AddedTags {
				tagsToAdd = append(tagsToAdd, tagtoAdd.Value)
			}
			err = channel_groups.AddChannelGroupByTags(db, tagsToAdd)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to add the tags for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				return nil, commons.Inactive, err
			}

			var tagsToDelete []int
			for _, tagToDelete := range params.RemovedTags {
				tagsToDelete = append(tagsToDelete, tagToDelete.Value)
			}

			_, err = channel_groups.RemoveChannelGroupByTags(db, tagsToDelete)
			if err != nil {
				log.Error().Err(err).Msgf("Failed remove the tags for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				return nil, commons.Inactive, err
			}

		case commons.WorkflowNodeTimeTrigger:
			childNodes, childNodeLinkDetails, err := GetTriggerGoupChildNodes(db, workflowNode.WorkflowVersionNodeId)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain the group nodes links")
				return nil, commons.Inactive, err
			}
			workflowNode.ChildNodes = childNodes
			workflowNode.LinkDetails = childNodeLinkDetails
			_, triggerCancel := context.WithCancel(context.Background())
			commons.SetTrigger(nodeSettings.NodeId, reference, workflowNode.WorkflowVersionId, triggeringWorkflowVersionNodeId, commons.Inactive, triggerCancel)

		case commons.WorkflowNodeStageTrigger:
			if iteration > 0 {
				// There shouldn't be any stage nodes except when it's the first node
				return outputs, commons.Deleted, nil
			}
		case commons.WorkflowNodeChannelPolicyConfigurator:
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
		case commons.WorkflowNodeChannelPolicyRun:
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
			// If activeOutputIndex is not -1 and the parent output index of the child link is not equal to activeOutputIndex, skip the rest of the loop body
			if activeOutputIndex != -1 && workflowNode.LinkDetails[childLinkId].ParentOutputIndex != activeOutputIndex {
				continue
			}
			// If there is no entry in the workflowNodeStagingParametersCache map for the child node's workflow version node ID, initialize an empty map
			if workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] == nil {
				workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] = make(map[string]string)
			}
			// Retrieve the child node's parameters from the commons.GetWorkflowNodes() map
			childParameters := commons.GetWorkflowNodes()[childNode.Type]
			var parameterWithLabel commons.WorkflowParameterWithLabel
			// Initialize parameterWithLabel to the appropriate input parameter for the child node (either required or optional)
			if workflowNode.LinkDetails[childLinkId].ChildInputIndex >= len(childParameters.RequiredInputs) {
				parameterWithLabel = childParameters.OptionalInputs[workflowNode.LinkDetails[childLinkId].ChildInputIndex-len(childParameters.OptionalInputs)]
			} else {
				parameterWithLabel = childParameters.RequiredInputs[workflowNode.LinkDetails[childLinkId].ChildInputIndex]
			}
			// Iterate over the outputs map and, for each key-value pair, add an entry to the workflowNodeStagingParametersCache map for
			// the child node if the key is equal to the string representation of parameterWithLabel.WorkflowParameter or
			// if parameterWithLabel.WorkflowParameter is equal to commons.WorkflowParameterAny
			for key, value := range outputs {
				if key == string(parameterWithLabel.WorkflowParameter) || parameterWithLabel.WorkflowParameter == commons.WorkflowParameterAny {
					workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId][parameterWithLabel.Label] = value
				}
			}
			// Call ProcessWorkflowNode with several arguments, including childNode, workflowNode.WorkflowVersionNodeId, and workflowNodeStagingParametersCache
			childOutputs, childProcessingStatus, err := ProcessWorkflowNode(ctx, db, nodeSettings, *childNode, workflowNode.WorkflowVersionNodeId,
				workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache, reference, outputs, iteration)
			// If childProcessingStatus is not equal to commons.Pending, call AddWorkflowVersionNodeLog with several arguments, including nodeSettings.NodeId, reference, workflowNode.WorkflowVersionNodeId, triggeringWorkflowVersionNodeId, inputs, and childOutputs
			if childProcessingStatus != commons.Pending {
				AddWorkflowVersionNodeLog(db, nodeSettings.NodeId, reference,
					workflowNode.WorkflowVersionNodeId, triggeringWorkflowVersionNodeId, inputs, childOutputs, err)
			}
			// If err is not nil, return nil, commons.Inactive, and err
			if err != nil {
				// Probably doesn't make sense to wrap in recursive loop
				return nil, commons.Inactive, err
			}
		}
	}
	return outputs, commons.Active, nil
}

func AddWorkflowVersionNodeLog(db *sqlx.DB, nodeId int, reference string, workflowVersionNodeId int,
	triggeringWorkflowVersionNodeId int, inputs map[string]string, outputs map[string]string, workflowError error) {
	workflowVersionNodeLog := WorkflowVersionNodeLog{
		NodeId:                nodeId,
		WorkflowVersionNodeId: workflowVersionNodeId,
		TriggerReference:      reference,
	}
	if triggeringWorkflowVersionNodeId > 0 {
		workflowVersionNodeLog.TriggeringWorkflowVersionNodeId = &triggeringWorkflowVersionNodeId
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
