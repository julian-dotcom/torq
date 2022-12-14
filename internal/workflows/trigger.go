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

func ProcessWorkflowNode(ctx context.Context, db *sqlx.DB, workflowNode WorkflowNode, triggeredWorkflowVersionNodeId int,
	reference string, inputs map[string]string, eventChannel chan interface{}, iteration int) (map[string]string, error) {

	iteration++
	if iteration > 100 {
		return nil, errors.Newf("Infinate loop for WorkflowVersionId: %v", workflowNode.WorkflowVersionId)
	}
	select {
	case <-ctx.Done():
		return nil, errors.Newf("Context terminated for WorkflowVersionId: %v", workflowNode.WorkflowVersionId)
	default:
	}
	outputs := copyInputs(inputs)
	var err error
	var workflowNodeParameters WorkflowNodeParameters
	activeOutputIndex := -1
	if workflowNode.Status == commons.Active {
		workflowNodeParameters, err = getWorkflowNodeParameters(workflowNode)
		if err != nil {
			return nil, errors.Wrapf(err, "Obtaining parameters for WorkflowVersionId: %v", workflowNode.WorkflowVersionId)
		}
		switch workflowNode.Type {
		case commons.WorkflowNodeSetVariable:
			variableName := getWorkflowNodeParameter(workflowNodeParameters, commons.WorkflowParameterVariableName).ValueString
			stringVariableParameter := getWorkflowNodeParameter(workflowNodeParameters, commons.WorkflowParameterVariableValueString)
			if stringVariableParameter.ValueString != "" {
				outputs[variableName] = stringVariableParameter.ValueString
			} else {
				outputs[variableName] = fmt.Sprintf("%d", getWorkflowNodeParameter(workflowNodeParameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
			}
		case commons.WorkflowNodeFilterOnVariable:
			variableName := getWorkflowNodeParameter(workflowNodeParameters, commons.WorkflowParameterVariableName).ValueString
			stringVariableParameter := getWorkflowNodeParameter(workflowNodeParameters, commons.WorkflowParameterVariableValueString)
			stringValue := ""
			if stringVariableParameter.ValueString != "" {
				stringValue = stringVariableParameter.ValueString
			} else {
				stringValue = fmt.Sprintf("%d", getWorkflowNodeParameter(workflowNodeParameters, commons.WorkflowParameterVariableValueNumber).ValueNumber)
			}
			if inputs[variableName] == stringValue {
				activeOutputIndex = 0
			} else {
				activeOutputIndex = 1
			}
		case commons.WorkflowNodeChannelFilter:

		case commons.WorkflowNodeDeferredLink:
		case commons.WorkflowNodeCostParameters:
		case commons.WorkflowNodeRebalanceParameters:
		case commons.WorkflowNodeRebalanceRun:
		case commons.WorkflowNodeRoutingPolicyRun:
		}
		for outputIndex, childNodeOutputArray := range workflowNode.ChildNodes {
			if activeOutputIndex != -1 && activeOutputIndex != outputIndex {
				continue
			}
			for _, childNode := range childNodeOutputArray {
				childOutputs, err := ProcessWorkflowNode(ctx, db, *childNode, workflowNode.WorkflowVersionNodeId, reference, outputs, eventChannel, iteration)
				AddWorkflowVersionNodeLog(db, reference, workflowNode.WorkflowVersionNodeId, inputs, childOutputs, err)
				if err != nil {
					return nil, err
				}
				for k, v := range childOutputs {
					outputs[k] = v
				}
			}
		}
	}
	return outputs, nil
}

func AddWorkflowVersionNodeLog(db *sqlx.DB, reference string, workflowVersionNodeId int,
	inputs map[string]string, outputs map[string]string, workflowError error) {

	workflowVersionNodeLog := WorkflowVersionNodeLog{
		WorkflowVersionNodeId: workflowVersionNodeId,
		TriggerReference:      reference,
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

func copyInputs(inputs map[string]string) map[string]string {
	inputsCopy := make(map[string]string)
	for k, v := range inputs {
		inputsCopy[k] = v
	}
	return inputsCopy
}
