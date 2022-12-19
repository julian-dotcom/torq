package commons

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
)

// GetNetwork defaults to MainNet when no match is found
func GetNetwork(network string) Network {
	switch network {
	case "testnet":
		return TestNet
	case "signet":
		return SigNet
	case "simnet":
		return SimNet
	case "regtest":
		return RegTest
	}
	return MainNet
}

// GetChain defaults to Bitcoin when no match is found
func GetChain(chain string) Chain {
	switch chain {
	case "litecoin":
		return Litecoin
	}
	return Bitcoin
}

const mutexLocked = 1

func MutexLocked(m *sync.Mutex) bool {
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func RWMutexWriteLocked(rw *sync.RWMutex) bool {
	// RWMutex has a "w" sync.Mutex field for write lock
	state := reflect.ValueOf(rw).Elem().FieldByName("w").FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func RWMutexReadLocked(rw *sync.RWMutex) bool {
	return reflect.ValueOf(rw).Elem().FieldByName("readerCount").Int() > 0
}

func ConvertLNDShortChannelID(LNDShortChannelID uint64) string {
	blockHeight := uint32(LNDShortChannelID >> 40)
	txIndex := uint32(LNDShortChannelID>>16) & 0xFFFFFF
	outputIndex := uint16(LNDShortChannelID)
	return strconv.FormatUint(uint64(blockHeight), 10) +
		"x" + strconv.FormatUint(uint64(txIndex), 10) +
		"x" + strconv.FormatUint(uint64(outputIndex), 10)
}

func ConvertShortChannelIDToLND(ShortChannelID string) (uint64, error) {
	parts := strings.Split(ShortChannelID, "x")
	blockHeight, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, errors.Wrap(err, "Converting block height from string to int")
	}
	txIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, errors.Wrap(err, "Converting tx index from string to int")
	}
	txPosition, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, errors.Wrap(err, "Converting tx position from string to int")
	}

	return (uint64(blockHeight) << 40) |
		(uint64(txIndex) << 16) |
		(uint64(txPosition)), nil
}

func ParseChannelPoint(channelPoint string) (string, int) {
	parts := strings.Split(channelPoint, ":")
	if channelPoint != "" && strings.Contains(channelPoint, ":") && len(parts) == 2 {
		outputIndex, err := strconv.Atoi(parts[1])
		if err == nil {
			return parts[0], outputIndex
		}
		log.Debug().Err(err).Msgf("Failed to parse channelPoint %v", channelPoint)
	}
	return "", 0
}

func CreateChannelPoint(fundingTransactionHash string, fundingOutputIndex int) string {
	return fmt.Sprintf("%s:%v", fundingTransactionHash, fundingOutputIndex)
}

func CopyParameters(parameters map[string]string) map[string]string {
	parametersCopy := make(map[string]string)
	for k, v := range parameters {
		parametersCopy[k] = v
	}
	return parametersCopy
}

type WorkflowParameterWithLabel struct {
	Label string
	WorkflowParameter
}

