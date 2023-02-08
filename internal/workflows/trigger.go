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
		if !getWorkflowNodeInputsComplete(workflowNode, inputs) {
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
			}
			ba, err := json.Marshal(filteredChannelIds)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Marshal the filteredChannelIds: %v for WorkflowVersionNodeId: %v", filteredChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			outputs[commons.WorkflowParameterLabelChannels] = string(ba)
		case commons.WorkflowNodeAddTag, commons.WorkflowNodeRemoveTag:
			linkedChannelIds, err := getLinkedChannelIds(inputs, workflowNode)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Getting the Linked ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			if len(linkedChannelIds) != 0 {
				err = addOrRemoveTags(db, linkedChannelIds, workflowNode)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Adding or removing tags with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
				}
			}
		case commons.WorkflowNodeChannelPolicyConfigurator:
			linkedChannelIds, err := getLinkedChannelIds(inputs, workflowNode)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Getting the Linked ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			if len(linkedChannelIds) != 0 {
				var channelPolicyInputConfiguration ChannelPolicyConfiguration
				channelPolicyInputConfiguration, err = processRoutingPolicyConfigurator(linkedChannelIds, inputs, workflowNode)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
				}
				channelPolicyInputConfiguration.ChannelIds = linkedChannelIds

				marshalledChannelPolicyConfiguration, err := json.Marshal(channelPolicyInputConfiguration)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
				}
				outputs[commons.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)
			}
		case commons.WorkflowNodeChannelPolicyAutoRun:
			linkedChannelIds, err := getLinkedChannelIds(inputs, workflowNode)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Getting the Linked ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			var routingPolicySettings ChannelPolicyConfiguration
			routingPolicySettings, err = processRoutingPolicyConfigurator(linkedChannelIds, inputs, workflowNode)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			routingPolicySettings.ChannelIds = linkedChannelIds

			if len(linkedChannelIds) != 0 {
				var responses []commons.RoutingPolicyUpdateResponse
				responses, err = processRoutingPolicyRun(routingPolicySettings, lightningRequestChannel, workflowNode, reference)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
				}

				marshalledChannelPolicyConfiguration, err := json.Marshal(routingPolicySettings)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
				}
				outputs[commons.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)

				marshalledResponses, err := json.Marshal(responses)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Responses with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
				}
				// TODO FIXME create a more uniform status object
				outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
			}
		case commons.WorkflowNodeChannelPolicyRun:
			routingPolicySettingsString, exists := inputs[commons.WorkflowParameterLabelRoutingPolicySettings]
			if !exists {
				return nil, commons.Inactive, errors.Wrapf(err, "No Routing Policy Configuration found for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			var routingPolicySettings ChannelPolicyConfiguration
			err = json.Unmarshal([]byte(routingPolicySettingsString), &routingPolicySettings)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(routingPolicySettings.ChannelIds) != 0 {
				var responses []commons.RoutingPolicyUpdateResponse
				responses, err = processRoutingPolicyRun(routingPolicySettings, lightningRequestChannel, workflowNode, reference)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				marshalledResponses, err := json.Marshal(responses)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				// TODO FIXME create a more uniform status object
				outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
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
			rebalanceSettingsString, exists := inputs[commons.WorkflowParameterLabelRebalanceSettings]
			if !exists {
				return nil, commons.Inactive, errors.Wrapf(err, "No rebalance settings found for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
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
			//		log.Error().Err(errors.New(response.Error)).Msg("Channel Interval Trigger Fired")
			//	}
			//}
		case commons.WorkflowNodeIntervalTrigger:
			log.Debug().Msgf("Interval Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		case commons.WorkflowNodeCronTrigger:
			log.Debug().Msgf("Cron Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		case commons.WorkflowNodeChannelBalanceEventTrigger:
			log.Debug().Msgf("Channel Balance Event Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		case commons.WorkflowNodeChannelOpenEventTrigger:
			log.Debug().Msgf("Channel Open Event Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		case commons.WorkflowNodeChannelCloseEventTrigger:
			log.Debug().Msgf("Channel Close Event Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		case commons.WorkflowTrigger:
			fallthrough
		case commons.WorkflowNodeManualTrigger:
			log.Debug().Msgf("Manual Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		case commons.WorkflowNodeStageTrigger:
			if iteration > 1 {
				// There shouldn't be any stage nodes except when it's the first node
				return outputs, commons.Deleted, nil
			}
			if workflowNode.Stage > 1 {
				for label, value := range workflowStageExitConfigurationCache[workflowNode.Stage-1] {
					outputs[label] = value
				}
			}
		}
		workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = commons.Active
		for childLinkId, childNode := range workflowNode.ChildNodes {
			// If there is no entry in the workflowNodeStagingParametersCache map for the child node's workflow version node ID, initialize an empty map
			if workflowStageExitConfigurationCache[workflowNode.Stage] == nil {
				workflowStageExitConfigurationCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
			}
			childInput := workflowNode.LinkDetails[childLinkId].ChildInput
			parentOutput := workflowNode.LinkDetails[childLinkId].ParentOutput
			if outputs[parentOutput] != "" {
				workflowStageExitConfigurationCache[workflowNode.Stage][childInput] = outputs[parentOutput]
			} else if outputs[childInput] != "" {
				workflowStageExitConfigurationCache[workflowNode.Stage][childInput] = outputs[childInput]
			}
			// Call ProcessWorkflowNode with several arguments, including childNode, workflowNode.WorkflowVersionNodeId, and workflowNodeStagingParametersCache
			childOutputs, childProcessingStatus, err := ProcessWorkflowNode(ctx, db, *childNode, workflowNode.WorkflowVersionNodeId,
				workflowNodeCache, workflowNodeStatus, reference, outputs, iteration,
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

func processRoutingPolicyRun(
	routingPolicySettings ChannelPolicyConfiguration,
	lightningRequestChannel chan interface{},
	workflowNode WorkflowNode,
	reference string) ([]commons.RoutingPolicyUpdateResponse, error) {
	if lightningRequestChannel == nil {
		log.Info().Msgf("Routing policy update skipped because lightningRequestChannel is nil for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		return nil, nil
	}
	if routingPolicySettings.FeeBaseMsat == nil &&
		routingPolicySettings.FeeRateMilliMsat == nil &&
		routingPolicySettings.MaxHtlcMsat == nil &&
		routingPolicySettings.MinHtlcMsat == nil &&
		routingPolicySettings.TimeLockDelta == nil {
		log.Info().Msgf("Routing policy update skipped because no data was found for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		return nil, nil
	}

	now := time.Now()
	torqNodeIds := commons.GetAllTorqNodeIds()
	var responses []commons.RoutingPolicyUpdateResponse
	for index := range routingPolicySettings.ChannelIds {
		channelSettings := commons.GetChannelSettingByChannelId(routingPolicySettings.ChannelIds[index])
		nodeId := channelSettings.FirstNodeId
		if !slices.Contains(torqNodeIds, nodeId) {
			nodeId = channelSettings.SecondNodeId
		}
		if !slices.Contains(torqNodeIds, nodeId) {
			log.Info().Msgf("Routing policy update on unmanaged channel for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			continue
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
					log.Error().Err(errors.New(response.Error)).Msgf("Workflow Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				responses = append(responses, response)
			}
		}
	}
	return responses, nil
}

func processRoutingPolicyConfigurator(
	linkedChannelIds []int,
	inputs map[commons.WorkflowParameterLabel]string,
	workflowNode WorkflowNode) (ChannelPolicyConfiguration, error) {

	var channelPolicyInputConfiguration ChannelPolicyConfiguration
	channelPolicyInputConfigurationString, exists := inputs[commons.WorkflowParameterLabelRoutingPolicySettings]
	if exists && channelPolicyInputConfigurationString != "" && channelPolicyInputConfigurationString != "null" {
		err := json.Unmarshal([]byte(channelPolicyInputConfigurationString), &channelPolicyInputConfiguration)
		if err != nil {
			return ChannelPolicyConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	}

	var channelPolicyConfiguration ChannelPolicyConfiguration
	err := json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &channelPolicyConfiguration)
	if err != nil {
		return ChannelPolicyConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
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
	if channelPolicyConfiguration.TimeLockDelta != nil {
		channelPolicyInputConfiguration.TimeLockDelta = channelPolicyConfiguration.TimeLockDelta
	}
	channelPolicyInputConfiguration.ChannelIds = linkedChannelIds
	return channelPolicyInputConfiguration, nil
}

func addOrRemoveTags(db *sqlx.DB, linkedChannelIds []int, workflowNode WorkflowNode) error {
	var params TagParameters
	err := json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &params)
	if err != nil {
		return errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}

	torqNodeIds := commons.GetAllTorqNodeIds()
	var processedNodeIds []int
	for _, tagToDelete := range params.RemovedTags {
		for index := range linkedChannelIds {
			var tag tags.TagEntityRequest
			processedNodeIds, tag = getTagEntityRequest(linkedChannelIds[index], tagToDelete.Value, params, torqNodeIds, processedNodeIds)
			if tag.TagId == 0 {
				continue
			}
			err = tags.UntagEntity(db, tag)
			if err != nil {
				return errors.Wrapf(err, "Failed to remove the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagToDelete.Value)
			}
		}
	}

	processedNodeIds = []int{}
	for _, tagtoAdd := range params.AddedTags {
		for index := range linkedChannelIds {
			var tag tags.TagEntityRequest
			processedNodeIds, tag = getTagEntityRequest(linkedChannelIds[index], tagtoAdd.Value, params, torqNodeIds, processedNodeIds)
			if tag.TagId == 0 {
				continue
			}
			err = tags.TagEntity(db, tag)
			if err != nil {
				return errors.Wrapf(err, "Failed to add the tags for WorkflowVersionNodeId: %v tagIDd", workflowNode.WorkflowVersionNodeId, tagtoAdd.Value)
			}
		}
	}
	return nil
}

func getTagEntityRequest(channelId int, tagId int, params TagParameters, torqNodeIds []int, processedNodeIds []int) ([]int, tags.TagEntityRequest) {
	if params.ApplyTo == "nodes" {
		channelSettings := commons.GetChannelSettingByChannelId(channelId)
		nodeId := channelSettings.FirstNodeId
		if slices.Contains(torqNodeIds, nodeId) {
			nodeId = channelSettings.SecondNodeId
		}
		if slices.Contains(torqNodeIds, nodeId) {
			log.Info().Msgf("Both nodes are managed by Torq nodeIds: %v and %v", channelSettings.FirstNodeId, channelSettings.SecondNodeId)
			return processedNodeIds, tags.TagEntityRequest{}
		}
		if slices.Contains(processedNodeIds, nodeId) {
			return processedNodeIds, tags.TagEntityRequest{}
		}
		processedNodeIds = append(processedNodeIds, nodeId)
		return processedNodeIds, tags.TagEntityRequest{
			NodeId: &nodeId,
			TagId:  tagId,
		}
	} else {
		return processedNodeIds, tags.TagEntityRequest{
			ChannelId: &channelId,
			TagId:     tagId,
		}
	}
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
	workflowVersionNodeLog.InputData = "[]"
	if len(inputs) > 0 {
		marshalledInputs, err := json.Marshal(inputs)
		if err == nil {
			workflowVersionNodeLog.InputData = string(marshalledInputs)
		} else {
			log.Error().Err(err).Msgf("Failed to marshal inputs for WorkflowVersionNodeId: %v", workflowVersionNodeId)
		}
	}
	workflowVersionNodeLog.OutputData = "[]"
	if len(outputs) != 0 {
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
