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

type workflowVersionNodeIdInt int
type channelIdInt int
type stageInt int
type originIdInt int

func ProcessWorkflow(ctx context.Context, db *sqlx.DB,
	workflowTriggerNode WorkflowNode,
	reference string,
	events []any,
	lightningRequestChannel chan<- interface{},
	rebalanceRequestChannel chan<- commons.RebalanceRequests) error {

	workflowNodeInputCache := make(map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string)
	workflowNodeInputByReferenceIdCache := make(map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string)
	workflowNodeOutputCache := make(map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string)
	workflowNodeOutputByReferenceIdCache := make(map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string)
	workflowStageOutputCache := make(map[stageInt]map[commons.WorkflowParameterLabel]string)
	workflowStageOutputByReferenceIdCache := make(map[stageInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string)

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
	marshalledEventChannelIdsFromEvents, err := json.Marshal(eventChannelIds)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal eventChannelIds for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}
	marshalledEvents, err := json.Marshal(events)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal events for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	}

	var allChannelIds []int
	torqNodeIds := commons.GetAllTorqNodeIds()
	for _, torqNodeId := range torqNodeIds {
		// Force Response because we don't care about balance accuracy
		channelIdsByNode := commons.GetChannelStateNotSharedChannelIds(torqNodeId, true)
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
		allChannelIds, marshalledEventChannelIdsFromEvents, marshalledAllChannelIds, marshalledEvents,
		workflowTriggerNode, workflowStageOutputCache, workflowStageOutputByReferenceIdCache)

	switch workflowTriggerNode.Type {
	case commons.WorkflowNodeIntervalTrigger:
		log.Debug().Msgf("Interval Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	case commons.WorkflowNodeCronTrigger:
		log.Debug().Msgf("Cron Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
	case commons.WorkflowNodeChannelBalanceEventTrigger:
		log.Debug().Msgf("Channel Balance Event Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowVersionNodeIdInt(workflowTriggerNode.WorkflowVersionNodeId)][commons.WorkflowParameterLabelChannels] = string(marshalledEventChannelIdsFromEvents)
	case commons.WorkflowNodeChannelOpenEventTrigger:
		log.Debug().Msgf("Channel Open Event Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowVersionNodeIdInt(workflowTriggerNode.WorkflowVersionNodeId)][commons.WorkflowParameterLabelChannels] = string(marshalledEventChannelIdsFromEvents)
	case commons.WorkflowNodeChannelCloseEventTrigger:
		log.Debug().Msgf("Channel Close Event Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
		workflowNodeOutputCache[workflowVersionNodeIdInt(workflowTriggerNode.WorkflowVersionNodeId)][commons.WorkflowParameterLabelChannels] = string(marshalledEventChannelIdsFromEvents)
	case commons.WorkflowTrigger:
		log.Debug().Msgf("Trigger Fired for WorkflowVersionNodeId: %v", workflowTriggerNode.WorkflowVersionNodeId)
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
			processStatus, err = processWorkflowNode(ctx, db, workflowVersionNode, workflowVersionNodes, workflowTriggerNode,
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
		workflowNodeInputCache = make(map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string)
		workflowNodeInputByReferenceIdCache = make(map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string)
		workflowNodeOutputCache = make(map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string)
		workflowNodeOutputByReferenceIdCache = make(map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string)

		workflowVersionNodes, err = GetWorkflowVersionNodesByStage(db, workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
		if err != nil {
			return errors.Wrapf(err, "Failed to obtain workflow nodes for WorkflowVersionId: %v (stage: %v)",
				workflowTriggerNode.WorkflowVersionId, workflowStageTriggerNode.Stage)
		}

		initializeInputCache(workflowVersionNodes, workflowNodeInputCache, workflowNodeInputByReferenceIdCache,
			workflowNodeOutputCache, workflowNodeOutputByReferenceIdCache,
			allChannelIds, marshalledEventChannelIdsFromEvents, marshalledAllChannelIds, marshalledEvents,
			workflowStageTriggerNode, workflowStageOutputCache, workflowStageOutputByReferenceIdCache)
		done = false
		iteration = 0
		for !done {
			iteration++
			if iteration > 100 {
				return errors.New(fmt.Sprintf("Infinite loop for WorkflowVersionId: %v", workflowStageTriggerNode.WorkflowVersionId))
			}
			done = true
			for _, workflowVersionNode := range workflowVersionNodes {
				processStatus, err = processWorkflowNode(ctx, db, workflowVersionNode, workflowVersionNodes, workflowTriggerNode,
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
	workflowNodeInputCache map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string,
	workflowNodeInputByReferenceIdCache map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputCache map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputByReferenceIdCache map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string,
	allChannelIds []int,
	marshalledChannelIdsFromEvents []byte,
	marshalledAllChannelIds []byte,
	marshalledEvents []byte,
	workflowStageTriggerNode WorkflowNode,
	workflowStageOutputCache map[stageInt]map[commons.WorkflowParameterLabel]string,
	workflowStageOutputByReferenceIdCache map[stageInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string) {

	for _, workflowVersionNode := range workflowVersionNodes {
		if workflowNodeInputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeInputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] = make(map[commons.WorkflowParameterLabel]string)
		}
		if workflowNodeOutputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeOutputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] = make(map[commons.WorkflowParameterLabel]string)
		}
		workflowNodeInputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][commons.WorkflowParameterLabelEventChannels] = string(marshalledChannelIdsFromEvents)
		workflowNodeInputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][commons.WorkflowParameterLabelAllChannels] = string(marshalledAllChannelIds)
		workflowNodeInputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][commons.WorkflowParameterLabelEvents] = string(marshalledEvents)

		if workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] = make(map[channelIdInt]map[commons.WorkflowParameterLabel]string)
		}
		if workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] == nil {
			workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)] = make(map[channelIdInt]map[commons.WorkflowParameterLabel]string)
		}
		for _, channelId := range allChannelIds {
			if workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelIdInt(channelId)] == nil {
				workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelIdInt(channelId)] = make(map[commons.WorkflowParameterLabel]string)
			}
			if workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelIdInt(channelId)] == nil {
				workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelIdInt(channelId)] = make(map[commons.WorkflowParameterLabel]string)
			}
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelIdInt(channelId)][commons.WorkflowParameterLabelEventChannels] = string(marshalledChannelIdsFromEvents)
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelIdInt(channelId)][commons.WorkflowParameterLabelAllChannels] = string(marshalledAllChannelIds)
			workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelIdInt(channelId)][commons.WorkflowParameterLabelEvents] = string(marshalledEvents)
		}
		if workflowStageTriggerNode.Stage > 0 {
			for label, value := range workflowStageOutputCache[stageInt(workflowStageTriggerNode.Stage-1)] {
				workflowNodeInputCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][label] = value
			}
			for channelId, labelValueMap := range workflowStageOutputByReferenceIdCache[stageInt(workflowStageTriggerNode.Stage-1)] {
				if workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelId] == nil {
					workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelId] = make(map[commons.WorkflowParameterLabel]string)
				}
				for label, value := range labelValueMap {
					workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowVersionNode.WorkflowVersionNodeId)][channelId][label] = value
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
	workflowNodes []WorkflowNode,
	workflowTriggerNode WorkflowNode,
	workflowNodeStatus map[int]commons.Status,
	reference string,
	workflowNodeInputCache map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string,
	workflowNodeInputByReferenceIdCache map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputCache map[workflowVersionNodeIdInt]map[commons.WorkflowParameterLabel]string,
	workflowNodeOutputByReferenceIdCache map[workflowVersionNodeIdInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string,
	workflowStageOutputCache map[stageInt]map[commons.WorkflowParameterLabel]string,
	workflowStageOutputByReferenceIdCache map[stageInt]map[channelIdInt]map[commons.WorkflowParameterLabel]string,
	lightningRequestChannel chan<- interface{},
	rebalanceRequestChannel chan<- commons.RebalanceRequests) (commons.Status, error) {

	select {
	case <-ctx.Done():
		return commons.Inactive, errors.New(fmt.Sprintf("Context terminated for WorkflowVersionId: %v", workflowNode.WorkflowVersionId))
	default:
	}

	if workflowNode.Status != Active {
		return commons.Deleted, nil
	}

	status, statusExists := workflowNodeStatus[workflowNode.WorkflowVersionNodeId]
	if statusExists && status == commons.Active {
		// When the node is in the cache and active then it's already been processed successfully
		return commons.Deleted, nil
	}

	if commons.IsWorkflowNodeTypeGrouped(workflowNode.Type) {
		workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = commons.Active
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
			status, statusExists = workflowNodeStatus[parentWorkflowNode.WorkflowVersionNodeId]
			if statusExists && status == commons.Active {
				continue linkedInputLoop
			}
		}
		// Not all inputs are available yet
		return commons.Pending, nil
	}

	inputs := workflowNodeInputCache[workflowVersionNodeIdInt(workflowNode.WorkflowVersionNodeId)]
	inputsByReferenceId := workflowNodeInputByReferenceIdCache[workflowVersionNodeIdInt(workflowNode.WorkflowVersionNodeId)]
	outputs := workflowNodeOutputCache[workflowVersionNodeIdInt(workflowNode.WorkflowVersionNodeId)]
	outputsByReferenceId := workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdInt(workflowNode.WorkflowVersionNodeId)]

	for parentWorkflowNodeLinkId, parentWorkflowNode := range workflowNode.ParentNodes {
		parentLink := workflowNode.LinkDetails[parentWorkflowNodeLinkId]
		parentOutputValue, labelExists := workflowNodeOutputCache[workflowVersionNodeIdInt(parentWorkflowNode.WorkflowVersionNodeId)][parentLink.ParentOutput]
		if labelExists {
			inputs[parentLink.ChildInput] = parentOutputValue
			outputs[parentLink.ChildInput] = parentOutputValue
		}
		for referencId, labelValueMap := range workflowNodeOutputByReferenceIdCache[workflowVersionNodeIdInt(parentWorkflowNode.WorkflowVersionNodeId)] {
			parentOutputValueByReferenceId, labelByReferenceIdExists := labelValueMap[parentLink.ParentOutput]
			if labelByReferenceIdExists {
				inputsByReferenceId[referencId][parentLink.ChildInput] = parentOutputValueByReferenceId
				outputsByReferenceId[referencId][parentLink.ChildInput] = parentOutputValueByReferenceId
			}
			for _, workflowNodeParameterLabelEnforced := range commons.GetWorkflowParameterLabelsEnforced() {
				parentByReferenceId, parentByReferenceIdExists := labelValueMap[workflowNodeParameterLabelEnforced]
				if parentByReferenceIdExists {
					inputsByReferenceId[referencId][workflowNodeParameterLabelEnforced] = parentByReferenceId
					outputsByReferenceId[referencId][workflowNodeParameterLabelEnforced] = parentByReferenceId
				}
			}
		}
	}

	updateReferencIds := make(map[channelIdInt]bool)

	switch workflowNode.Type {
	case commons.WorkflowNodeDataSourceTorqChannels:
		var params TorqChannelsConfiguration
		err := json.Unmarshal([]byte(workflowNode.Parameters), &params)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var channelIds []int
		switch params.Source {
		case "all":
			channelIds, err = getChannelIds(inputs, commons.WorkflowParameterLabelAllChannels)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Obtaining allChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		case "event":
			channelIds, err = getChannelIds(inputs, commons.WorkflowParameterLabelEventChannels)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Obtaining eventChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		case "eventXorAll":
			channelIds, _ = getChannelIds(inputs, commons.WorkflowParameterLabelEventChannels)
			if len(channelIds) == 0 {
				channelIds, err = getChannelIds(inputs, commons.WorkflowParameterLabelAllChannels)
				if err != nil {
					return commons.Inactive, errors.Wrapf(err, "Obtaining allChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
			}
		}

		err = setChannelIds(outputs, commons.WorkflowParameterLabelChannels, channelIds)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Adding All ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
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
	case commons.WorkflowNodeEventFilter:
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var params EventFilterConfiguration
		err = json.Unmarshal([]byte(workflowNode.Parameters), &params)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var filteredChannelIds []int
		events, err := getChannelBalanceEvents(inputs)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(events) == 0 {
			if !params.IgnoreWhenEventless {
				return commons.Inactive, errors.Wrapf(err, "No event(s) to filter found for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			filteredChannelIds = linkedChannelIds
		} else {
			if params.FilterClauses.Filter.FuncName != "" || len(params.FilterClauses.Or) != 0 || len(params.FilterClauses.And) != 0 {
				filteredChannelIds = filterChannelBalanceEventChannelIds(params.FilterClauses, linkedChannelIds, events)
			} else {
				filteredChannelIds = linkedChannelIds
			}
		}

		err = setChannelIds(outputs, commons.WorkflowParameterLabelChannels, filteredChannelIds)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case commons.WorkflowNodeChannelFilter:
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var params FilterClauses
		err = json.Unmarshal([]byte(workflowNode.Parameters), &params)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parsing parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var filteredChannelIds []int
		if params.Filter.FuncName != "" || len(params.Or) != 0 || len(params.And) != 0 {
			var linkedChannels []channels.ChannelBody
			torqNodeIds := commons.GetAllTorqNodeIds()
			for _, torqNodeId := range torqNodeIds {
				linkedChannelsByNode, err := channels.GetChannelsByIds(torqNodeId, linkedChannelIds)
				if err != nil {
					return commons.Inactive, errors.Wrapf(err, "Getting the linked channels to filters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}
				linkedChannels = append(linkedChannels, linkedChannelsByNode...)
			}
			filteredChannelIds = FilterChannelBodyChannelIds(params, linkedChannels)
		} else {
			filteredChannelIds = linkedChannelIds
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

		if len(linkedChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		err = addOrRemoveTags(db, linkedChannelIds, workflowNode)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Adding or removing tags with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
		}
	case commons.WorkflowNodeChannelPolicyConfigurator:
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		for channelId, labelValueMap := range inputsByReferenceId {
			if !slices.Contains(linkedChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			var routingPolicySettings ChannelPolicyConfiguration
			routingPolicySettings, err = processRoutingPolicyConfigurator(channelId, inputsByReferenceId, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}

			marshalledChannelPolicyConfiguration, err := json.Marshal(routingPolicySettings)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Routing Policy Configurator with ChannelIds: %v for WorkflowVersionNodeId: %v", linkedChannelIds, workflowNode.WorkflowVersionNodeId)
			}
			outputsByReferenceId[channelId][commons.WorkflowParameterLabelRoutingPolicySettings] = string(marshalledChannelPolicyConfiguration)
			updateReferencIds[channelId] = true
		}

		err = setChannelIds(outputs, commons.WorkflowParameterLabelChannels, linkedChannelIds)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case commons.WorkflowNodeChannelPolicyAutoRun:
		linkedChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining linkedChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(linkedChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No ChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		for channelId, labelValueMap := range inputsByReferenceId {
			if !slices.Contains(linkedChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

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
			updateReferencIds[channelId] = true
		}

		err = setChannelIds(outputs, commons.WorkflowParameterLabelChannels, linkedChannelIds)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Adding ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	case commons.WorkflowNodeChannelPolicyRun:
		for channelId, labelValueMap := range inputsByReferenceId {
			routingPolicySettingsString, exists := labelValueMap[commons.WorkflowParameterLabelRoutingPolicySettings]
			if !exists {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			var routingPolicySettings ChannelPolicyConfiguration
			err := json.Unmarshal([]byte(routingPolicySettingsString), &routingPolicySettings)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Unmarshalling Routing Policy Configuration with for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
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
				updateReferencIds[channelId] = true
			}
		}
	case commons.WorkflowNodeRebalanceConfigurator:
		var rebalanceConfiguration RebalanceConfiguration
		err := json.Unmarshal([]byte(workflowNode.Parameters), &rebalanceConfiguration)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		incomingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelIncomingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
				return commons.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && len(incomingChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No IncomingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		outgoingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelOutgoingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
				return commons.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && len(outgoingChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No OutgoingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		unfocusedPath := createUnfocusedPath(workflowNode, rebalanceConfiguration, workflowNodes)

		for channelId, labelValueMap := range inputsByReferenceId {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && !slices.Contains(incomingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && !slices.Contains(outgoingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			rebalanceConfiguration, err = processRebalanceConfigurator(rebalanceConfiguration, channelId, incomingChannelIds, outgoingChannelIds, inputsByReferenceId, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			rebalanceConfiguration.WorkflowUnfocusedPath = unfocusedPath

			marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			outputsByReferenceId[channelId][commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
			updateReferencIds[channelId] = true
		}

		if incomingChannelIds != nil {
			err = setChannelIds(outputs, commons.WorkflowParameterLabelIncomingChannels, incomingChannelIds)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Adding Incoming ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if outgoingChannelIds != nil {
			err = setChannelIds(outputs, commons.WorkflowParameterLabelOutgoingChannels, outgoingChannelIds)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Adding Outgoing ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}
	case commons.WorkflowNodeRebalanceAutoRun:
		var rebalanceConfiguration RebalanceConfiguration
		err := json.Unmarshal([]byte(workflowNode.Parameters), &rebalanceConfiguration)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		incomingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelIncomingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
				return commons.Inactive, errors.Wrapf(err, "Obtaining incomingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && len(incomingChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No IncomingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		outgoingChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelOutgoingChannels)
		if err != nil {
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
				return commons.Inactive, errors.Wrapf(err, "Obtaining outgoingChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && len(outgoingChannelIds) == 0 {
			return commons.Inactive, errors.Wrapf(err, "No OutgoingChannelIds found in the inputs for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		unfocusedPath := createUnfocusedPath(workflowNode, rebalanceConfiguration, workflowNodes)

		var rebalanceConfigurations []RebalanceConfiguration
		for channelId, labelValueMap := range inputsByReferenceId {
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels && !slices.Contains(incomingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels && !slices.Contains(outgoingChannelIds, int(channelId)) {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			rebalanceConfiguration, err = processRebalanceConfigurator(rebalanceConfiguration, channelId, incomingChannelIds, outgoingChannelIds, inputsByReferenceId, workflowNode)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Processing Rebalance configurator for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if rebalanceConfiguration.AmountMsat == nil {
				continue
			}

			rebalanceConfiguration.WorkflowUnfocusedPath = unfocusedPath

			if len(rebalanceConfiguration.IncomingChannelIds) != 0 && len(rebalanceConfiguration.OutgoingChannelIds) != 0 {
				marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
				if err != nil {
					return commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				outputsByReferenceId[channelId][commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
				updateReferencIds[channelId] = true
				rebalanceConfigurations = append(rebalanceConfigurations, rebalanceConfiguration)
			}
		}

		eventChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelEventChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining eventChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		var responses []commons.RebalanceResponse
		responses, err = processRebalanceRun(eventChannelIds, rebalanceConfigurations, rebalanceRequestChannel, workflowNode, reference)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(rebalanceConfigurations) != 0 {
			marshalledResponses, err := json.Marshal(responses)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
		}

		if incomingChannelIds != nil {
			err = setChannelIds(outputs, commons.WorkflowParameterLabelIncomingChannels, incomingChannelIds)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Adding Incoming ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}

		if outgoingChannelIds != nil {
			err = setChannelIds(outputs, commons.WorkflowParameterLabelOutgoingChannels, outgoingChannelIds)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Adding Outgoing ChannelIds to the output for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
		}
	case commons.WorkflowNodeRebalanceRun:
		var rebalanceConfigurations []RebalanceConfiguration
		for channelId, labelValueMap := range inputsByReferenceId {
			rebalanceConfigurationString, exists := labelValueMap[commons.WorkflowParameterLabelRebalanceSettings]
			if !exists {
				outputsByReferenceId[channelId] = labelValueMap
				continue
			}

			var rebalanceConfiguration RebalanceConfiguration
			err := json.Unmarshal([]byte(rebalanceConfigurationString), &rebalanceConfiguration)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Unmarshalling Rebalance Configuration with for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}

			if len(rebalanceConfiguration.IncomingChannelIds) != 0 && len(rebalanceConfiguration.OutgoingChannelIds) != 0 {
				marshalledRebalanceConfiguration, err := json.Marshal(rebalanceConfiguration)
				if err != nil {
					return commons.Inactive, errors.Wrapf(err, "Marshalling parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
				}

				outputsByReferenceId[channelId][commons.WorkflowParameterLabelRebalanceSettings] = string(marshalledRebalanceConfiguration)
				updateReferencIds[channelId] = true
				rebalanceConfigurations = append(rebalanceConfigurations, rebalanceConfiguration)
			}
		}

		eventChannelIds, err := getChannelIds(inputs, commons.WorkflowParameterLabelEventChannels)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Obtaining eventChannelIds for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		responses, err := processRebalanceRun(eventChannelIds, rebalanceConfigurations, rebalanceRequestChannel, workflowNode, reference)
		if err != nil {
			return commons.Inactive, errors.Wrapf(err, "Processing Rebalance for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}

		if len(rebalanceConfigurations) != 0 {
			marshalledResponses, err := json.Marshal(responses)
			if err != nil {
				return commons.Inactive, errors.Wrapf(err, "Marshalling Rebalance Responses for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
			}
			// TODO FIXME create a more uniform status object
			outputs[commons.WorkflowParameterLabelStatus] = string(marshalledResponses)
		}
	}
	workflowNodeStatus[workflowNode.WorkflowVersionNodeId] = commons.Active

	if workflowStageOutputCache[stageInt(workflowNode.Stage)] == nil {
		workflowStageOutputCache[stageInt(workflowNode.Stage)] = make(map[commons.WorkflowParameterLabel]string)
	}
	for label, value := range outputs {
		workflowStageOutputCache[stageInt(workflowNode.Stage)][label] = value
	}

	if workflowStageOutputByReferenceIdCache[stageInt(workflowNode.Stage)] == nil {
		workflowStageOutputByReferenceIdCache[stageInt(workflowNode.Stage)] = make(map[channelIdInt]map[commons.WorkflowParameterLabel]string)
	}
	for channelId, labelValueMap := range outputsByReferenceId {
		if workflowStageOutputByReferenceIdCache[stageInt(workflowNode.Stage)][channelId] == nil {
			workflowStageOutputByReferenceIdCache[stageInt(workflowNode.Stage)][channelId] = make(map[commons.WorkflowParameterLabel]string)
		}
		if updateReferencIds[channelId] {
			for label, value := range labelValueMap {
				workflowStageOutputByReferenceIdCache[stageInt(workflowNode.Stage)][channelId][label] = value
			}
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
		TriggeringWorkflowVersionNodeId: &workflowTriggerNode.WorkflowVersionNodeId,
		CreatedOn:                       time.Now().UTC(),
	})
	if err != nil {
		log.Error().Err(err).Msgf("Storing log for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	return commons.Active, nil
}

func createUnfocusedPath(
	workflowNode WorkflowNode,
	rebalanceConfiguration RebalanceConfiguration,
	workflowNodes []WorkflowNode) []WorkflowNode {

	var reversedPath []WorkflowNode
	childWorkflowNode := workflowNode
	goodPath := false
	for {
		var labelsOrdered []commons.WorkflowParameterLabel
		switch childWorkflowNode.Type {
		case commons.WorkflowNodeRebalanceAutoRun, commons.WorkflowNodeRebalanceConfigurator:
			if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
				labelsOrdered = []commons.WorkflowParameterLabel{
					commons.WorkflowParameterLabelIncomingChannels,
					commons.WorkflowParameterLabelOutgoingChannels,
				}
			}
			if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
				labelsOrdered = []commons.WorkflowParameterLabel{
					commons.WorkflowParameterLabelOutgoingChannels,
					commons.WorkflowParameterLabelIncomingChannels,
				}
			}
		default:
			labelsOrdered = []commons.WorkflowParameterLabel{
				commons.WorkflowParameterLabelChannels,
			}
		}
		label, parentWorkflowNode := getParent(childWorkflowNode, workflowNodes, labelsOrdered)
		if label == "" {
			break
		}

		childWorkflowNode = parentWorkflowNode

		if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels &&
			label == commons.WorkflowParameterLabelIncomingChannels {
			goodPath = true
		} else if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels &&
			label == commons.WorkflowParameterLabelOutgoingChannels {
			goodPath = true
		} else if !goodPath {
			continue
		}
		reversedPath = append(reversedPath, WorkflowNode{
			WorkflowVersionNodeId: parentWorkflowNode.WorkflowVersionNodeId,
			Name:                  parentWorkflowNode.Name,
			Status:                parentWorkflowNode.Status,
			Stage:                 parentWorkflowNode.Stage,
			Type:                  parentWorkflowNode.Type,
			Parameters:            parentWorkflowNode.Parameters,
			WorkflowVersionId:     parentWorkflowNode.WorkflowVersionId,
		})
	}

	// change direction
	for i, j := 0, len(reversedPath)-1; i < j; i, j = i+1, j-1 {
		reversedPath[i], reversedPath[j] = reversedPath[j], reversedPath[i]
	}
	return reversedPath
}

func getParent(
	workflowNode WorkflowNode,
	workflowNodes []WorkflowNode,
	labelsOrdered []commons.WorkflowParameterLabel) (commons.WorkflowParameterLabel, WorkflowNode) {

	for _, label := range labelsOrdered {
		for _, link := range workflowNode.LinkDetails {
			if link.ChildWorkflowVersionNodeId == workflowNode.WorkflowVersionNodeId &&
				link.ChildInput == label {
				for _, parentWorkflowNode := range workflowNodes {
					if parentWorkflowNode.WorkflowVersionNodeId == link.ParentWorkflowVersionNodeId {
						return label, parentWorkflowNode
					}
				}
			}
		}
	}
	return "", WorkflowNode{}
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

func getChannelBalanceEvents(inputs map[commons.WorkflowParameterLabel]string) ([]commons.ChannelBalanceEvent, error) {
	channelBalanceEventsString, exists := inputs[commons.WorkflowParameterLabelEvents]
	if !exists {
		return nil, errors.New(fmt.Sprintf("Parse %v", commons.WorkflowParameterLabelEvents))
	}
	var channelBalanceEvents []commons.ChannelBalanceEvent
	err := json.Unmarshal([]byte(channelBalanceEventsString), &channelBalanceEvents)
	if err != nil {
		return nil, errors.Wrapf(err, "Unmarshalling  %v", commons.WorkflowParameterLabelEvents)
	}
	if len(channelBalanceEvents) == 1 && channelBalanceEvents[0].ChannelId == 0 {
		return nil, nil
	}
	return channelBalanceEvents, nil
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
	rebalanceConfiguration RebalanceConfiguration,
	channelId channelIdInt,
	incomingChannelIds []int,
	outgoingChannelIds []int,
	inputsByReferenceId map[channelIdInt]map[commons.WorkflowParameterLabel]string,
	workflowNode WorkflowNode) (RebalanceConfiguration, error) {

	var rebalanceInputConfiguration RebalanceConfiguration
	rebalanceInputConfigurationString, exists := inputsByReferenceId[channelId][commons.WorkflowParameterLabelRebalanceSettings]
	if exists && rebalanceInputConfigurationString != "" && rebalanceInputConfigurationString != "null" {
		err := json.Unmarshal([]byte(rebalanceInputConfigurationString), &rebalanceInputConfiguration)
		if err != nil {
			return RebalanceConfiguration{}, errors.Wrapf(err, "Parse parameters for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
		}
	}

	if rebalanceInputConfiguration.Focus != "" && rebalanceInputConfiguration.Focus != rebalanceConfiguration.Focus {
		return RebalanceConfiguration{}, errors.New(fmt.Sprintf("RebalanceConfiguration has mismatching focus for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId))
	}
	rebalanceInputConfiguration.Focus = rebalanceConfiguration.Focus

	if rebalanceConfiguration.AmountMsat != nil {
		rebalanceInputConfiguration.AmountMsat = rebalanceConfiguration.AmountMsat
	}
	if rebalanceConfiguration.MaximumCostMilliMsat != nil {
		rebalanceInputConfiguration.MaximumCostMilliMsat = rebalanceConfiguration.MaximumCostMilliMsat
		rebalanceInputConfiguration.MaximumCostMsat = nil
	}
	if rebalanceConfiguration.MaximumCostMsat != nil {
		rebalanceInputConfiguration.MaximumCostMilliMsat = nil
		rebalanceInputConfiguration.MaximumCostMsat = rebalanceConfiguration.MaximumCostMsat
	}
	if rebalanceConfiguration.Focus == RebalancerFocusIncomingChannels {
		rebalanceInputConfiguration.IncomingChannelIds = []int{int(channelId)}
		if len(outgoingChannelIds) != 0 {
			rebalanceInputConfiguration.OutgoingChannelIds = outgoingChannelIds
		}
	}
	if rebalanceConfiguration.Focus == RebalancerFocusOutgoingChannels {
		if len(incomingChannelIds) != 0 {
			rebalanceInputConfiguration.IncomingChannelIds = incomingChannelIds
		}
		rebalanceInputConfiguration.OutgoingChannelIds = []int{int(channelId)}
	}
	return rebalanceInputConfiguration, nil
}

func processRebalanceRun(
	eventChannelIds []int,
	rebalanceSettings []RebalanceConfiguration,
	rebalanceRequestChannel chan<- commons.RebalanceRequests,
	workflowNode WorkflowNode,
	reference string) ([]commons.RebalanceResponse, error) {

	if rebalanceRequestChannel == nil {
		return nil, errors.New("Lightning request channel is nil")
	}

	now := time.Now()
	requestsMap := make(map[int]*commons.RebalanceRequests)
	for _, rebalanceSetting := range rebalanceSettings {
		var maxCostMsat uint64
		if rebalanceSetting.MaximumCostMsat != nil {
			maxCostMsat = *rebalanceSetting.MaximumCostMsat
		}
		if rebalanceSetting.MaximumCostMilliMsat != nil {
			maxCostMsat = uint64(*rebalanceSetting.MaximumCostMilliMsat) * (*rebalanceSetting.AmountMsat) / 1_000_000
		}

		workflowUnfocusedPathMarshalled, err := json.Marshal(rebalanceSetting.WorkflowUnfocusedPath)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Marshalling WorkflowUnfocusedPath of the rebalanceSetting for incomingChannelIds: %v, outgoingChannelIds: %v",
				rebalanceSetting.IncomingChannelIds, rebalanceSetting.OutgoingChannelIds)
		}

		if len(rebalanceSetting.WorkflowUnfocusedPath) != 0 {
			if rebalanceSetting.Focus == RebalancerFocusIncomingChannels {
				for _, incomingChannelId := range rebalanceSetting.IncomingChannelIds {
					var channelIds []int
					for _, outgoingChannelId := range rebalanceSetting.OutgoingChannelIds {
						if outgoingChannelId != 0 {
							channelIds = append(channelIds, outgoingChannelId)
						}
					}
					if len(channelIds) > 0 {
						channelSetting := commons.GetChannelSettingByChannelId(incomingChannelId)
						nodeId := channelSetting.FirstNodeId
						if !slices.Contains(commons.GetAllTorqNodeIds(), nodeId) {
							nodeId = channelSetting.SecondNodeId
						}
						_, exists := requestsMap[nodeId]
						if !exists {
							requestsMap[nodeId] = &commons.RebalanceRequests{
								CommunicationRequest: commons.CommunicationRequest{
									RequestId:   reference,
									RequestTime: &now,
									NodeId:      nodeId,
								},
							}
						}
						requestsMap[nodeId].Requests = append(requestsMap[nodeId].Requests, commons.RebalanceRequest{
							Origin:                commons.RebalanceRequestWorkflowNode,
							OriginId:              workflowNode.WorkflowVersionNodeId,
							OriginReference:       reference,
							IncomingChannelId:     incomingChannelId,
							OutgoingChannelId:     0,
							ChannelIds:            channelIds,
							AmountMsat:            *rebalanceSetting.AmountMsat,
							MaximumCostMsat:       maxCostMsat,
							MaximumConcurrency:    1,
							WorkflowUnfocusedPath: string(workflowUnfocusedPathMarshalled),
						})
					}
				}
			}
			if rebalanceSetting.Focus == RebalancerFocusOutgoingChannels {
				for _, outgoingChannelId := range rebalanceSetting.OutgoingChannelIds {
					var channelIds []int
					for _, incomingChannelId := range rebalanceSetting.IncomingChannelIds {
						if incomingChannelId != 0 {
							channelIds = append(channelIds, incomingChannelId)
						}
					}
					if len(channelIds) > 0 {
						channelSetting := commons.GetChannelSettingByChannelId(outgoingChannelId)
						nodeId := channelSetting.FirstNodeId
						if !slices.Contains(commons.GetAllTorqNodeIds(), nodeId) {
							nodeId = channelSetting.SecondNodeId
						}
						_, exists := requestsMap[nodeId]
						if !exists {
							requestsMap[nodeId] = &commons.RebalanceRequests{
								CommunicationRequest: commons.CommunicationRequest{
									RequestId:   reference,
									RequestTime: &now,
									NodeId:      nodeId,
								},
							}
						}
						requestsMap[nodeId].Requests = append(requestsMap[nodeId].Requests, commons.RebalanceRequest{
							Origin:                commons.RebalanceRequestWorkflowNode,
							OriginId:              workflowNode.WorkflowVersionNodeId,
							OriginReference:       reference,
							IncomingChannelId:     0,
							OutgoingChannelId:     outgoingChannelId,
							ChannelIds:            channelIds,
							AmountMsat:            *rebalanceSetting.AmountMsat,
							MaximumCostMsat:       maxCostMsat,
							MaximumConcurrency:    1,
							WorkflowUnfocusedPath: string(workflowUnfocusedPathMarshalled),
						})
					}
				}
			}
		}
	}
	var activeChannelIds []int
	var responses []commons.RebalanceResponse
	for nodeId := range requestsMap {
		if commons.RunningServices[commons.RebalanceService].GetStatus(nodeId) != commons.ServiceActive {
			return nil, errors.New(fmt.Sprintf("Lightning communication service is not active for nodeId: %v", nodeId))
		}
		responseChannel := make(chan []commons.RebalanceResponse)
		requestsMap[nodeId].ResponseChannel = responseChannel
		requests := *requestsMap[nodeId]
		for _, req := range requests.Requests {
			if req.IncomingChannelId != 0 {
				activeChannelIds = append(activeChannelIds, req.IncomingChannelId)
			}
			if req.OutgoingChannelId != 0 {
				activeChannelIds = append(activeChannelIds, req.OutgoingChannelId)
			}
		}
		rebalanceRequestChannel <- requests
		responses = append(responses, <-responseChannel...)
	}
	if len(eventChannelIds) == 0 {
		cancelRebalancersExcept(commons.RebalanceRequestWorkflowNode, workflowNode.WorkflowVersionNodeId,
			activeChannelIds)
	} else {
		for _, eventChannelId := range eventChannelIds {
			if !slices.Contains(activeChannelIds, eventChannelId) {
				cancelRebalancer(commons.RebalanceRequestWorkflowNode, workflowNode.WorkflowVersionNodeId,
					eventChannelId)
			}
		}
	}
	return responses, nil
}

func processRoutingPolicyConfigurator(
	channelId channelIdInt,
	inputsByChannelId map[channelIdInt]map[commons.WorkflowParameterLabel]string,
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
	err := json.Unmarshal([]byte(workflowNode.Parameters), &channelPolicyConfiguration)
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
	channelPolicyInputConfiguration.ChannelId = int(channelId)
	return channelPolicyInputConfiguration, nil
}

func processRoutingPolicyRun(
	routingPolicySettings ChannelPolicyConfiguration,
	lightningRequestChannel chan<- interface{},
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
	response := channels.SetRoutingPolicy(routingPolicyUpdateRequest, lightningRequestChannel)
	if response.Error != "" {
		log.Error().Err(errors.New(response.Error)).Msgf("Workflow Trigger Fired for WorkflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	return response, nil
}

func addOrRemoveTags(db *sqlx.DB, linkedChannelIds []int, workflowNode WorkflowNode) error {
	var params TagParameters
	err := json.Unmarshal([]byte(workflowNode.Parameters), &params)
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
			tag.CreatedByWorkflowVersionNodeId = &workflowNode.WorkflowVersionNodeId
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

func filterChannelBalanceEventChannelIds(params FilterClauses, linkedChannelIds []int, events []commons.ChannelBalanceEvent) []int {
	filteredChannelIds := extractChannelIds(ApplyFilters(params, ChannelBalanceEventToMap(events)))
	var resultChannelIds []int
	for _, linkedChannelId := range linkedChannelIds {
		if slices.Contains(filteredChannelIds, linkedChannelId) {
			resultChannelIds = append(resultChannelIds, linkedChannelId)
		}
	}
	return resultChannelIds
}

func FilterChannelBodyChannelIds(params FilterClauses, linkedChannels []channels.ChannelBody) []int {
	filteredChannelIds := extractChannelIds(ApplyFilters(params, ChannelBodyToMap(linkedChannels)))
	log.Trace().Msgf("Filtering applied to %d of %d channels", len(filteredChannelIds), len(linkedChannels))
	return filteredChannelIds
}

func extractChannelIds(filteredChannels []interface{}) []int {
	var filteredChannelIds []int
	for _, filteredChannel := range filteredChannels {
		channel, ok := filteredChannel.(map[string]interface{})
		if ok {
			filteredChannelIds = append(filteredChannelIds, channel["channelid"].(int))
			log.Trace().Msgf("Filter applied to channelId: %v", channel["channelid"])
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

func ChannelBalanceEventToMap(structs []commons.ChannelBalanceEvent) []map[string]interface{} {
	var maps []map[string]interface{}
	for _, s := range structs {
		maps = AddStructToMap(maps, s)
	}
	return maps
}

func ChannelBodyToMap(structs []channels.ChannelBody) []map[string]interface{} {
	var maps []map[string]interface{}
	for _, s := range structs {
		maps = AddStructToMap(maps, s)
	}
	return maps
}

func AddStructToMap(maps []map[string]interface{}, data any) []map[string]interface{} {
	structValue := reflect.ValueOf(data)
	structType := reflect.TypeOf(data)
	mapValue := make(map[string]interface{})

	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		mapValue[strings.ToLower(field.Name)] = structValue.Field(i).Interface()
	}
	maps = append(maps, mapValue)
	return maps
}
