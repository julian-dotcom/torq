package core

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
)

func (s Network) String() string {
	switch s {
	case MainNet:
		return "MainNet"
	case TestNet:
		return "TestNet"
	case RegTest:
		return "RegTest"
	case SigNet:
		return "SigNet"
	case SimNet:
		return "SimNet"
	}
	return UnknownEnumString
}

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

func (s Chain) String() string {
	switch s {
	case Bitcoin:
		return "MainNet"
	case Litecoin:
		return "Litecoin"
	}
	return UnknownEnumString
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
		return UnknownEnumString
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
	return UnknownEnumString
}

func (s *ServiceStatus) String() string {
	if s == nil {
		return UnknownEnumString
	}
	switch *s {
	case ServiceInactive:
		return "Inactive"
	case ServiceActive:
		return "Active"
	case ServicePending:
		return "Pending"
	case ServiceInitializing:
		return "Initializing"
	}
	return UnknownEnumString
}

func (ps PingSystem) AddPingSystem(pingSystem PingSystem) PingSystem {
	return ps | pingSystem
}
func (ps PingSystem) HasPingSystem(pingSystem PingSystem) bool {
	return ps&pingSystem != 0
}
func (ps PingSystem) RemovePingSystem(pingSystem PingSystem) PingSystem {
	return ps & ^pingSystem
}
func (ps PingSystem) GetServiceType() *ServiceType {
	switch {
	case ps.HasPingSystem(Vector):
		vectorService := VectorService
		return &vectorService
	case ps.HasPingSystem(Amboss):
		ambossService := AmbossService
		return &ambossService
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: PingSystem not supported")
		return nil
	}
}

func (cs NodeConnectionDetailCustomSettings) AddNodeConnectionDetailCustomSettings(
	customSettings NodeConnectionDetailCustomSettings) NodeConnectionDetailCustomSettings {

	return cs | customSettings
}
func (cs NodeConnectionDetailCustomSettings) HasNodeConnectionDetailCustomSettings(
	customSettings NodeConnectionDetailCustomSettings) bool {

	return cs&customSettings != 0
}
func (cs NodeConnectionDetailCustomSettings) RemoveNodeConnectionDetailCustomSettings(
	customSettings NodeConnectionDetailCustomSettings) NodeConnectionDetailCustomSettings {

	return cs & ^customSettings
}
func (cs NodeConnectionDetailCustomSettings) GetServiceType() *ServiceType {
	switch {
	case cs.HasNodeConnectionDetailCustomSettings(ImportFailedPayments),
		cs.HasNodeConnectionDetailCustomSettings(ImportPayments):
		lndServicePaymentStream := LndServicePaymentStream
		return &lndServicePaymentStream
	case cs.HasNodeConnectionDetailCustomSettings(ImportHtlcEvents):
		lndServiceHtlcEventStream := LndServiceHtlcEventStream
		return &lndServiceHtlcEventStream
	case cs.HasNodeConnectionDetailCustomSettings(ImportTransactions):
		lndServiceTransactionStream := LndServiceTransactionStream
		return &lndServiceTransactionStream
	case cs.HasNodeConnectionDetailCustomSettings(ImportInvoices):
		lndServiceInvoiceStream := LndServiceInvoiceStream
		return &lndServiceInvoiceStream
	case cs.HasNodeConnectionDetailCustomSettings(ImportForwards),
		cs.HasNodeConnectionDetailCustomSettings(ImportHistoricForwards):
		lndServiceForwardStream := LndServiceForwardStream
		return &lndServiceForwardStream
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: NodeConnectionDetailCustomSettings not supported")
		return nil
	}
}

func (cf ChannelFlags) AddChannelFlag(channelFlags ChannelFlags) ChannelFlags {
	return cf | channelFlags
}
func (cf ChannelFlags) HasChannelFlag(channelFlags ChannelFlags) bool {
	return cf&channelFlags != 0
}
func (cf ChannelFlags) RemoveChannelFlag(channelFlags ChannelFlags) ChannelFlags {
	return cf & ^channelFlags
}

