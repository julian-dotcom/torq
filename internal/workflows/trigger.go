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

func Trigger(ctx context.Context, db *sqlx.DB, workflowTriggerNode WorkflowNode, reference string, eventChannel chan interface{}) error {
	return processWorkflowNode(ctx, db, workflowTriggerNode, 0, reference, make(map[string]string), eventChannel)
}

func processWorkflowNode(ctx context.Context, db *sqlx.DB, workflowNode WorkflowNode, triggeredWorkflowVersionNodeId int, reference string, inputs map[string]string, eventChannel chan interface{}) error {
	select {
	case <-ctx.Done():
		return errors.Newf("Context terminated for WorkflowVersionId: %v", workflowNode.WorkflowVersionId)
	default:
	}
	outputs := copyInputs(inputs)
	var err error
	var workflowNodeParameters WorkflowNodeParameters
	activeOutputIndex := -1
	if workflowNode.Status == commons.Active {
		workflowNodeParameters, err = getWorkflowNodeParameters(workflowNode)
		if err != nil {
			return errors.Wrapf(err, "Obtaining parameters for WorkflowVersionId: %v", workflowNode.WorkflowVersionId)
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

		case commons.WorkflowNodeDeferredApply:
		case commons.WorkflowNodeCostParameters:
		case commons.WorkflowNodeRebalanceParameters:
		case commons.WorkflowNodeRebalanceRun:
		case commons.WorkflowNodeRoutingPolicyRun:
		}
		for outputIndex, childNodeOutputArray := range workflowNode.ChildNodes {
			if activeOutputIndex != -1 && activeOutputIndex != outputIndex {
				continue
			}
			marhalledInputs, err := json.Marshal(inputs)
			if err != nil {
				return err
			}
			marhalledOutputs, err := json.Marshal(outputs)
			if err != nil {
				return err
			}
			_, err = AddWorkflowVersionNodeLog(db, WorkflowVersionNodeLog{
				TriggerReference:               reference,
				InputData:                      string(marhalledInputs),
				OutputData:                     string(marhalledOutputs),
				WorkflowVersionNodeId:          workflowNode.WorkflowVersionNodeId,
				TriggeredWorkflowVersionNodeId: triggeredWorkflowVersionNodeId,
			})
			if err != nil {
				log.Error().Err(err).Msgf("Failed to write a log entry for WorkflowVersionNodeId: %v, OutputIndex: %v", workflowNode.WorkflowVersionNodeId, outputIndex)
			}
			for _, childNode := range childNodeOutputArray {
				err := processWorkflowNode(ctx, db, *childNode, workflowNode.WorkflowVersionNodeId, reference, outputs, eventChannel)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func copyInputs(inputs map[string]string) map[string]string {
	inputsCopy := make(map[string]string)
	for k, v := range inputs {
		inputsCopy[k] = v
	}
	return inputsCopy
}
