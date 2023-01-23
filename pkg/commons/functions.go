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

func CopyParameters(parameters map[WorkflowParameterLabel]string) map[WorkflowParameterLabel]string {
	parametersCopy := make(map[WorkflowParameterLabel]string)
	for k, v := range parameters {
		parametersCopy[k] = v
	}
	return parametersCopy
}

func GetServiceTypes() []ServiceType {
	return []ServiceType{
		LndService,
		VectorService,
		AmbossService,
		TorqService,
		AutomationService,
		LightningCommunicationService,
		RebalanceService,
		MaintenanceService,
	}
}

func getDeltaPerMille(base uint64, amt uint64) int {
	if base > amt {
		return int((base - amt) / base * 1_000)
	} else if base == amt {
		return 0
	} else {
		return int((amt - base) / amt * 1_000)
	}
}

func GetVectorUrl(vectorUrl string, suffix string) string {
	return vectorUrl + suffix
}

func GetWorkflowNodes() map[WorkflowNodeType]WorkflowNodeTypeParameters {
	allTriggeredOnly := make(map[WorkflowParameterLabel]WorkflowParameterType)
	allTriggeredOnly[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	allTriggeredOnly[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	all := make(map[WorkflowParameterLabel]WorkflowParameterType)
	all[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	all[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered
	all[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	all[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings
	all[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings
	all[WorkflowParameterLabelSourceChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelDestinationChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus

	channelsOnly := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelsOnly[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds

	//WorkflowNodeTimeTrigger
	timeTriggerRequiredOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	timeTriggerRequiredOutputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered

	//WorkflowNodeChannelBalanceEventTrigger
	channelBalanceEventTriggerRequiredOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelBalanceEventTriggerRequiredOutputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	//WorkflowNodeChannelFilter
	channelFilterOptionalInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelFilterOptionalInputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelFilterOptionalInputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	channelFilterOptionalInputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered
	channelFilterRequiredOutputs := channelsOnly
	channelFilterOptionalOutputs := allTriggeredOnly

	//WorkflowNodeChannelPolicyConfigurator
	channelPolicyConfiguratorOptionalInputs := all
	channelPolicyConfiguratorOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyConfiguratorOptionalOutputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	channelPolicyConfiguratorOptionalOutputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	channelPolicyConfiguratorOptionalOutputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	//WorkflowNodeRebalanceParameters
	rebalanceParametersOptionalInputs := all
	rebalanceParametersOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceParametersOptionalOutputs[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings
	rebalanceParametersOptionalOutputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	rebalanceParametersOptionalOutputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	//WorkflowNodeAddTag
	addTagOptionalInputs := channelsOnly
	addTagOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	addTagOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	addTagOptionalOutputs[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings
	addTagOptionalOutputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	addTagOptionalOutputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	//WorkflowNodeRemoveTag
	removeTagOptionalInputs := channelsOnly
	removeTagOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	removeTagOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	removeTagOptionalOutputs[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings
	removeTagOptionalOutputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	removeTagOptionalOutputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	//WorkflowNodeStageTrigger
	stageTriggerOptionalOutputs := all

	//WorkflowNodeRebalanceRun
	rebalanceRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceRunRequiredInputs[WorkflowParameterLabelSourceChannels] = WorkflowParameterTypeChannelIds
	rebalanceRunRequiredInputs[WorkflowParameterLabelDestinationChannels] = WorkflowParameterTypeChannelIds
	rebalanceRunOptionalInputs := allTriggeredOnly
	rebalanceRunOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceRunOptionalOutputs[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings
	rebalanceRunOptionalOutputs[WorkflowParameterLabelSourceChannels] = WorkflowParameterTypeChannelIds
	rebalanceRunOptionalOutputs[WorkflowParameterLabelDestinationChannels] = WorkflowParameterTypeChannelIds
	rebalanceRunOptionalOutputs[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus
	rebalanceRunOptionalOutputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	rebalanceRunOptionalOutputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	//WorkflowNodeChannelPolicyRun
	channelPolicyRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyRunRequiredInputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	channelPolicyRunRequiredInputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyRunOptionalInputs := allTriggeredOnly
	channelPolicyRunOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyRunOptionalOutputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	channelPolicyRunOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyRunOptionalOutputs[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus
	channelPolicyRunOptionalOutputs[WorkflowParameterLabelTimeTriggered] = WorkflowParameterTypeTimeTriggered
	channelPolicyRunOptionalOutputs[WorkflowParameterLabelChannelEventTriggered] = WorkflowParameterTypeChannelEventTriggered

	//WorkflowNodeSetVariable
	setVariableOptionalInputs := all
	setVariableOptionalOutputs := all

	//WorkflowNodeFilterOnVariable
	filterOnVariableOptionalInputs := all
	filterOnVariableOptionalOutputs := all

	return map[WorkflowNodeType]WorkflowNodeTypeParameters{
		WorkflowNodeTimeTrigger: {
			WorkflowNodeType: WorkflowNodeTimeTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  timeTriggerRequiredOutputs,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelBalanceEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelBalanceEventTriggerRequiredOutputs,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelFilter: {
			WorkflowNodeType: WorkflowNodeChannelFilter,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   channelFilterOptionalInputs,
			RequiredOutputs:  channelFilterRequiredOutputs,
			OptionalOutputs:  channelFilterOptionalOutputs,
		},
		WorkflowNodeChannelPolicyConfigurator: {
			WorkflowNodeType: WorkflowNodeChannelPolicyConfigurator,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   channelPolicyConfiguratorOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  channelPolicyConfiguratorOptionalOutputs,
		},
		WorkflowNodeRebalanceParameters: {
			WorkflowNodeType: WorkflowNodeRebalanceParameters,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   rebalanceParametersOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  rebalanceParametersOptionalOutputs,
		},
		WorkflowNodeAddTag: {
			WorkflowNodeType: WorkflowNodeAddTag,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   addTagOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  addTagOptionalOutputs,
		},
		WorkflowNodeRemoveTag: {
			WorkflowNodeType: WorkflowNodeRemoveTag,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   removeTagOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  removeTagOptionalOutputs,
		},
		WorkflowNodeStageTrigger: {
			WorkflowNodeType: WorkflowNodeStageTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  stageTriggerOptionalOutputs,
		},
		WorkflowNodeRebalanceRun: {
			WorkflowNodeType: WorkflowNodeRebalanceRun,
			RequiredInputs:   rebalanceRunRequiredInputs,
			OptionalInputs:   rebalanceRunOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  rebalanceRunOptionalOutputs,
		},
		WorkflowNodeChannelPolicyRun: {
			WorkflowNodeType: WorkflowNodeChannelPolicyRun,
			RequiredInputs:   channelPolicyRunRequiredInputs,
			OptionalInputs:   channelPolicyRunOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  channelPolicyRunOptionalOutputs,
		},
		WorkflowNodeSetVariable: {
			WorkflowNodeType: WorkflowNodeSetVariable,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   setVariableOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  setVariableOptionalOutputs,
		},
		WorkflowNodeFilterOnVariable: {
			WorkflowNodeType: WorkflowNodeFilterOnVariable,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   filterOnVariableOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  filterOnVariableOptionalOutputs,
		},
	}
}
