package commons

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

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

func CloneParameters(parameters map[WorkflowParameterLabel]string) map[WorkflowParameterLabel]string {
	parametersCopy := make(map[WorkflowParameterLabel]string)
	for k, v := range parameters {
		parametersCopy[k] = v
	}
	return parametersCopy
}

func CopyParameters(destination map[WorkflowParameterLabel]string, source map[WorkflowParameterLabel]string) {
	for k, v := range source {
		destination[k] = v
	}
}

func (s *Status) String() string {
	if s == nil {
		return "Unknown"
	}
	switch *s {
	case Inactive:
		return "Inactive"
	case Active:
		return "Active"
	case Pending:
		return "Pending"
	case Deleted:
		return "Deleted"
	case Initializing:
		return "Initializing"
	case Archived:
		return "Archived"
	case TimedOut:
		return "TimedOut"
	}
	return "Unknown"
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
		CronService,
	}
}

func (st *ServiceType) String() string {
	if st == nil {
		return "Unknown"
	}
	switch *st {
	case LndService:
		return "LndService"
	case VectorService:
		return "VectorService"
	case AmbossService:
		return "AmbossService"
	case TorqService:
		return "TorqService"
	case AutomationService:
		return "AutomationService"
	case LightningCommunicationService:
		return "LightningCommunicationService"
	case RebalanceService:
		return "RebalanceService"
	case MaintenanceService:
		return "MaintenanceService"
	case CronService:
		return "CronService"
	}
	return "Unknown"
}

func (ss *SubscriptionStream) IsChannelBalanceCache() bool {
	if ss != nil && (*ss == ForwardStream ||
		*ss == InvoiceStream ||
		*ss == PaymentStream ||
		*ss == PeerEventStream ||
		*ss == ChannelEventStream ||
		*ss == GraphEventStream ||
		*ss == HtlcEventStream) {
		return true
	}
	return false
}

func (ss *SubscriptionStream) String() string {
	if ss == nil {
		return "Unknown"
	}
	switch *ss {
	case TransactionStream:
		return "TransactionStream"
	case HtlcEventStream:
		return "HtlcEventStream"
	case ChannelEventStream:
		return "ChannelEventStream"
	case GraphEventStream:
		return "GraphEventStream"
	case ForwardStream:
		return "ForwardStream"
	case InvoiceStream:
		return "InvoiceStream"
	case PaymentStream:
		return "PaymentStream"
	case InFlightPaymentStream:
		return "InFlightPaymentStream"
	case PeerEventStream:
		return "PeerEventStream"
	case ChannelBalanceCacheStream:
		return "ChannelBalanceCacheStream"
	}
	return "Unknown"
}

