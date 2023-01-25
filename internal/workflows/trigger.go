package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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
	workflowNodeCache map[int]WorkflowNode, workflowNodeStatus map[int]commons.Status, workflowNodeStagingParametersCache map[int]map[commons.WorkflowParameterLabel]string,
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

	if workflowNode.Status == Active {
		status, exists := workflowNodeStatus[workflowNode.WorkflowVersionNodeId]
		if exists && status == commons.Active {
			// When the node is in the cache and active then it's already been processed successfully
			return nil, commons.Deleted, nil
		}
		var complete bool
		complete, inputs = getWorkflowNodeInputsComplete(workflowNode, inputs, workflowNodeStagingParametersCache[workflowNode.WorkflowVersionNodeId])
		if !complete {
			// When the node is pending then not all inputs are available yet
			return nil, commons.Pending, nil
		}

		cachedWorkflowNode, exists := workflowNodeCache[workflowNode.WorkflowVersionNodeId]
		if !exists {
			cachedWorkflowNode, err = GetWorkflowNode(db, workflowNode.WorkflowVersionNodeId)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "GetWorkflowNode for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			workflowNodeCache[workflowNode.WorkflowVersionNodeId] = cachedWorkflowNode
		}
		// Do not override when it's a grouped node because the data was manipulated.
		if !commons.IsWorkflowNodeTypeGrouped(workflowNode.Type) {
			workflowNode = cachedWorkflowNode
		}

		parentLinkedInputs := make(map[commons.WorkflowParameterLabel][]WorkflowNode)
		for parentWorkflowNodeLinkId, parentWorkflowNode := range workflowNode.ParentNodes {
			parentLink := workflowNode.LinkDetails[parentWorkflowNodeLinkId]
			parentLinkedInputs[parentLink.ChildInput] = append(parentLinkedInputs[parentLink.ChildInput], *parentWorkflowNode)
		}
	linkedInputLoop:
		for _, parentWorkflowNodesByInput := range parentLinkedInputs {
			for _, parentWorkflowNode := range parentWorkflowNodesByInput {
				status, exists = workflowNodeStatus[parentWorkflowNode.WorkflowVersionNodeId]
				if exists && status == commons.Active {
					continue linkedInputLoop
				}
			}
			// Not all inputs are available yet
			return nil, commons.Pending, nil
		}

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
			var params FilterClauses
			err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &params)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			noFiltering := params.Filter.FuncName == ""

			linkedChannelIdsString, exists := inputs[commons.WorkflowParameterLabelChannels]
			if exists {
				var linkedChannelIds []int
				err = json.Unmarshal([]byte(linkedChannelIdsString), &linkedChannelIds)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Unmarshal the parent channelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIdsString, workflowNode.WorkflowVersionNodeId)
				}
				if noFiltering {
					outputs[commons.WorkflowParameterLabelChannels] = linkedChannelIdsString
				} else {
					linkedChannels, err := channels.GetChannelsByIds(db, nodeSettings.NodeId, linkedChannelIds)
					if err != nil {
						return nil, commons.Inactive, errors.Wrapf(err, "Getting the linked channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
					}
					filteredChannels := ApplyFilters(params, StructToMap(linkedChannels))
					// Todo Add filteredChannels to outputs
					fmt.Printf("\n ------- filteredChannels %#v \n\n", filteredChannels)
				}
			} else {
				// Force Response because we don't care about balance accuracy
				channelIds := commons.GetChannelStateChannelIds(nodeSettings.NodeId, true)
				if noFiltering {
					ba, err := json.Marshal(channelIds)
					if err != nil {
						return nil, commons.Inactive, errors.Wrapf(err, "Marshal the channelIds: %v for WorkflowVersionNodeId: %v", channelIds, workflowNode.WorkflowVersionNodeId)
					}
					outputs[commons.WorkflowParameterLabelChannels] = string(ba)
				} else {
					channelsBodyByNode, err := channels.GetChannelsByIds(db, nodeSettings.NodeId, channelIds)
					if err != nil {
						return nil, commons.Inactive, errors.Wrapf(err, "Getting the linked channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
					}
					filteredChannels := ApplyFilters(params, StructToMap(channelsBodyByNode))
					// Todo Add filteredChannels to outputs
					fmt.Printf("\n ------- filteredChannels %#v \n\n", filteredChannels)
				}
			}

		case commons.WorkflowNodeAddTag, commons.WorkflowNodeRemoveTag:
			var linkedChannelIds []int
			linkedChannelIdsString, exists := inputs[commons.WorkflowParameterLabelChannels]
			if !exists {
				return nil, commons.Inactive, errors.Wrapf(err, "Finding channel parameter for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			err = json.Unmarshal([]byte(linkedChannelIdsString), &linkedChannelIds)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Unmarshal the parent channelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIdsString, workflowNode.WorkflowVersionNodeId)
			}

			if len(linkedChannelIds) != 0 {
				var params TagParameters
				err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &params)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				for _, tagToDelete := range params.RemovedTags {
					for index := range linkedChannelIds {
						tag := tags.TagEntityRequest{
							ChannelId: &linkedChannelIds[index],
							TagId:     tagToDelete.Value,
						}
						err := tags.UntagEntity(db, tag)
						if err != nil {
							return nil, commons.Inactive, errors.Wrapf(err, "Failed to remove the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagToDelete.Value)
						}
					}
				}

				for _, tagtoAdd := range params.AddedTags {
					for index := range linkedChannelIds {
						tag := tags.TagEntityRequest{
							ChannelId: &linkedChannelIds[index],
							TagId:     tagtoAdd.Value,
						}
						err := tags.TagEntity(db, tag)
						if err != nil {
							return nil, commons.Inactive, errors.Wrapf(err, "Failed to add the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagtoAdd.Value)
						}
					}
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
			childInput := commons.WorkflowParameterLabel(workflowNode.LinkDetails[childLinkId].ChildInput)
			parentOutput := commons.WorkflowParameterLabel(workflowNode.LinkDetails[childLinkId].ParentOutput)
			if childInput != "" && parentOutput != "" && outputs[parentOutput] != "" {
				workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId][childInput] = outputs[parentOutput]
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
		log.Error().Err(err).Msgf("Failed to log root node execution for workflowVersionNodeId: %v, nodeId: %#v", workflowVersionNodeId, nodeId)
	}
}

func StructToMap(structs []channels.ChannelBody) []map[string]interface{} {
	var maps []map[string]interface{}
	for _, s := range structs {
		structValue := reflect.ValueOf(s)
		structType := reflect.TypeOf(s)
		mapValue := make(map[string]interface{})

		for i := 0; i < structValue.NumField(); i++ {
			field := structType.Field(i)
			mapValue[strings.ToLower(field.Name)] = structValue.Field(i).Interface()
		}
		maps = append(maps, mapValue)
	}

	return maps
}
