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
	workflowNodeLinkId int,
	triggeringWorkflowVersionNodeId int,
	workflowNodeCache map[int]WorkflowNode,
	workflowNodeStatus map[int]commons.Status,
	reference string,
	inputs map[commons.WorkflowParameterLabel]string,
	stagedInputsByWorkflowVersionNodeId map[int]map[commons.WorkflowParameterLabel]string,
	iteration int,
	triggerType commons.WorkflowNodeType,
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
	var err error

	outputs := commons.CloneParameters(inputs)
	if workflowNode.Status == Active {
		status, exists := workflowNodeStatus[workflowNode.WorkflowVersionNodeId]
		if exists && status == commons.Active {
			// When the node is in the cache and active then it's already been processed successfully
			return nil, commons.Deleted, nil
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

		stagedInputs, exists := stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId]
		if exists {
			commons.CopyParameters(inputs, stagedInputs)
		}

		if workflowNodeLinkId != 0 &&
			workflowNode.LinkDetails[workflowNodeLinkId].ChildInput != workflowNode.LinkDetails[workflowNodeLinkId].ParentOutput {

			inputs[workflowNode.LinkDetails[workflowNodeLinkId].ChildInput] = inputs[workflowNode.LinkDetails[workflowNodeLinkId].ParentOutput]
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
			if stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId] == nil {
				stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId] = make(map[commons.WorkflowParameterLabel]string)
			}
			commons.CopyParameters(stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId], inputs)
			// Not all inputs are available yet
			return nil, commons.Pending, nil
		}

		if !getWorkflowNodeInputsComplete(workflowNode, inputs) {
			if stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId] == nil {
				stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId] = make(map[commons.WorkflowParameterLabel]string)
			}
			commons.CopyParameters(stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId], inputs)
			// When the node is pending then not all inputs are available yet
			return nil, commons.Pending, nil
		}

		_, exists = stagedInputsByWorkflowVersionNodeId[workflowNode.WorkflowVersionNodeId]
		if exists {
			delete(stagedInputsByWorkflowVersionNodeId, workflowNode.WorkflowVersionNodeId)
		}

		// Inputs might have been updated
		outputs = commons.CloneParameters(inputs)
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
			linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			var params FilterClauses
			err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &params)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			var filteredChannelIds []int
			if len(linkedChannelIds) > 0 {
				if params.Filter.FuncName != "" || len(params.Or) != 0 || len(params.And) != 0 {
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

			err = setChannelIds(outputs, commons.WorkflowParameterLabelChannels, filteredChannelIds)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		case commons.WorkflowNodeAddTag, commons.WorkflowNodeRemoveTag:
			linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(linkedChannelIds) != 0 {
				err = addOrRemoveTags(db, linkedChannelIds, workflowNode)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Adding or removing tags with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
				}
			}
		case commons.WorkflowNodeChannelPolicyConfigurator:
			linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
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
			linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			var routingPolicySettings ChannelPolicyConfiguration
			routingPolicySettings, err = processRoutingPolicyConfigurator(linkedChannelIds, inputs, workflowNode)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			routingPolicySettings.ChannelIds = linkedChannelIds

			if len(linkedChannelIds) != 0 {
				var responses []commons.RoutingPolicyUpdateResponse
				responses, err = processRoutingPolicyRun(routingPolicySettings, lightningRequestChannel, workflowNode, reference, triggerType)
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
				responses, err = processRoutingPolicyRun(routingPolicySettings, lightningRequestChannel, workflowNode, reference, triggerType)
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
			incomingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelIncomingChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			outgoingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelOutgoingChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(incomingChannelIds) != 0 && len(outgoingChannelIds) != 0 {
				rebalanceConfiguration, err := processRebalanceConfigurator(incomingChannelIds, outgoingChannelIds, inputs, workflowNode)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				outputs[commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
			}
		case commons.WorkflowNodeRebalanceAutoRun:
			incomingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelIncomingChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			outgoingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelOutgoingChannels)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(incomingChannelIds) != 0 && len(outgoingChannelIds) != 0 {
				rebalanceConfiguration, err := processRebalanceConfigurator(incomingChannelIds, outgoingChannelIds, inputs, workflowNode)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				var responses []commons.RebalanceResponse
				responses, err = processRebalanceRun(rebalanceConfiguration, rebalanceRequestChannel, workflowNode, reference)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				outputs[commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)

				marshalledResponses, err := json.Marshal(responses)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				// TODO FIXME create a more uniform status object
				outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
			}
		case commons.WorkflowNodeRebalanceRun:
			var rebalanceConfiguration RebalanceConfiguration
			err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &rebalanceConfiguration)
			if err != nil {
				return nil, commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(rebalanceConfiguration.IncomingChannelIds) != 0 && len(rebalanceConfiguration.OutgoingChannelIds) != 0 {
				var responses []commons.RebalanceResponse
				responses, err = processRebalanceRun(rebalanceConfiguration, rebalanceRequestChannel, workflowNode, reference)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				outputs[commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)

				marshalledResponses, err := json.Marshal(responses)
				if err != nil {
					return nil, commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				// TODO FIXME create a more uniform status object
				outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
			}
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
		if workflowStageExitConfigurationCache[workflowNode.Stage] == nil {
			workflowStageExitConfigurationCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
		}
		for label, value := range outputs {
			workflowStageExitConfigurationCache[workflowNode.Stage][label] = value
		}
		for linkId, childNode := range workflowNode.ChildNodes {
			// Call ProcessWorkflowNode with several arguments, including childNode, workflowNode.WorkflowVersionNodeId, and workflowNodeStagingParametersCache
			childOutputs, childProcessingStatus, err := ProcessWorkflowNode(ctx, db, *childNode, linkId, workflowNode.WorkflowVersionNodeId,
				workflowNodeCache, workflowNodeStatus, reference, outputs, stagedInputsByWorkflowVersionNodeId, iteration, triggerType,
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

func getChannelIds(inputs map[commons.WorkflowParameterLabel]string, label commons.WorkflowParameterLabel) ([]int, error) {
	outgoingChannelIdsString, exists := inputs[commons.WorkflowParameterLabelOutgoingChannels]
	if !exists {
		return nil, errors.New(fmt.Sprintf("Parse %v", label))
	}
	var channelIds []int
	err := json.Unmarshal([]byte(outgoingChannelIdsString), &channelIds)
	if err != nil {
		return nil, errors.Wrapf(err, "Unmarshalling  %v", label)
	}
	return channelIds, nil
}

func setChannelIds(outputs map[commons.WorkflowParameterLabel]string, label commons.WorkflowParameterLabel, channelIds []int) error {
	ba, err := json.Marshal(channelIds)
	if err != nil {
		return errors.Wrapf(err, "Marshal the channelIds: %v", channelIds)
	}
	outputs[label] = string(ba)
	return nil
}

func processRebalanceConfigurator(
	incomingChannelIds []int,
	outgoingChannelIds []int,
	inputs map[commons.WorkflowParameterLabel]string,
	workflowNode WorkflowNode) (RebalanceConfiguration, error) {

	var rebalanceInputConfiguration RebalanceConfiguration
	rebalanceInputConfigurationString, exists := inputs[commons.WorkflowParameterLabelRebalanceSettings]
	if exists && rebalanceInputConfigurationString != "" && rebalanceInputConfigurationString != "null" {
		err := json.Unmarshal([]byte(rebalanceInputConfigurationString), &rebalanceInputConfiguration)
		if err != nil {
			return RebalanceConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	}

	var rebalanceConfiguration RebalanceConfiguration
	err := json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &rebalanceConfiguration)
	if err != nil {
		return RebalanceConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	if rebalanceConfiguration.AmountMsat != nil {
		rebalanceInputConfiguration.AmountMsat = rebalanceConfiguration.AmountMsat
	}
	if rebalanceConfiguration.MaximumCostMilliMsat != nil {
		rebalanceInputConfiguration.MaximumCostMilliMsat = rebalanceConfiguration.MaximumCostMilliMsat
	}
	if rebalanceConfiguration.MaximumCostMsat != nil {
		rebalanceInputConfiguration.MaximumCostMsat = rebalanceConfiguration.MaximumCostMsat
	}
	rebalanceInputConfiguration.IncomingChannelIds = incomingChannelIds
	rebalanceInputConfiguration.OutgoingChannelIds = outgoingChannelIds
	return rebalanceInputConfiguration, nil
}

func processRebalanceRun(
	rebalanceSettings RebalanceConfiguration,
	rebalanceRequestChannel chan commons.RebalanceRequest,
	workflowNode WorkflowNode,
	reference string) ([]commons.RebalanceResponse, error) {

	var responses []commons.RebalanceResponse
	now := time.Now()
	if len(rebalanceSettings.OutgoingChannelIds) == 1 {
		outgoingChannelId := rebalanceSettings.OutgoingChannelIds[0]
		channelSetting := commons.GetChannelSettingByChannelId(outgoingChannelId)
		nodeId := channelSetting.FirstNodeId
		if !slices.Contains(commons.GetAllTorqNodeIds(), nodeId) {
			nodeId = channelSetting.SecondNodeId
		}
		request := commons.RebalanceRequest{
			CommunicationRequest: commons.CommunicationRequest{
				RequestId:   reference,
				RequestTime: &now,
				NodeId:      nodeId,
			},
			Origin:          commons.RebalanceRequestWorkflowNode,
			OriginId:        workflowNode.WorkflowVersionNodeId,
			OriginReference: reference,
			ChannelIds:      rebalanceSettings.IncomingChannelIds,
			AmountMsat:      *rebalanceSettings.AmountMsat,
			MaximumCostMsat: *rebalanceSettings.MaximumCostMsat,
			//MaximumConcurrency: *rebalanceSettings.MaximumConcurrency,
		}
		request.OutgoingChannelId = outgoingChannelId

		response := channels.SetRebalanceWithTimeout(request, rebalanceRequestChannel)
		if response.Error != "" {
			log.Error().Err(errors.New(response.Error)).Msgf("Workflow Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
		responses = append(responses, response)
	} else {
		for _, incomingChannelId := range rebalanceSettings.IncomingChannelIds {
			channelSetting := commons.GetChannelSettingByChannelId(incomingChannelId)
			nodeId := channelSetting.FirstNodeId
			if !slices.Contains(commons.GetAllTorqNodeIds(), nodeId) {
				nodeId = channelSetting.SecondNodeId
			}
			request := commons.RebalanceRequest{
				CommunicationRequest: commons.CommunicationRequest{
					RequestId:   reference,
					RequestTime: &now,
					NodeId:      nodeId,
				},
				Origin:          commons.RebalanceRequestWorkflowNode,
				OriginId:        workflowNode.WorkflowVersionNodeId,
				OriginReference: reference,
				ChannelIds:      rebalanceSettings.OutgoingChannelIds,
				AmountMsat:      *rebalanceSettings.AmountMsat,
				MaximumCostMsat: *rebalanceSettings.MaximumCostMsat,
				//MaximumConcurrency: *rebalanceSettings.MaximumConcurrency,
			}
			request.IncomingChannelId = incomingChannelId

			response := channels.SetRebalanceWithTimeout(request, rebalanceRequestChannel)
			if response.Error != "" {
				log.Error().Err(errors.New(response.Error)).Msgf("Workflow Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			responses = append(responses, response)
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

func processRoutingPolicyRun(
	routingPolicySettings ChannelPolicyConfiguration,
	lightningRequestChannel chan interface{},
	workflowNode WorkflowNode,
	reference string,
	triggerType commons.WorkflowNodeType) ([]commons.RoutingPolicyUpdateResponse, error) {

	now := time.Now()
	torqNodeIds := commons.GetAllTorqNodeIds()
	var responses []commons.RoutingPolicyUpdateResponse
	for _, channelId := range routingPolicySettings.ChannelIds {
		channelSettings := commons.GetChannelSettingByChannelId(channelId)
		nodeId := channelSettings.FirstNodeId
		if !slices.Contains(torqNodeIds, nodeId) {
			nodeId = channelSettings.SecondNodeId
		}
		if !slices.Contains(torqNodeIds, nodeId) {
			log.Info().Msgf("Routing policy update on unmanaged channel for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			continue
		}
		routingPolicyUpdateRequest := commons.RoutingPolicyUpdateRequest{
			CommunicationRequest: commons.CommunicationRequest{
				RequestId:   reference,
				RequestTime: &now,
				NodeId:      nodeId,
			},
			ChannelId:        channelId,
			FeeRateMilliMsat: routingPolicySettings.FeeRateMilliMsat,
			FeeBaseMsat:      routingPolicySettings.FeeBaseMsat,
			MaxHtlcMsat:      routingPolicySettings.MaxHtlcMsat,
			MinHtlcMsat:      routingPolicySettings.MinHtlcMsat,
			TimeLockDelta:    routingPolicySettings.TimeLockDelta,
		}
		if triggerType == commons.WorkflowNodeManualTrigger {
			// DISABLE rate limiter
			routingPolicyUpdateRequest.RateLimitSeconds = 1
			routingPolicyUpdateRequest.RateLimitCount = 10
		}
		response := channels.SetRoutingPolicyWithTimeout(routingPolicyUpdateRequest, lightningRequestChannel)
		if response.Error != "" {
			log.Error().Err(errors.New(response.Error)).Msgf("Workflow Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
		responses = append(responses, response)
	}
	return responses, nil
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

func filterChannelIds(params FilterClauses, linkedChannels []channels.ChannelBody) []int {
	var filteredChannelIds []int
	filteredChannels := ApplyFilters(params, StructToMap(linkedChannels))
	for _, filteredChannel := range filteredChannels {
		channel, ok := filteredChannel.(map[string]interface{})
		if ok {
			filteredChannelIds = append(filteredChannelIds, channel["channelid"].(int))
			log.Trace().Msgf("Filter applied to channelId: %v", channel["lndshortchannelid"])
		}
	}
	log.Debug().Msgf("Filtering applied to %d of %d channels", len(filteredChannelIds),len(linkedChannels))
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
