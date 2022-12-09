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

func GetWorkflowNodes() map[WorkflowNodeType]WorkflowNodeTypeParameters {
	return map[WorkflowNodeType]WorkflowNodeTypeParameters{
		WorkflowNodeTrigger: {
			WorkflowNodeType: WorkflowNodeTrigger,
			RequiredInputs:   map[string]WorkflowParameter{},
			OptionalInputs:   map[string]WorkflowParameter{},
			RequiredOutputs:  map[string]WorkflowParameter{"triggered": WorkflowParameterTriggered},
			OptionalOutputs:  map[string]WorkflowParameter{},
		},
		WorkflowNodeChannelFilter: {
			WorkflowNodeType: WorkflowNodeChannelFilter,
			RequiredInputs:   map[string]WorkflowParameter{},
			OptionalInputs: map[string]WorkflowParameter{
				"channels":  WorkflowParameterChannelIds,
				"triggered": WorkflowParameterTriggered,
			},
			RequiredOutputs: map[string]WorkflowParameter{"channels": WorkflowParameterChannelIds},
			OptionalOutputs: map[string]WorkflowParameter{"triggered": WorkflowParameterTriggered},
		},
		WorkflowNodeCostParameters: {
			WorkflowNodeType: WorkflowNodeCostParameters,
			RequiredInputs:   map[string]WorkflowParameter{"channels": WorkflowParameterChannelIds},
			OptionalInputs:   map[string]WorkflowParameter{"triggered": WorkflowParameterTriggered},
			RequiredOutputs:  map[string]WorkflowParameter{},
			OptionalOutputs: map[string]WorkflowParameter{
				"routingPolicySettings": WorkflowParameterRoutingPolicySettings,
				"channels":              WorkflowParameterChannelIds,
				"triggered":             WorkflowParameterTriggered,
			},
		},
		WorkflowNodeRebalanceParameters: {
			WorkflowNodeType: WorkflowNodeRebalanceParameters,
			RequiredInputs:   map[string]WorkflowParameter{"channels": WorkflowParameterChannelIds},
			OptionalInputs:   map[string]WorkflowParameter{"triggered": WorkflowParameterTriggered},
			RequiredOutputs:  map[string]WorkflowParameter{},
			OptionalOutputs: map[string]WorkflowParameter{
				"rebalanceSettings": WorkflowParameterRebalanceSettings,
				"channels":          WorkflowParameterChannelIds,
				"triggered":         WorkflowParameterTriggered,
			},
		},
		WorkflowNodeDeferredApply: {
			WorkflowNodeType: WorkflowNodeDeferredApply,
			RequiredInputs:   map[string]WorkflowParameter{"deferredData": WorkflowParameterDeferredData},
			OptionalInputs:   map[string]WorkflowParameter{},
			RequiredOutputs:  map[string]WorkflowParameter{},
			OptionalOutputs: map[string]WorkflowParameter{
				"channels":  WorkflowParameterChannelIds,
				"status":    WorkflowParameterStatus,
				"triggered": WorkflowParameterTriggered,
			},
		},
		WorkflowNodeRebalanceRun: {
			WorkflowNodeType: WorkflowNodeRebalanceRun,
			RequiredInputs: map[string]WorkflowParameter{
				"rebalanceSettings":   WorkflowParameterRebalanceSettings,
				"sourceChannels":      WorkflowParameterChannelIds,
				"destinationChannels": WorkflowParameterChannelIds,
			},
			OptionalInputs:  map[string]WorkflowParameter{"triggered": WorkflowParameterTriggered},
			RequiredOutputs: map[string]WorkflowParameter{},
			OptionalOutputs: map[string]WorkflowParameter{
				"sourceChannels":      WorkflowParameterChannelIds,
				"destinationChannels": WorkflowParameterChannelIds,
				"status":              WorkflowParameterStatus,
				"triggered":           WorkflowParameterTriggered,
			},
		},
		WorkflowNodeRoutingPolicyRun: {
			WorkflowNodeType: WorkflowNodeRoutingPolicyRun,
			RequiredInputs: map[string]WorkflowParameter{
				"routingPolicySettings": WorkflowParameterRoutingPolicySettings,
				"channels":              WorkflowParameterChannelIds,
			},
			OptionalInputs:  map[string]WorkflowParameter{"triggered": WorkflowParameterTriggered},
			RequiredOutputs: map[string]WorkflowParameter{},
			OptionalOutputs: map[string]WorkflowParameter{
				"sourceChannels":      WorkflowParameterChannelIds,
				"destinationChannels": WorkflowParameterChannelIds,
				"status":              WorkflowParameterStatus,
				"triggered":           WorkflowParameterTriggered,
			},
		},
		WorkflowNodeSetVariable: {
			WorkflowNodeType: WorkflowNodeSetVariable,
			RequiredInputs:   map[string]WorkflowParameter{},
			OptionalInputs:   map[string]WorkflowParameter{"any": WorkflowParameterAny},
			RequiredOutputs:  map[string]WorkflowParameter{},
			OptionalOutputs:  map[string]WorkflowParameter{"any": WorkflowParameterAny},
		},
		WorkflowNodeFilterOnVariable: {
			WorkflowNodeType: WorkflowNodeFilterOnVariable,
			RequiredInputs:   map[string]WorkflowParameter{},
			OptionalInputs:   map[string]WorkflowParameter{"any": WorkflowParameterAny},
			RequiredOutputs:  map[string]WorkflowParameter{},
			OptionalOutputs:  map[string]WorkflowParameter{"any": WorkflowParameterAny},
		},
	}
}