func GetCoreServiceTypes() []ServiceType {
	return []ServiceType{
		RootService,
		MaintenanceService,
		AutomationIntervalTriggerService,
		AutomationChannelBalanceEventTriggerService,
		AutomationChannelEventTriggerService,
		AutomationScheduledTriggerService,
		CronService,
		NotifierService,
		SlackService,
		TelegramHighService,
		TelegramLowService,
	}
}

func GetLndServiceTypes() []ServiceType {
	return []ServiceType{
		VectorService,
		AmbossService,
		RebalanceService,
		LndServiceChannelEventStream,
		LndServiceGraphEventStream,
		LndServiceTransactionStream,
		LndServiceHtlcEventStream,
		LndServiceForwardStream,
		LndServiceInvoiceStream,
		LndServicePaymentStream,
		LndServicePeerEventStream,
		LndServiceInFlightPaymentStream,
		LndServiceChannelBalanceCacheStream,
	}
}

func GetNodeConnectionDetailCustomSettings() []NodeConnectionDetailCustomSettings {
	return []NodeConnectionDetailCustomSettings{
		ImportFailedPayments,
		ImportHtlcEvents,
		ImportTransactions,
		ImportPayments,
		ImportInvoices,
		ImportForwards,
		ImportHistoricForwards,
	}
}

func (st *ServiceType) String() string {
	if st == nil {
		return UnknownEnumString
	}
	switch *st {
	case VectorService:
		return "VectorService"
	case AmbossService:
		return "AmbossService"
	case RootService:
		return "RootService"
	case AutomationChannelBalanceEventTriggerService:
		return "AutomationChannelBalanceEventTriggerService"
	case AutomationChannelEventTriggerService:
		return "AutomationChannelEventTriggerService"
	case AutomationIntervalTriggerService:
		return "AutomationIntervalTriggerService"
	case AutomationScheduledTriggerService:
		return "AutomationScheduledTriggerService"
	case RebalanceService:
		return "RebalanceService"
	case MaintenanceService:
		return "MaintenanceService"
	case CronService:
		return "CronService"
	case NotifierService:
		return "NotifierService"
	case SlackService:
		return "SlackService"
	case TelegramHighService:
		return "TelegramHighService"
	case TelegramLowService:
		return "TelegramLowService"
	case LndServiceChannelEventStream:
		return "LndServiceChannelEventStream"
	case LndServiceGraphEventStream:
		return "LndServiceGraphEventStream"
	case LndServiceTransactionStream:
		return "LndServiceTransactionStream"
	case LndServiceHtlcEventStream:
		return "LndServiceHtlcEventStream"
	case LndServiceForwardStream:
		return "LndServiceForwardStream"
	case LndServiceInvoiceStream:
		return "LndServiceInvoiceStream"
	case LndServicePaymentStream:
		return "LndServicePaymentStream"
	case LndServicePeerEventStream:
		return "LndServicePeerEventStream"
	case LndServiceInFlightPaymentStream:
		return "LndServiceInFlightPaymentStream"
	case LndServiceChannelBalanceCacheStream:
		return "LndServiceChannelBalanceCacheStream"
	}
	return UnknownEnumString
}

func (st *ServiceType) IsChannelBalanceCache() bool {
	if st != nil && (*st == LndServiceForwardStream ||
		*st == LndServiceInvoiceStream ||
		*st == LndServicePaymentStream ||
		*st == LndServicePeerEventStream ||
		*st == LndServiceChannelEventStream ||
		*st == LndServiceGraphEventStream) {
		return true
	}
	return false
}

func (st *ServiceType) IsLndService() bool {
	if st != nil && (*st == VectorService ||
		*st == AmbossService ||
		*st == RebalanceService ||
		*st == LndServiceChannelEventStream ||
		*st == LndServiceGraphEventStream ||
		*st == LndServiceTransactionStream ||
		*st == LndServiceHtlcEventStream ||
		*st == LndServiceForwardStream ||
		*st == LndServiceInvoiceStream ||
		*st == LndServicePaymentStream ||
		*st == LndServicePeerEventStream ||
		*st == LndServiceInFlightPaymentStream ||
		*st == LndServiceChannelBalanceCacheStream) {
		return true
	}
	return false
}