func GetDeltaPerMille(base uint64, amt uint64) int {
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

func IsWorkflowNodeTypeGrouped(workflowNodeType WorkflowNodeType) bool {
	switch workflowNodeType {
	case WorkflowNodeIntervalTrigger:
		return true
	case WorkflowNodeCronTrigger:
		return true
	case WorkflowNodeChannelBalanceEventTrigger:
		return true
	case WorkflowNodeChannelOpenEventTrigger:
		return true
	case WorkflowNodeChannelCloseEventTrigger:
		return true
	}
	return false
}

func SignMessageWithTimeout(unixTime time.Time, nodeId int, message string, singleHash *bool,
	lightningRequestChannel chan interface{}) SignMessageResponse {

	responseChannel := make(chan SignMessageResponse, 1)
	request := SignMessageRequest{
		CommunicationRequest: CommunicationRequest{
			RequestId:   fmt.Sprintf("%v", unixTime.Unix()),
			RequestTime: &unixTime,
			NodeId:      nodeId,
		},
		ResponseChannel: responseChannel,
		Message:         message,
		SingleHash:      singleHash,
	}
	lightningRequestChannel <- request
	time.AfterFunc(LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS*time.Second, func() {
		timeOutMessage := fmt.Sprintf("Sign Message timed out after %v seconds.", LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS)
		responseChannel <- SignMessageResponse{
			Request: request,
			CommunicationResponse: CommunicationResponse{
				Status:  TimedOut,
				Message: timeOutMessage,
				Error:   timeOutMessage,
			},
		}
	})
	response := <-responseChannel
	return response
}

func SignatureVerificationRequestWithTimeout(unixTime time.Time, nodeId int, message string, signature string,
	lightningRequestChannel chan interface{}) SignatureVerificationResponse {

	responseChannel := make(chan SignatureVerificationResponse, 1)
	request := SignatureVerificationRequest{
		CommunicationRequest: CommunicationRequest{
			RequestId:   fmt.Sprintf("%v", unixTime.Unix()),
			RequestTime: &unixTime,
			NodeId:      nodeId,
		},
		ResponseChannel: responseChannel,
		Message:         message,
		Signature:       signature,
	}
	lightningRequestChannel <- request
	time.AfterFunc(LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS*time.Second, func() {
		timeOutMessage := fmt.Sprintf("Signature Verification timed out after %v seconds.", LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS)
		responseChannel <- SignatureVerificationResponse{
			Request: request,
			CommunicationResponse: CommunicationResponse{
				Status:  TimedOut,
				Message: timeOutMessage,
				Error:   timeOutMessage,
			},
		}
	})
	response := <-responseChannel
	return response
}

func GetWorkflowNodes() map[WorkflowNodeType]WorkflowNodeTypeParameters {
	all := make(map[WorkflowParameterLabel]WorkflowParameterType)
	all[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	all[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings
	all[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings
	all[WorkflowParameterLabelIncomingChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelOutgoingChannels] = WorkflowParameterTypeChannelIds
	all[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus

	channelsOnly := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelsOnly[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds

	channelFilterRequiredInputs := channelsOnly
	channelFilterRequiredOutputs := channelsOnly

	channelPolicyConfiguratorOptionalInputs := all
	channelPolicyConfiguratorOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyConfiguratorOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyConfiguratorOptionalOutputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings

	channelPolicyAutoRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyAutoRunRequiredInputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyAutoRunOptionalInputs := all
	channelPolicyAutoRunRequiredOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyAutoRunOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyAutoRunOptionalOutputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	channelPolicyAutoRunOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	channelPolicyAutoRunOptionalOutputs[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus

	channelPolicyRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	channelPolicyRunRequiredInputs[WorkflowParameterLabelRoutingPolicySettings] = WorkflowParameterTypeRoutingPolicySettings
	channelPolicyRunOptionalInputs := channelPolicyAutoRunOptionalInputs
	channelPolicyRunRequiredOutputs := channelPolicyAutoRunRequiredOutputs
	channelPolicyRunOptionalOutputs := channelPolicyAutoRunOptionalOutputs

	rebalanceConfiguratorOptionalInputs := all
	rebalanceConfiguratorOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceConfiguratorOptionalOutputs[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings

	rebalanceAutoRunRequiredInputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceAutoRunRequiredInputs[WorkflowParameterLabelIncomingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunRequiredInputs[WorkflowParameterLabelOutgoingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunOptionalInputs := all
	rebalanceAutoRunRequiredOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceAutoRunOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelRebalanceSettings] = WorkflowParameterTypeRebalanceSettings
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelIncomingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelOutgoingChannels] = WorkflowParameterTypeChannelIds
	rebalanceAutoRunOptionalOutputs[WorkflowParameterLabelStatus] = WorkflowParameterTypeStatus

	rebalanceRunRequiredInputs := rebalanceAutoRunRequiredInputs
	rebalanceRunOptionalInputs := rebalanceAutoRunOptionalInputs
	rebalanceRunRequiredOutputs := rebalanceAutoRunRequiredOutputs
	rebalanceRunOptionalOutputs := rebalanceAutoRunOptionalOutputs

	addTagOptionalInputs := channelsOnly
	addTagOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	addTagOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	addTagOptionalOutputs[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings

	removeTagOptionalInputs := channelsOnly
	removeTagOptionalOutputs := make(map[WorkflowParameterLabel]WorkflowParameterType)
	removeTagOptionalOutputs[WorkflowParameterLabelChannels] = WorkflowParameterTypeChannelIds
	removeTagOptionalOutputs[WorkflowParameterLabelTagSettings] = WorkflowParameterTypeTagSettings

	return map[WorkflowNodeType]WorkflowNodeTypeParameters{
		WorkflowTrigger: {
			WorkflowNodeType: WorkflowTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeManualTrigger: {
			WorkflowNodeType: WorkflowNodeManualTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeIntervalTrigger: {
			WorkflowNodeType: WorkflowNodeIntervalTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeCronTrigger: {
			WorkflowNodeType: WorkflowNodeCronTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeChannelBalanceEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeChannelOpenEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeChannelCloseEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeStageTrigger: {
			WorkflowNodeType: WorkflowNodeStageTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeChannelFilter: {
			WorkflowNodeType: WorkflowNodeChannelFilter,
			RequiredInputs:   channelFilterRequiredInputs,
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelFilterRequiredOutputs,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelPolicyConfigurator: {
			WorkflowNodeType: WorkflowNodeChannelPolicyConfigurator,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   channelPolicyConfiguratorOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  channelPolicyConfiguratorOptionalOutputs,
		},
		WorkflowNodeChannelPolicyAutoRun: {
			WorkflowNodeType: WorkflowNodeChannelPolicyAutoRun,
			RequiredInputs:   channelPolicyAutoRunRequiredInputs,
			OptionalInputs:   channelPolicyAutoRunOptionalInputs,
			RequiredOutputs:  channelPolicyAutoRunRequiredOutputs,
			OptionalOutputs:  channelPolicyAutoRunOptionalOutputs,
		},
		WorkflowNodeChannelPolicyRun: {
			WorkflowNodeType: WorkflowNodeChannelPolicyRun,
			RequiredInputs:   channelPolicyRunRequiredInputs,
			OptionalInputs:   channelPolicyRunOptionalInputs,
			RequiredOutputs:  channelPolicyRunRequiredOutputs,
			OptionalOutputs:  channelPolicyRunOptionalOutputs,
		},
		WorkflowNodeRebalanceConfigurator: {
			WorkflowNodeType: WorkflowNodeRebalanceConfigurator,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   rebalanceConfiguratorOptionalInputs,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  rebalanceConfiguratorOptionalOutputs,
		},
		WorkflowNodeRebalanceAutoRun: {
			WorkflowNodeType: WorkflowNodeRebalanceAutoRun,
			RequiredInputs:   rebalanceAutoRunRequiredInputs,
			OptionalInputs:   rebalanceAutoRunOptionalInputs,
			RequiredOutputs:  rebalanceAutoRunRequiredOutputs,
			OptionalOutputs:  rebalanceAutoRunOptionalOutputs,
		},
		WorkflowNodeRebalanceRun: {
			WorkflowNodeType: WorkflowNodeRebalanceRun,
			RequiredInputs:   rebalanceRunRequiredInputs,
			OptionalInputs:   rebalanceRunOptionalInputs,
			RequiredOutputs:  rebalanceRunRequiredOutputs,
			OptionalOutputs:  rebalanceRunOptionalOutputs,
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
		WorkflowNodeSetVariable: {
			WorkflowNodeType: WorkflowNodeSetVariable,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   all,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeFilterOnVariable: {
			WorkflowNodeType: WorkflowNodeFilterOnVariable,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   all,
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
	}
}
