package workflows

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/pkg/commons"
)

// ProcessWorkflowNode workflowNodeStagingParametersCache[WorkflowVersionNodeId][parameterLabel] (i.e. parameterLabel = sourceChannels)
func ProcessWorkflowNode(ctx context.Context, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings, workflowNode WorkflowNode, triggeringWorkflowVersionNodeId int,
	workflowNodeStatus map[int]commons.Status, workflowNodeStagingParametersCache map[int]map[commons.WorkflowParameterLabel]string,
	reference string, inputs map[commons.WorkflowParameterLabel]string, iteration int) (map[commons.WorkflowParameterLabel]string, commons.Status, error) {

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
		var activeOutput commons.WorkflowParameterLabel
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
			fmt.Println("I am running channel filters")

			channels, err := channels.GetChannelsByNetwork(db, 0)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Failed to get the channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			fmt.Println("channelsx", channels)

		case commons.WorkflowNodeAddTag, commons.WorkflowNodeRemoveTag:
			var params TagParameters
			err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &params)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Failed to parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			for _, tagtoAdd := range params.AddedTags {
				tag := tags.TagEntityRequest{
					// TODO the nodeId and channelId will come from the input target when ready
					NodeId: &nodeSettings.NodeId,
					TagId:  tagtoAdd.Value,
				}
				err := tags.TagEntity(db, tag)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Failed to add the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagtoAdd.Value)
				}
			}

			for _, tagToDelete := range params.RemovedTags {
				tag := tags.TagEntityRequest{
					// TODO the nodeId and channelId will come from the input target when ready
					NodeId: &nodeSettings.NodeId,
					TagId:  tagToDelete.Value,
				}
				err := tags.UntagEntity(db, tag)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Failed to remove the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagToDelete.Value)
				}
			}

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
		case commons.WorkflowNodeChannelOpenEventTrigger:
			log.Debug().Msg("Channel Open Event Trigger Fired")
		case commons.WorkflowNodeChannelCloseEventTrigger:
			log.Debug().Msg("Channel Close Event Trigger Fired")
		}
		workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = commons.Active
		for childLinkId, childNode := range workflowNode.ChildNodes {
			// If activeOutputIndex is not -1 and the parent output index of the child link is not equal to activeOutputIndex, skip the rest of the loop body
			if activeOutput != "" && workflowNode.LinkDetails[childLinkId].ParentOutput != activeOutput {
				continue
			}
			// If there is no entry in the workflowNodeStagingParametersCache map for the child node's workflow version node ID, initialize an empty map
			if workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] == nil {
				workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] = make(map[commons.WorkflowParameterLabel]string)
			}
			childInput := workflowNode.LinkDetails[childLinkId].ChildInput
			parentOutput := workflowNode.LinkDetails[childLinkId].ParentOutput
			workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId][childInput] = outputs[parentOutput]
				if len(childParameters.OptionalInputs) > workflowNode.LinkDetails[childLinkId].ChildInputIndex {
					parameterWithLabel = childParameters.OptionalInputs[workflowNode.LinkDetails[childLinkId].ChildInputIndex]
				} else {
					parameterWithLabel = childParameters.OptionalInputs[workflowNode.LinkDetails[childLinkId].ChildInputIndex-len(childParameters.OptionalInputs)]
				}
			// Call ProcessWorkflowNode with several arguments, including childNode, workflowNode.WorkflowVersionNodeId, and workflowNodeStagingParametersCache
			childOutputs, childProcessingStatus, err := ProcessWorkflowNode(ctx, db, nodeSettings, *childNode, workflowNode.WorkflowVersionNodeId,
				workflowNodeStatus, workflowNodeStagingParametersCache, reference, outputs, iteration)
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

func AddWorkflowVersionNodeLog(db *sqlx.DB,
	nodeId int,
	reference string,
	workflowVersionNodeId int,
	triggeringWorkflowVersionNodeId int,
	inputs map[commons.WorkflowParameterLabel]string,
	outputs map[commons.WorkflowParameterLabel]string,
	workflowError error) {

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
		log.Error().Err(err).Msgf("Failed to log root node execution for workflowVersionNodeId: %v  NODE %#v", workflowVersionNodeId, nodeId)
	}
}
