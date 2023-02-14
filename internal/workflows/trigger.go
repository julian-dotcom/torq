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

func ProcessWorkflow(ctx context.Context, db *sqlx.DB,
	workflowTriggerNode WorkflowNode,
	reference string,
	events []any,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest) error {

	// workflowNodeInputCache: map[workflowVersionNodeId][label]value
	workflowNodeInputCache := make(map[int]map[commons.WorkflowParameterLabel]string)
	// workflowNodeInputByChannelCache: map[workflowVersionNodeId][channelId][label]value
	workflowNodeInputByReferenceIdCache := make(map[int]map[int]map[commons.WorkflowParameterLabel]string)
	// workflowNodeOutputCache: map[workflowVersionNodeId][label]value
	workflowNodeOutputCache := make(map[int]map[commons.WorkflowParameterLabel]string)
	// workflowNodeOutputByChannelCache: map[workflowVersionNodeId][channelId][label]value
	workflowNodeOutputByReferenceIdCache := make(map[int]map[int]map[commons.WorkflowParameterLabel]string)
	// workflowStageOutputCache: map[stage][label]value
	workflowStageOutputCache := make(map[int]map[commons.WorkflowParameterLabel]string)
	// workflowStageOutputByChannelCache: map[stage][channelId][label]value
	workflowStageOutputByReferenceIdCache := make(map[int]map[int]map[commons.WorkflowParameterLabel]string)

	select {
	case <-ctx.Done():
		return errors.New(fmt.Sprintf("Context terminated for WorkflowVersionId: %v", workflowTriggerNode.WorkflowVersionId))
	default:
	}

	if workflowTriggerNode.Status != Active {
		return nil
	}

	workflowNodeStatus := make(map[int]commons.Status)
	workflowNodeStatus[workflowTriggerNode.WorkflowVersionNodeId] = commons.Active

	var eventChannelIds []int
	for _, event := range events {
		channelBalanceEvent, ok := event.(commons.ChannelBalanceEvent)
		if ok {
			eventChannelIds = append(eventChannelIds, channelBalanceEvent.ChannelId)
		}
		channelEvent, ok := event.(commons.ChannelEvent)
		if ok {
			eventChannelIds = append(eventChannelIds, channelEvent.ChannelId)
		}
	}
	marshalledChannelIdsFromEvents, err := json.Marshal(eventChannelIds)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal eventChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}

	var allChannelIds []int
	torqNodeIds := commons.GetAllTorqNodeIds()
	for _, torqNodeId := range torqNodeIds {
		// Force Response because we don't care about balance accuracy
		channelIdsByNode := commons.GetChannelStateChannelIds(torqNodeId, true)
		allChannelIds = append(allChannelIds, channelIdsByNode...)
	}
	marshalledAllChannelIds, err := json.Marshal(allChannelIds)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal allChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}

	workflowVersionNodes, err := GetWorkflowVersionNodesByStage(db, workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.Stage)
	if err != nil {
		return errors.Wrapf(err, "Failed to obtain workflow nodes for WorkflowVersionId: %v (stage: %v)",
			workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.Stage)
	}
	initializeInputCache(workflowVersionNodes, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
		workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
		allChannelIds, eventChannelIds, marshalledChannelIdsFromEvents, marshalledAllChannelIds, workflowTriggerNode,
		workflowStageOutputCache, workflowStageOutputByReferenceIdCache)

	switch workflowTriggerNode.Type {
	case commons.WorkflowNodeIntervalTrigger:
		log.Debug().Msgf("Interval Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	case commons.WorkflowNodeCronTrigger:
		log.Debug().Msgf("Cron Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	case commons.WorkflowNodeChannelBalanceEventTrigger:
		log.Debug().Msgf("Channel Balance Event Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowTriggerNode.WorkflowVersionNodeId][commons.WorkflowParameterLabelChannels] = string(marshalledChannelIdsFromEvents)
	case commons.WorkflowNodeChannelOpenEventTrigger:
		log.Debug().Msgf("Channel Open Event Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowTriggerNode.WorkflowVersionNodeId][commons.WorkflowParameterLabelChannels] = string(marshalledChannelIdsFromEvents)
	case commons.WorkflowNodeChannelCloseEventTrigger:
		log.Debug().Msgf("Channel Close Event Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowTriggerNode.WorkflowVersionNodeId][commons.WorkflowParameterLabelChannels] = string(marshalledChannelIdsFromEvents)
	case commons.WorkflowTrigger:
		fallthrough
	case commons.WorkflowNodeManualTrigger:
		log.Debug().Msgf("Manual Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}

	done := false
	iteration := 0
	var processStatus commons.Status
	for !done {
		iteration++
		if iteration > 100 {
			return errors.New(fmt.Sprintf("Infinite loop for WorkflowVersionId: %v", workflowTriggerNode.WorkflowVersionId))
		}
		done = true
		for _, workflowVersionNode := range workflowVersionNodes {
			processStatus, err = processWorkflowNode(ctx, db, workflowVersionNode, workflowTriggerNode,
				workflowNodeStatus, reference, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
				workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
				workflowStageOutputCache, workflowStageOutputByReferenceIdCache,
				lightningRequestChannel, rebalanceRequestChannel)
			if err != nil {
				return errors.Wrapf(err, "Failed to process workflow nodes for WorkflowVersionId: %v (stage: %v)",
					workflowTriggerNode.WorkflowVersionId, workflowTriggerNode.Stage)
			}
			if processStatus == commons.Active {
				done = false
			}
		}
	}

	workflowStageTriggerNodes, err := GetActiveSortedStageTriggerNodeForWorkflowVersionId(db,
		workflowTriggerNode.WorkflowVersionId)
	if err != nil {
		return errors.Wrapf(err, "Failed to obtain stage workflow trigger nodes for WorkflowVersionId: %v",
			workflowTriggerNode.WorkflowVersionId)
	}

	for _, workflowStageTriggerNode := range workflowStageTriggerNodes {
		// workflowNodeInputCache: map[workflowVersionNodeId][label]value
		workflowNodeInputCache = make(map[int]map[commons.WorkflowParameterLabel]string)
		// workflowNodeInputByReferenceIdCache: map[workflowVersionNodeId][channelId][label]value
		workflowNodeInputByReferenceIdCache = make(map[int]map[int]map[commons.WorkflowParameterLabel]string)
		// workflowNodeOutputCache: map[workflowVersionNodeId][label]value
		workflowNodeOutputCache = make(map[int]map[commons.WorkflowParameterLabel]string)
		// workflowNodeOutputByReferenceIdCache: map[workflowVersionNodeId][channelId][label]value
		workflowNodeOutputByReferenceIdCache = make(map[int]map[int]map[commons.WorkflowParameterLabel]string)

		workflowVersionNodes, err = GetWorkflowVersionNodesByStage(db, workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
		if err != nil {
			return errors.Wrapf(err, "Failed to obtain workflow nodes for WorkflowVersionId: %v (stage: %v)",
				workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
		}

		initializeInputCache(workflowVersionNodes, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
			workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
			allChannelIds, eventChannelIds, marshalledChannelIdsFromEvents, marshalledAllChannelIds, workflowStageTriggerNode,
			workflowStageOutputCache, workflowStageOutputByReferenceIdCache)
		done = false
		iteration = 0
		for !done {
			iteration++
			if iteration > 100 {
				return errors.New(fmt.Sprintf("Infinite loop for WorkflowVersionId: %v", workflowStageTriggerNode.WorkflowVersionId))
			}
			done = true
			for _, workflowVersionNode := range workflowVersionNodes {
				processStatus, err = processWorkflowNode(ctx, db, workflowVersionNode, workflowTriggerNode,
					workflowNodeStatus, reference, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
					workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
					workflowStageOutputCache, workflowStageOutputByReferenceIdCache,
					lightningRequestChannel, rebalanceRequestChannel)
				if err != nil {
					return errors.Wrapf(err, "Failed to process workflow nodes for WorkflowVersionId: %v (stage: %v)",
						workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
				}
				if processStatus == commons.Active {
					done = false
				}
			}
		}
	}
	return nil
}

func initializeInputCache(workflowVersionNodes []WorkflowNode,
	workflowNodeInputCache map[int]map[commons.WorkflowParameterLabel]string,
	workflowNodeInputByReferenceIdCache map[int]map[int]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputCache map[int]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputByReferenceIdCache map[int]map[int]map[commons.WorkflowParameterLabel]string,
	allChannelIds []int,
	eventChannelIds []int,
	marshalledChannelIdsFromEvents []byte,
	marshalledAllChannelIds []byte,
	workflowStageTriggerNode WorkflowNode,
	workflowStageOutputCache map[int]map[commons.WorkflowParameterLabel]string,
	workflowStageOutputByReferenceIdCache map[int]map[int]map[commons.WorkflowParameterLabel]string) {

	for _, workflowVersionNode := range workflowVersionNodes {
		if workflowNodeInputCache[workflowVersionNode.WorkflowVersionNodeId] == nil {
			workflowNodeInputCache[workflowVersionNode.WorkflowVersionNodeId] = make(map[commons.WorkflowParameterLabel]string)
		}
		if workflowNodeOutputCache[workflowVersionNode.WorkflowVersionNodeId] == nil {
			workflowNodeOutputCache[workflowVersionNode.WorkflowVersionNodeId] = make(map[commons.WorkflowParameterLabel]string)
		}
		if len(eventChannelIds) != 0 {
			workflowNodeInputCache[workflowVersionNode.WorkflowVersionNodeId][commons.WorkflowParameterLabelEventChannels] = string(marshalledChannelIdsFromEvents)
		}
		workflowNodeInputCache[workflowVersionNode.WorkflowVersionNodeId][commons.WorkflowParameterLabelAllChannels] = string(marshalledAllChannelIds)

		if workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId] == nil {
			workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId] = make(map[int]map[commons.WorkflowParameterLabel]string)
		}
		if workflowNodeOutputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId] == nil {
			workflowNodeOutputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId] = make(map[int]map[commons.WorkflowParameterLabel]string)
		}
		for _, channelId := range allChannelIds {
			if workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId] == nil {
				workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId] = make(map[commons.WorkflowParameterLabel]string)
			}
			if workflowNodeOutputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId] == nil {
				workflowNodeOutputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId] = make(map[commons.WorkflowParameterLabel]string)
			}
			if len(eventChannelIds) != 0 {
				workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId][commons.WorkflowParameterLabelEventChannels] = string(marshalledChannelIdsFromEvents)
			}
			workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId][commons.WorkflowParameterLabelAllChannels] = string(marshalledAllChannelIds)
		}
		if workflowStageTriggerNode.Stage > 0 {
			for label, value := range workflowStageOutputCache[workflowStageTriggerNode.Stage-1] {
				workflowNodeInputCache[workflowVersionNode.WorkflowVersionNodeId][label] = value
			}
			for channelId, labelValueMap := range workflowStageOutputByReferenceIdCache[workflowStageTriggerNode.Stage-1] {
				if workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId] == nil {
					workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId] = make(map[commons.WorkflowParameterLabel]string)
				}
				for label, value := range labelValueMap {
					workflowNodeInputByReferenceIdCache[workflowVersionNode.WorkflowVersionNodeId][channelId][label] = value
				}
			}
		}
	}
}

