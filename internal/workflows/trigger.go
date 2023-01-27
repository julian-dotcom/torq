package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/pkg/commons"
)

// ProcessWorkflowNode workflowNodeStagingParametersCache[WorkflowVersionNodeId][parameterLabel] (i.e. parameterLabel = sourceChannels)
func ProcessWorkflowNode(ctx context.Context, db *sqlx.DB,
	workflowNode WorkflowNode,
	triggeringWorkflowVersionNodeId int,
	workflowNodeCache map[int]WorkflowNode,
	workflowNodeStatus map[int]commons.Status,
	workflowNodeStagingParametersCache map[int]map[commons.WorkflowParameterLabel]string,
	reference string,
	inputs map[commons.WorkflowParameterLabel]string,
	iteration int,
	workflowStageExitConfigurationCache map[int]map[commons.WorkflowParameterLabel]string,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest) (map[commons.WorkflowParameterLabel]string, commons.Status, error) {

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

		var activeOutputs []commons.WorkflowParameterLabel
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
			filtered := params.Filter.FuncName != "" || len(params.Or) != 0 || len(params.And) != 0

			var filteredChannelIds []int
			linkedChannelIdsString, exists := inputs[commons.WorkflowParameterLabelChannels]
			if exists && linkedChannelIdsString != "" && linkedChannelIdsString != "null" {
				var linkedChannelIds []int
				err = json.Unmarshal([]byte(linkedChannelIdsString), &linkedChannelIds)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Unmarshal the parent channelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIdsString, workflowNode.WorkflowVersionNodeId)
				}
				if filtered {
					var linkedChannels []channels.ChannelBody
					torqNodeIds := commons.GetAllTorqNodeIds()
					for _, torqNodeId := range torqNodeIds {
						linkedChannelsByNode, err := channels.GetChannelsByIds(db, torqNodeId, linkedChannelIds)
						if err != nil {
							return nil, commons.Inactive, errors.Wrapf(err, "Getting the linked channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
						}
						linkedChannels = append(linkedChannels, linkedChannelsByNode...)
					}
					filteredChannelIds = filterChannelIds(params, linkedChannels)
				} else {
					filteredChannelIds = linkedChannelIds
				}
			} else {
				torqNodeIds := commons.GetAllTorqNodeIds()
				for _, torqNodeId := range torqNodeIds {
					// Force Response because we don't care about balance accuracy
					channelIdsByNode := commons.GetChannelStateChannelIds(torqNodeId, true)
					if filtered {
						channelsBodyByNode, err := channels.GetChannelsByIds(db, torqNodeId, channelIdsByNode)
						if err != nil {
							return nil, commons.Inactive, errors.Wrapf(err, "Getting the linked channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
						}
						filteredChannelIds = append(filteredChannelIds, filterChannelIds(params, channelsBodyByNode)...)
					} else {
						filteredChannelIds = append(filteredChannelIds, channelIdsByNode...)
					}
				}
			}
			if len(filteredChannelIds) == 0 {
				activeOutputs = append(activeOutputs,
					commons.WorkflowParameterLabelTimeTriggered,
					commons.WorkflowParameterLabelChannelEventTriggered)
			} else {
				ba, err := json.Marshal(filteredChannelIds)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshal the filteredChannelIds: %v for WorkflowVersionNodeId: %v", filteredChannelIds, workflowNode.WorkflowVersionNodeId)
				}
				outputs[commons.WorkflowParameterLabelChannels] = string(ba)
			}
		case commons.WorkflowNodeAddTag, commons.WorkflowNodeRemoveTag:
			linkedChannelIds, err := getLinkedChannelIds(inputs, workflowNode)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Getting the Linked ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			if len(linkedChannelIds) == 0 {
				activeOutputs = append(activeOutputs,
					commons.WorkflowParameterLabelTimeTriggered,
					commons.WorkflowParameterLabelChannelEventTriggered)
			} else {
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
		case commons.WorkflowNodeChannelPolicyConfigurator:
			linkedChannelIds, err := getLinkedChannelIds(inputs, workflowNode)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Getting the Linked ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			if len(linkedChannelIds) == 0 {
				activeOutputs = append(activeOutputs,
					commons.WorkflowParameterLabelTimeTriggered,
					commons.WorkflowParameterLabelChannelEventTriggered)
			} else {
				var channelPolicyInputConfiguration ChannelPolicyConfiguration
				channelPolicyInputConfigurationString, exists := inputs[commons.WorkflowParameterLabelRoutingPolicySettings]
				if exists && channelPolicyInputConfigurationString != "" && channelPolicyInputConfigurationString != "null" {
					err = json.Unmarshal([]byte(channelPolicyInputConfigurationString), &channelPolicyInputConfiguration)
					if err != nil {
						return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
					}
				}

				var channelPolicyConfiguration ChannelPolicyConfiguration
				err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &channelPolicyConfiguration)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				if channelPolicyConfiguration.FeeBaseMsat != nil {
					channelPolicyInputConfiguration.FeeBaseMsat = channelPolicyConfiguration.FeeBaseMsat
				}
				if channelPolicyConfiguration.FeeRateMilliMsat != nil {
					channelPolicyInputConfiguration.FeeRateMilliMsat = channelPolicyConfiguration.FeeRateMilliMsat
				}
				if channelPolicyConfiguration.MaxHtlcMsat != nil {
					channelPolicyInputConfiguration.MaxHtlcMsat = channelPolicyConfiguration.MaxHtlcMsat
				}
				if channelPolicyConfiguration.MinHtlcMsat != nil {
					channelPolicyInputConfiguration.MinHtlcMsat = channelPolicyConfiguration.MinHtlcMsat
				}
				channelPolicyInputConfiguration.ChannelIds = linkedChannelIds

				marshalledChannelPolicyConfiguration, err := json.Marshal(channelPolicyInputConfiguration)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				if workflowStageExitConfigurationCache[workflowNode.Stage] == nil {
					workflowStageExitConfigurationCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
				}
				workflowStageExitConfigurationCache[workflowNode.Stage][commons.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)
				outputs[commons.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)
			}
		case commons.WorkflowNodeChannelPolicyRun:
			routingPolicySettingsString := inputs[commons.WorkflowParameterLabelRoutingPolicySettings]
			var routingPolicySettings ChannelPolicyConfiguration
			err = json.Unmarshal([]byte(routingPolicySettingsString), &routingPolicySettings)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(routingPolicySettings.ChannelIds) == 0 {
				activeOutputs = append(activeOutputs,
					commons.WorkflowParameterLabelTimeTriggered,
					commons.WorkflowParameterLabelChannelEventTriggered)
			} else {
				activeOutputs = append(activeOutputs,
					commons.WorkflowParameterLabelChannels,
					commons.WorkflowParameterLabelStatus)
				inputs[commons.WorkflowParameterLabelStatus] = ""
				if (routingPolicySettings.FeeBaseMsat != nil ||
					routingPolicySettings.FeeRateMilliMsat != nil ||
					routingPolicySettings.MaxHtlcMsat != nil ||
					routingPolicySettings.MinHtlcMsat != nil ||
					routingPolicySettings.TimeLockDelta != nil) &&
					lightningRequestChannel != nil {

					now := time.Now()
					torqNodeIds := commons.GetAllTorqNodeIds()
					for index := range routingPolicySettings.ChannelIds {
						channelSettings := commons.GetChannelSettingByChannelId(routingPolicySettings.ChannelIds[index])
						nodeId := channelSettings.FirstNodeId
						if !slices.Contains(torqNodeIds, nodeId) {
							nodeId = channelSettings.SecondNodeId
						}
						if !slices.Contains(torqNodeIds, nodeId) {
							return nil, commons.Inactive, errors.Wrapf(err, "Routing policy update on unmanaged channel for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
						}
						if commons.RunningServices[commons.LightningCommunicationService].GetStatus(nodeId) == commons.Active {
							if hasRoutingPolicyChanges(nodeId, routingPolicySettings.ChannelIds[index], routingPolicySettings) {
								response := channels.SetRoutingPolicyWithTimeout(commons.RoutingPolicyUpdateRequest{
									CommunicationRequest: commons.CommunicationRequest{
										RequestId:   reference,
										RequestTime: &now,
										NodeId:      nodeId,
									},
									ChannelId:        routingPolicySettings.ChannelIds[index],
									FeeRateMilliMsat: routingPolicySettings.FeeRateMilliMsat,
									FeeBaseMsat:      routingPolicySettings.FeeBaseMsat,
									MaxHtlcMsat:      routingPolicySettings.MaxHtlcMsat,
									MinHtlcMsat:      routingPolicySettings.MinHtlcMsat,
									TimeLockDelta:    routingPolicySettings.TimeLockDelta,
								}, lightningRequestChannel)
								if response.Error != "" {
									log.Error().Err(errors.New(response.Error)).Msgf("Channel Time Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
								}
								marshalledResponse, err := json.Marshal(response)
								if err != nil {
									return nil, commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
								}
								inputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponse)
							}
						}
					}
				}
			}
		case commons.WorkflowNodeRebalanceConfigurator:
			var incomingChannelIds []int
			incomingChannelIdsString, exists := inputs[commons.WorkflowParameterLabelIncomingChannels]
			if !exists {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			err = json.Unmarshal([]byte(incomingChannelIdsString), &incomingChannelIds)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			var outgoingChannelIds []int
			outgoingChannelIdsString, exists := inputs[commons.WorkflowParameterLabelOutgoingChannels]
			if !exists {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			err = json.Unmarshal([]byte(outgoingChannelIdsString), &outgoingChannelIds)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			//if len(incomingChannelIds) == 0 && len(outgoingChannelIds) == 0 {
			//
			//}

			var rebalanceInputConfiguration RebalanceConfiguration
			rebalanceInputConfigurationString, exists := inputs[commons.WorkflowParameterLabelRebalanceSettings]
			if exists && rebalanceInputConfigurationString != "" && rebalanceInputConfigurationString != "null" {
				err = json.Unmarshal([]byte(rebalanceInputConfigurationString), &rebalanceInputConfiguration)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
			}

			var rebalanceConfiguration RebalanceConfiguration
			err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &rebalanceConfiguration)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			marshalledRebalanceConfiguration, err := json.Marshal(rebalanceInputConfiguration)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			outputs[commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
		case commons.WorkflowNodeRebalanceRun:
			rebalanceSettingsString := inputs[commons.WorkflowParameterLabelRebalanceSettings]
			var rebalanceSettings RebalanceConfiguration
			err = json.Unmarshal([]byte(rebalanceSettingsString), &rebalanceSettings)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			//if len(sourceChannelIds) != 0 &&
			//	len(destinationChannelIds) != 0 &&
			//	rebalanceRequestChannel != nil &&
			//	commons.RunningServices[commons.RebalanceService].GetStatus(nodeSettings.NodeId) == commons.Active {
			//
			//	now := time.Now()
			//	responseChannel := make(chan commons.RebalanceResponse, 1)
			//	request := commons.RebalanceRequest{
			//		CommunicationRequest: commons.CommunicationRequest{
			//			RequestId:   reference,
			//			RequestTime: &now,
			//			NodeId:      nodeSettings.NodeId,
			//		},
			//		ResponseChannel:    responseChannel,
			//		Origin:             commons.RebalanceRequestWorkflowNode,
			//		OriginId:           workflowNode.WorkflowVersionNodeId,
			//		OriginReference:    reference,
			//		ChannelIds:         rebalanceSettings.ChannelIds,
			//		AmountMsat:         *rebalanceSettings.AmountMsat,
			//		MaximumCostMsat:    *rebalanceSettings.MaximumCostMsat,
			//		MaximumConcurrency: 1,
			//	}
			//	if rebalanceSettings.IncomingChannelIds != nil {
			//		request.IncomingChannelId = rebalanceSettings.IncomingChannelIds
			//	}
			//	if rebalanceSettings.OutgoingChannelIds != nil {
			//		request.OutgoingChannelId = *rebalanceSettings.OutgoingChannelIds
			//	}
			//	lightningRequestChannel <- request
			//time.AfterFunc(commons.LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS*time.Second, func() {
			//	message := fmt.Sprintf("Routing policy update timed out after %v seconds.", commons.LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS)
			//	responseChannel <- commons.RoutingPolicyUpdateResponse{
			//		Request: request,
			//		CommunicationResponse: commons.CommunicationResponse{
			//			Status:  commons.TimedOut,
			//			Message: "Routing policy update timed out after 2 seconds.",
			//			Error:   "Routing policy update timed out after 2 seconds.",
			//		},
			//	}
			//})
			//	response := <-responseChannel
			//	if response.Error != "" {
			//		log.Error().Err(errors.New(response.Error)).Msg("Channel Time Trigger Fired")
			//	}
			//}
		case commons.WorkflowNodeTimeTrigger:
			log.Debug().Msgf("Channel Time Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			if workflowStageExitConfigurationCache[workflowNode.Stage] == nil {
				workflowStageExitConfigurationCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
			}
			workflowStageExitConfigurationCache[workflowNode.Stage][commons.WorkflowParameterLabelTimeTriggered] = inputs[commons.WorkflowParameterLabelTimeTriggered]
		case commons.WorkflowNodeChannelBalanceEventTrigger:
			log.Debug().Msgf("Channel Balance Event Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			if workflowStageExitConfigurationCache[workflowNode.Stage] == nil {
				workflowStageExitConfigurationCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
			}
			workflowStageExitConfigurationCache[workflowNode.Stage][commons.WorkflowParameterLabelChannelEventTriggered] = inputs[commons.WorkflowParameterLabelChannelEventTriggered]
		case commons.WorkflowNodeChannelOpenEventTrigger:
			log.Debug().Msgf("Channel Open Event Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			if workflowStageExitConfigurationCache[workflowNode.Stage] == nil {
				workflowStageExitConfigurationCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
			}
			workflowStageExitConfigurationCache[workflowNode.Stage][commons.WorkflowParameterLabelChannelEventTriggered] = inputs[commons.WorkflowParameterLabelChannelEventTriggered]
		case commons.WorkflowNodeChannelCloseEventTrigger:
			log.Debug().Msgf("Channel Close Event Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			if workflowStageExitConfigurationCache[workflowNode.Stage] == nil {
				workflowStageExitConfigurationCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
			}
			workflowStageExitConfigurationCache[workflowNode.Stage][commons.WorkflowParameterLabelChannelEventTriggered] = inputs[commons.WorkflowParameterLabelChannelEventTriggered]
		case commons.WorkflowNodeStageTrigger:
			if iteration > 0 {
				// There shouldn't be any stage nodes except when it's the first node
				return outputs, commons.Deleted, nil
			}
			if workflowNode.Stage > 1 {
				for label, value := range workflowStageExitConfigurationCache[workflowNode.Stage-1] {
					inputs[label] = value
				}
			}
		case commons.WorkflowNodeCronTrigger:
			log.Debug().Msg("Cron Trigger actually fired")
		}
		workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = commons.Active
		for childLinkId, childNode := range workflowNode.ChildNodes {
			if activeOutputs != nil && !slices.Contains(activeOutputs, workflowNode.LinkDetails[childLinkId].ParentOutput) {
				continue
			}
			// If there is no entry in the workflowNodeStagingParametersCache map for the child node's workflow version node ID, initialize an empty map
			if workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] == nil {
				workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId] = make(map[commons.WorkflowParameterLabel]string)
			}
			childInput := workflowNode.LinkDetails[childLinkId].ChildInput
			parentOutput := workflowNode.LinkDetails[childLinkId].ParentOutput
			if childInput != "" && parentOutput != "" && outputs[parentOutput] != "" {
				workflowNodeStagingParametersCache[childNode.WorkflowVersionNodeId][childInput] = outputs[parentOutput]
			}
			// Call ProcessWorkflowNode with several arguments, including childNode, workflowNode.WorkflowVersionNodeId, and workflowNodeStagingParametersCache
			childOutputs, childProcessingStatus, err := ProcessWorkflowNode(ctx, db, *childNode, workflowNode.WorkflowVersionNodeId,
				workflowNodeCache, workflowNodeStatus, workflowNodeStagingParametersCache, reference, outputs, iteration,
				workflowStageExitConfigurationCache, lightningRequestChannel, rebalanceRequestChannel)
			// If childProcessingStatus is not equal to commons.Pending, call AddWorkflowVersionNodeLog with several arguments, including nodeSettings.NodeId, reference, workflowNode.WorkflowVersionNodeId, triggeringWorkflowVersionNodeId, inputs, and childOutputs
			if childProcessingStatus != commons.Pending {
				AddWorkflowVersionNodeLog(db, reference,
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

func hasRoutingPolicyChanges(nodeId int, channelId int, routingPolicySettings ChannelPolicyConfiguration) bool {
	channelStateSettings := commons.GetChannelState(nodeId, channelId, true)
	if routingPolicySettings.FeeBaseMsat != nil &&
		*routingPolicySettings.FeeBaseMsat != channelStateSettings.LocalFeeBaseMsat {
		return true
	}
	if routingPolicySettings.FeeRateMilliMsat != nil &&
		*routingPolicySettings.FeeRateMilliMsat != channelStateSettings.LocalFeeRateMilliMsat {
		return true
	}
	if routingPolicySettings.MinHtlcMsat != nil &&
		*routingPolicySettings.MinHtlcMsat != channelStateSettings.LocalMinHtlcMsat {
		return true
	}
	if routingPolicySettings.MaxHtlcMsat != nil &&
		*routingPolicySettings.MaxHtlcMsat != channelStateSettings.LocalMaxHtlcMsat {
		return true
	}
	if routingPolicySettings.TimeLockDelta != nil &&
		*routingPolicySettings.TimeLockDelta != channelStateSettings.LocalTimeLockDelta {
		return true
	}
	return false
}

func getLinkedChannelIds(inputs map[commons.WorkflowParameterLabel]string, workflowNode WorkflowNode) ([]int, error) {
	var linkedChannelIds []int
	linkedChannelIdsString, exists := inputs[commons.WorkflowParameterLabelChannels]
	if !exists {
		return nil, errors.New(fmt.Sprintf("Finding channel parameter for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId))
	}
	err := json.Unmarshal([]byte(linkedChannelIdsString), &linkedChannelIds)
	if err != nil {
		return nil, errors.Wrapf(err, "Unmarshal the parent channelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIdsString, workflowNode.WorkflowVersionNodeId)
	}
	return linkedChannelIds, nil
}

func filterChannelIds(params FilterClauses, linkedChannels []channels.ChannelBody) []int {
	var filteredChannelIds []int
	filteredChannels := ApplyFilters(params, StructToMap(linkedChannels))
	for _, filteredChannel := range filteredChannels {
		channel, ok := filteredChannel.(map[string]interface{})
		if ok {
			filteredChannelIds = append(filteredChannelIds, channel["channelid"].(int))
		}
	}
	return filteredChannelIds
}

func AddWorkflowVersionNodeLog(db *sqlx.DB,
	reference string,
	workflowVersionNodeId int,
	triggeringWorkflowVersionNodeId int,
	inputs map[commons.WorkflowParameterLabel]string,
	outputs map[commons.WorkflowParameterLabel]string,
	workflowError error) {

	workflowVersionNodeLog := WorkflowVersionNodeLog{
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