func GetWorkflowNodes() map[WorkflowNodeType]WorkflowNodeTypeParameters {
	return map[WorkflowNodeType]WorkflowNodeTypeParameters{
		WorkflowNodeTimeTrigger: {
			WorkflowNodeType: WorkflowNodeTimeTrigger,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs:   []WorkflowParameterWithLabel{},
			RequiredOutputs:  []WorkflowParameterWithLabel{{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered}},
			OptionalOutputs:  []WorkflowParameterWithLabel{},
		},
		WorkflowNodeChannelBalanceEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs:   []WorkflowParameterWithLabel{},
			RequiredOutputs:  []WorkflowParameterWithLabel{{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered}},
			OptionalOutputs:  []WorkflowParameterWithLabel{},
		},
		WorkflowNodeChannelFilter: {
			WorkflowNodeType: WorkflowNodeChannelFilter,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs: []WorkflowParameterWithLabel{
				{Label: "channels", WorkflowParameter: WorkflowParameterChannelIds},
				{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered},
			},
			RequiredOutputs: []WorkflowParameterWithLabel{{Label: "channels", WorkflowParameter: WorkflowParameterChannelIds}},
			OptionalOutputs: []WorkflowParameterWithLabel{{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered}},
		},
		WorkflowNodeCostParameters: {
			WorkflowNodeType: WorkflowNodeCostParameters,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs:   []WorkflowParameterWithLabel{{Label: "any", WorkflowParameter: WorkflowParameterAny}},
			RequiredOutputs:  []WorkflowParameterWithLabel{},
			OptionalOutputs: []WorkflowParameterWithLabel{
				{Label: "routingPolicySettings", WorkflowParameter: WorkflowParameterRoutingPolicySettings},
				{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered},
			},
		},
		WorkflowNodeRebalanceParameters: {
			WorkflowNodeType: WorkflowNodeRebalanceParameters,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs:   []WorkflowParameterWithLabel{{Label: "any", WorkflowParameter: WorkflowParameterAny}},
			RequiredOutputs:  []WorkflowParameterWithLabel{},
			OptionalOutputs: []WorkflowParameterWithLabel{
				{Label: "rebalanceSettings", WorkflowParameter: WorkflowParameterRebalanceSettings},
				{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered},
			},
		},
		WorkflowNodeStageTrigger: {
			WorkflowNodeType: WorkflowNodeStageTrigger,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs:   []WorkflowParameterWithLabel{},
			RequiredOutputs:  []WorkflowParameterWithLabel{},
			OptionalOutputs:  []WorkflowParameterWithLabel{{Label: "any", WorkflowParameter: WorkflowParameterAny}},
		},
		WorkflowNodeRebalanceRun: {
			WorkflowNodeType: WorkflowNodeRebalanceRun,
			RequiredInputs: []WorkflowParameterWithLabel{
				{Label: "rebalanceSettings", WorkflowParameter: WorkflowParameterRoutingPolicySettings},
				{Label: "sourceChannels", WorkflowParameter: WorkflowParameterChannelIds},
				{Label: "destinationChannels", WorkflowParameter: WorkflowParameterChannelIds},
			},
			OptionalInputs:  []WorkflowParameterWithLabel{{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered}},
			RequiredOutputs: []WorkflowParameterWithLabel{},
			OptionalOutputs: []WorkflowParameterWithLabel{
				{Label: "rebalanceSettings", WorkflowParameter: WorkflowParameterRoutingPolicySettings},
				{Label: "sourceChannels", WorkflowParameter: WorkflowParameterChannelIds},
				{Label: "destinationChannels", WorkflowParameter: WorkflowParameterChannelIds},
				{Label: "status", WorkflowParameter: WorkflowParameterStatus},
				{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered},
			},
		},
		WorkflowNodeRoutingPolicyRun: {
			WorkflowNodeType: WorkflowNodeRoutingPolicyRun,
			RequiredInputs: []WorkflowParameterWithLabel{
				{Label: "routingPolicySettings", WorkflowParameter: WorkflowParameterRoutingPolicySettings},
				{Label: "channels", WorkflowParameter: WorkflowParameterChannelIds},
			},
			OptionalInputs:  []WorkflowParameterWithLabel{{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered}},
			RequiredOutputs: []WorkflowParameterWithLabel{},
			OptionalOutputs: []WorkflowParameterWithLabel{
				{Label: "routingPolicySettings", WorkflowParameter: WorkflowParameterRoutingPolicySettings},
				{Label: "channels", WorkflowParameter: WorkflowParameterChannelIds},
				{Label: "status", WorkflowParameter: WorkflowParameterStatus},
				{Label: "triggered", WorkflowParameter: WorkflowParameterTriggered},
			},
		},
		WorkflowNodeSetVariable: {
			WorkflowNodeType: WorkflowNodeSetVariable,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs:   []WorkflowParameterWithLabel{{Label: "any", WorkflowParameter: WorkflowParameterAny}},
			RequiredOutputs:  []WorkflowParameterWithLabel{},
			OptionalOutputs:  []WorkflowParameterWithLabel{{Label: "any", WorkflowParameter: WorkflowParameterAny}},
		},
		WorkflowNodeFilterOnVariable: {
			WorkflowNodeType: WorkflowNodeFilterOnVariable,
			RequiredInputs:   []WorkflowParameterWithLabel{},
			OptionalInputs:   []WorkflowParameterWithLabel{{Label: "any", WorkflowParameter: WorkflowParameterAny}},
			RequiredOutputs:  []WorkflowParameterWithLabel{},
			OptionalOutputs:  []WorkflowParameterWithLabel{{Label: "any", WorkflowParameter: WorkflowParameterAny}},
		},
	}
}