// processWorkflowNode
// workflowNodeInputCache: map[workflowVersionNodeId][channelId][label]value
// workflowNodeOutputCache: map[workflowVersionNodeId][channelId][label]value
// workflowStageExitConfigurationCache: map[stage][channelId][label]value
func processWorkflowNode(ctx context.Context, db *sqlx.DB,
	workflowNode WorkflowNode,
	workflowTriggerNode WorkflowNode,
	workflowNodeStatus map[int]commons.Status,
	reference string,
	workflowNodeInputCache map[int]map[commons.WorkflowParameterLabel]string,
	workflowNodeInputByReferenceIdCache map[int]map[int]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputCache map[int]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputByReferenceIdCache map[int]map[int]map[commons.WorkflowParameterLabel]string,
	workflowStageOutputCache map[int]map[commons.WorkflowParameterLabel]string,
	workflowStageOutputByReferenceIdCache map[int]map[int]map[commons.WorkflowParameterLabel]string,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest) (commons.Status, error) {

	select {
	case <-ctx.Done():
		return commons.Inactive, errors.New(fmt.Sprintf("Context terminated for WorkflowVersionId: %v", workflowNode.WorkflowVersionId))
	default:
	}
	var err error

	if workflowNode.Status != Active {
		return commons.Deleted, nil
	}

	status, exists := workflowNodeStatus[workflowNode.WorkflowVersionNodeId]
	if exists && status == commons.Active {
		// When the node is in the cache and active then it's already been processed successfully
		return commons.Deleted, nil
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
		return commons.Pending, nil
	}

	inputs := workflowNodeInputCache[workflowNode.WorkflowVersionNodeId]
	inputsByReferenceId := workflowNodeInputByReferenceIdCache[workflowNode.WorkflowVersionNodeId]
	outputs := workflowNodeOutputCache[workflowNode.WorkflowVersionNodeId]
	outputsByReferenceId := workflowNodeOutputByReferenceIdCache[workflowNode.WorkflowVersionNodeId]

	for parentWorkflowNodeLinkId, parentWorkflowNode := range workflowNode.ParentNodes {
		parentLink := workflowNode.LinkDetails[parentWorkflowNodeLinkId]
		inputs[parentLink.ChildInput] = workflowNodeOutputCache[parentWorkflowNode.WorkflowVersionNodeId][parentLink.ParentOutput]
		parentOutputsByReferenceIds := workflowNodeOutputByReferenceIdCache[parentWorkflowNode.WorkflowVersionNodeId]
		for referencId, labelValueMap := range parentOutputsByReferenceIds {
			inputsByReferenceId[referencId][parentLink.ChildInput] = labelValueMap[parentLink.ParentOutput]
		}
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
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var params FilterClauses
		err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &params)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var filteredChannelIds []int
		if len(linkedChannelIds) > 0 {
			if params.Filter.FuncName != "" || len(params.Or) != 0 || len(params.And) != 0 {
				var linkedChannels []channels.ChannelBody
				torqNodeIds := commons.GetAllTorqNodeIds()
				for _, torqNodeId := range torqNodeIds {
					linkedChannelsByNode, err := channels.GetChannelsByIds(db, torqNodeId, linkedChannelIds)
					if err != nil {
						return commons.Inactive, errors.Wrapf(err, "Getting the linked channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
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
			return commons.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case commons.WorkflowNodeAddTag, commons.WorkflowNodeRemoveTag:
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) != 0 {
			err = addOrRemoveTags(db, linkedChannelIds, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Adding or removing tags with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
		}
	case commons.WorkflowNodeChannelPolicyConfigurator:
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		for _, channelId := range linkedChannelIds {
			var channelPolicyInputConfiguration ChannelPolicyConfiguration
			channelPolicyInputConfiguration, err = processRoutingPolicyConfigurator(channelId, inputsByReferenceId, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			marshalledChannelPolicyConfiguration, err := json.Marshal(channelPolicyInputConfiguration)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			outputsByReferenceId[channelId][commons.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)
		}
	case commons.WorkflowNodeChannelPolicyAutoRun:
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		for _, channelId := range linkedChannelIds {
			var routingPolicySettings ChannelPolicyConfiguration
			routingPolicySettings, err = processRoutingPolicyConfigurator(channelId, inputsByReferenceId, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			var response commons.RoutingPolicyUpdateResponse
			response, err = processRoutingPolicyRun(routingPolicySettings, lightningRequestChannel, workflowNode, reference, workflowTriggerNode.Type)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			marshalledChannelPolicyConfiguration, err := json.Marshal(routingPolicySettings)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			outputsByReferenceId[channelId][commons.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)

			marshalledResponse, err := json.Marshal(response)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Response with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputsByReferenceId[channelId][commons.WorkflowParameterLabelStatus] = string(marshalledResponse)
		}
	case commons.WorkflowNodeChannelPolicyRun:
		for channelId, labelValueMap := range inputsByReferenceId {
			routingPolicySettingsString, exists := labelValueMap[commons.WorkflowParameterLabelRoutingPolicySettings]
			if !exists {
				continue
			}
			var routingPolicySettings ChannelPolicyConfiguration
			err = json.Unmarshal([]byte(routingPolicySettingsString), &routingPolicySettings)
			if err != nil {
				continue
			}

			if routingPolicySettings.ChannelId != 0 {
				var response commons.RoutingPolicyUpdateResponse
				response, err = processRoutingPolicyRun(routingPolicySettings, lightningRequestChannel, workflowNode, reference, workflowTriggerNode.Type)
				if err != nil {
					return commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				marshalledResponse, err := json.Marshal(response)
				if err != nil {
					return commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Response for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				// TODO FIXME create a more uniform status object
				outputsByReferenceId[channelId][commons.WorkflowParameterLabelStatus] = string(marshalledResponse)
			}
		}
	case commons.WorkflowNodeRebalanceConfigurator:
		incomingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelIncomingChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		outgoingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelOutgoingChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(incomingChannelIds) != 0 && len(outgoingChannelIds) != 0 {
			rebalanceConfiguration, err := processRebalanceConfigurator(incomingChannelIds, outgoingChannelIds, inputs, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			outputs[commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
		}
	case commons.WorkflowNodeRebalanceAutoRun:
		incomingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelIncomingChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		outgoingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelOutgoingChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(incomingChannelIds) != 0 && len(outgoingChannelIds) != 0 {
			rebalanceConfiguration, err := processRebalanceConfigurator(incomingChannelIds, outgoingChannelIds, inputs, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			var responses []commons.RebalanceResponse
			responses, err = processRebalanceRun(rebalanceConfiguration, rebalanceRequestChannel, workflowNode, reference)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			outputs[commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)

			marshalledResponses, err := json.Marshal(responses)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
		}
	case commons.WorkflowNodeRebalanceRun:
		var rebalanceConfiguration RebalanceConfiguration
		err = json.Unmarshal([]byte(workflowNode.Parameters.([]uint8)), &rebalanceConfiguration)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(rebalanceConfiguration.IncomingChannelIds) != 0 && len(rebalanceConfiguration.OutgoingChannelIds) != 0 {
			var responses []commons.RebalanceResponse
			responses, err = processRebalanceRun(rebalanceConfiguration, rebalanceRequestChannel, workflowNode, reference)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			outputs[commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)

			marshalledResponses, err := json.Marshal(responses)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
		}
	}
	workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = commons.Active

	if workflowStageOutputCache[workflowNode.Stage] == nil {
		workflowStageOutputCache[workflowNode.Stage] = make(map[commons.WorkflowParameterLabel]string)
	}
	for label, value := range outputs {
		workflowStageOutputCache[workflowNode.Stage][label] = value
	}

	if workflowStageOutputByReferenceIdCache[workflowNode.Stage] == nil {
		workflowStageOutputByReferenceIdCache[workflowNode.Stage] = make(map[int]map[commons.WorkflowParameterLabel]string)
	}
	for channelId, labelValueMap := range outputsByReferenceId {
		for label, value := range labelValueMap {
			workflowStageOutputByReferenceIdCache[workflowNode.Stage][channelId][label] = value
		}
	}

	marshalledInputs, err := json.Marshal([]any{inputs, inputsByReferenceId})
	if err != nil {
		log.Error().Err(err).Msgf("Marshalling inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	marshalledOutputs, err := json.Marshal([]any{outputs, outputsByReferenceId})
	if err != nil {
		log.Error().Err(err).Msgf("Marshalling outputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	_, err = addWorkflowVersionNodeLog(db, WorkflowVersionNodeLog{
		TriggerReference:                reference,
		InputData:                       string(marshalledInputs),
		OutputData:                      string(marshalledOutputs),
		DebugData:                       "",
		ErrorData:                       "",
		WorkflowVersionNodeId:           workflowNode.WorkflowVersionNodeId,
		TriggeringWorkflowVersionNodeId: &workflowTriggerNode.WorkflowVersionId,
		CreatedOn:                       time.Now().UTC(),
	})
	if err != nil {
		log.Error().Err(err).Msgf("Storing log for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	return commons.Active, nil
}

func getChannelIds(inputs map[commons.WorkflowParameterLabel]string, label commons.WorkflowParameterLabel) ([]int, error) {
	channelIdsString, exists := inputs[label]
	if !exists {
		return nil, errors.New(fmt.Sprintf("Parse %v", label))
	}
	var channelIds []int
	err := json.Unmarshal([]byte(channelIdsString), &channelIds)
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
	channelId int,
	inputsByChannelId map[int]map[commons.WorkflowParameterLabel]string,
	workflowNode WorkflowNode) (ChannelPolicyConfiguration, error) {

	var channelPolicyInputConfiguration ChannelPolicyConfiguration
	channelPolicyInputConfigurationString, exists := inputsByChannelId[channelId][commons.WorkflowParameterLabelRoutingPolicySettings]
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
	channelPolicyInputConfiguration.ChannelId = channelId
	return channelPolicyInputConfiguration, nil
}

func processRoutingPolicyRun(
	routingPolicySettings ChannelPolicyConfiguration,
	lightningRequestChannel chan interface{},
	workflowNode WorkflowNode,
	reference string,
	triggerType commons.WorkflowNodeType) (commons.RoutingPolicyUpdateResponse, error) {

	now := time.Now()
	torqNodeIds := commons.GetAllTorqNodeIds()
	channelSettings := commons.GetChannelSettingByChannelId(routingPolicySettings.ChannelId)
	nodeId := channelSettings.FirstNodeId
	if !slices.Contains(torqNodeIds, nodeId) {
		nodeId = channelSettings.SecondNodeId
	}
	if !slices.Contains(torqNodeIds, nodeId) {
		return commons.RoutingPolicyUpdateResponse{}, errors.New(fmt.Sprintf("Routing policy update on unmanaged channel for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId))
	}
	routingPolicyUpdateRequest := commons.RoutingPolicyUpdateRequest{
		CommunicationRequest: commons.CommunicationRequest{
			RequestId:   reference,
			RequestTime: &now,
			NodeId:      nodeId,
		},
		ChannelId:        routingPolicySettings.ChannelId,
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
	return response, nil
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