func (st *ServiceType) GetNodeConnectionDetailCustomSettings() []NodeConnectionDetailCustomSettings {
	if st == nil {
		return nil
	}
	switch *st {
	case LndServicePaymentStream:
		return []NodeConnectionDetailCustomSettings{ImportFailedPayments, ImportPayments}
	case LndServiceHtlcEventStream:
		return []NodeConnectionDetailCustomSettings{ImportHtlcEvents}
	case LndServiceTransactionStream:
		return []NodeConnectionDetailCustomSettings{ImportTransactions}
	case LndServiceInvoiceStream:
		return []NodeConnectionDetailCustomSettings{ImportInvoices}
	case LndServiceForwardStream:
		return []NodeConnectionDetailCustomSettings{ImportForwards, ImportHistoricForwards}
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: ServiceType not supported")
		return nil
	}
}

func (st *ServiceType) GetPingSystem() *PingSystem {
	if st == nil {
		return nil
	}
	switch *st {
	case AmbossService:
		amboss := Amboss
		return &amboss
	case VectorService:
		vector := Vector
		return &vector
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: ServiceType not supported")
		return nil
	}
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

func Sleep(ctx context.Context, d time.Duration) bool {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-ticker.C:
	}
	return true
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

func GetWorkflowParameterLabelsEnforced() []WorkflowParameterLabel {
	return []WorkflowParameterLabel{
		WorkflowParameterLabelRoutingPolicySettings,
		WorkflowParameterLabelRebalanceSettings,
		WorkflowParameterLabelStatus,
	}
}

func GetPeer(peer *lnrpc.Peer) Peer {
	p := Peer{
		PubKey:          peer.PubKey,
		Address:         peer.Address,
		BytesSent:       peer.BytesSent,
		BytesRecv:       peer.BytesRecv,
		SatSent:         peer.SatSent,
		SatRecv:         peer.SatRecv,
		Inbound:         peer.Inbound,
		PingTime:        peer.PingTime,
		SyncType:        PeerSyncType(peer.SyncType),
		FlapCount:       peer.FlapCount,
		LastFlapNS:      peer.LastFlapNs,
		LastPingPayload: peer.LastPingPayload,
	}

	features := make([]FeatureEntry, len(peer.Features))
	for key, feature := range peer.Features {
		features = append(features, FeatureEntry{
			Key: key,
			Value: Feature{
				Name:       feature.Name,
				IsRequired: feature.IsRequired,
				IsKnown:    feature.IsKnown,
			},
		})
	}
	p.Features = features

	timeStampedErrors := make([]TimeStampedError, len(peer.Errors))
	for _, tse := range peer.Errors {
		timeStampedErrors = append(timeStampedErrors, TimeStampedError{
			Timestamp: tse.Timestamp,
			Error:     tse.Error,
		})
	}
	p.Errors = timeStampedErrors
	return p
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

	channelBalanceEventFilterRequiredInputs := channelsOnly
	channelBalanceEventFilterRequiredOutputs := channelsOnly

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
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeManualTrigger: {
			WorkflowNodeType: WorkflowNodeManualTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeIntervalTrigger: {
			WorkflowNodeType: WorkflowNodeIntervalTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeCronTrigger: {
			WorkflowNodeType: WorkflowNodeCronTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelBalanceEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelOpenEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelCloseEventTrigger: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeStageTrigger: {
			WorkflowNodeType: WorkflowNodeStageTrigger,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalOutputs:  all,
		},
		WorkflowNodeDataSourceTorqChannels: {
			WorkflowNodeType: WorkflowNodeDataSourceTorqChannels,
			RequiredInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelsOnly,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelFilter: {
			WorkflowNodeType: WorkflowNodeChannelFilter,
			RequiredInputs:   channelFilterRequiredInputs,
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelFilterRequiredOutputs,
			OptionalOutputs:  make(map[WorkflowParameterLabel]WorkflowParameterType),
		},
		WorkflowNodeChannelBalanceEventFilter: {
			WorkflowNodeType: WorkflowNodeChannelBalanceEventFilter,
			RequiredInputs:   channelBalanceEventFilterRequiredInputs,
			OptionalInputs:   make(map[WorkflowParameterLabel]WorkflowParameterType),
			RequiredOutputs:  channelBalanceEventFilterRequiredOutputs,
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
