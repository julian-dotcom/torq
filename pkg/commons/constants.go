package commons

type ServiceType int

// When adding here also add to GetServiceTypes
const (
	TorqService = ServiceType(iota)
	MaintenanceService
	AutomationService
	CronService
	VectorService
	AmbossService
	RebalanceService
	LndServiceChannelEventStream
	LndServiceGraphEventStream
	LndServiceTransactionStream
	LndServiceHtlcEventStream
	LndServiceForwardStream
	LndServiceInvoiceStream
	LndServicePaymentStream
	LndServicePeerEventStream
	LndServiceInFlightPaymentStream
	LndServiceChannelBalanceCacheStream
)

const UnknownEnumString = "Unknown"

type Status int

const (
	Inactive = Status(iota)
	Active
	Pending
	Deleted
	Initializing
	Archived
	TimedOut
)

type Implementation int

const (
	LND = Implementation(iota)
	CLN
)

type Chain int

const (
	Bitcoin = Chain(iota)
	Litecoin
)

type Network int

const (
	MainNet = Network(iota)
	TestNet
	RegTest
	SigNet
	SimNet
)

type ChannelStatus int

const (
	Opening = ChannelStatus(iota)
	Open
	Closing
	CooperativeClosed      = 100
	LocalForceClosed       = 101
	RemoteForceClosed      = 102
	BreachClosed           = 103
	FundingCancelledClosed = 104
	AbandonedClosed        = 105
)

type ChannelFlags uint32

const (
	FundedOn ChannelFlags = 1 << iota
	ClosedOn
)
const ChannelFlagsMax = int(ClosedOn)*2 - 1

// override to get human readable enum
func (s ChannelStatus) String() string {
	switch s {
	case Opening:
		return "Opening"
	case Open:
		return "Open"
	case Closing:
		return "Closing"
	case CooperativeClosed:
		return "Cooperative Closed"
	case LocalForceClosed:
		return "Local Force Closed"
	case RemoteForceClosed:
		return "Remote Force Closed"
	case BreachClosed:
		return "Breach Closed"
	case FundingCancelledClosed:
		return "Funding Cancelled Closed"
	case AbandonedClosed:
		return "Abandoned Closed"
	}
	return UnknownEnumString
}

type PingSystem uint32

const (
	Amboss PingSystem = 1 << iota
	Vector
)
const PingSystemMax = int(Vector)*2 - 1

type NodeConnectionDetailCustomSettings uint32

const (
	ImportFailedPayments NodeConnectionDetailCustomSettings = 1 << iota
	ImportHtlcEvents
	ImportPeerEventsDeleted
	ImportTransactions
	ImportPayments
	ImportInvoices
	ImportForwards
	ImportHistoricForwards
)
const NodeConnectionDetailCustomSettingsMax = int(ImportHistoricForwards)*2 - 1

type ImportType int

const (
	ImportChannelRoutingPolicies = ImportType(iota)
	ImportNodeInformation
	ImportAllChannels
	ImportPendingChannelsOnly
)

const (
	MEMPOOL string = "https://mempool.space/lightning/channel/"
	AMBOSS  string = "https://amboss.space/edge/"
	ONEML   string = "https://1ml.com/channel/"
)

type WorkflowNodeType int

const (
	WorkflowNodeIntervalTrigger = WorkflowNodeType(iota)
	WorkflowNodeChannelBalanceEventTrigger
	WorkflowNodeChannelFilter
	WorkflowNodeChannelPolicyConfigurator
	WorkflowNodeRebalanceConfigurator
	WorkflowNodeStageTrigger
	WorkflowNodeRebalanceRun
	WorkflowNodeChannelPolicyRun
	WorkflowNodeSetVariable
	WorkflowNodeFilterOnVariable
	WorkflowTrigger
	WorkflowNodeAddTag
	WorkflowNodeRemoveTag
	WorkflowNodeChannelOpenEventTrigger
	WorkflowNodeChannelCloseEventTrigger
	WorkflowNodeCronTrigger
	WorkflowNodeManualTrigger
	WorkflowNodeChannelPolicyAutoRun
	WorkflowNodeRebalanceAutoRun
	WorkflowNodeDataSourceTorqChannels
	WorkflowNodeChannelBalanceEventFilter
)

type WorkflowParameterType string

const (
	WorkflowParameterTypeChannelIds            = WorkflowParameterType("channelIds")
	WorkflowParameterTypeRoutingPolicySettings = WorkflowParameterType("routingPolicySettings")
	WorkflowParameterTypeRebalanceSettings     = WorkflowParameterType("rebalanceSettings")
	WorkflowParameterTypeTagSettings           = WorkflowParameterType("tagSettings")
	WorkflowParameterTypeStatus                = WorkflowParameterType("status")
)

type WorkflowParameterLabel string

const (
	WorkflowParameterLabelChannels              = WorkflowParameterLabel("channels")
	WorkflowParameterLabelRoutingPolicySettings = WorkflowParameterLabel("routingPolicySettings")
	WorkflowParameterLabelRebalanceSettings     = WorkflowParameterLabel("rebalanceSettings")
	WorkflowParameterLabelTagSettings           = WorkflowParameterLabel("tagSettings")
	WorkflowParameterLabelIncomingChannels      = WorkflowParameterLabel("incomingChannels")
	WorkflowParameterLabelOutgoingChannels      = WorkflowParameterLabel("outgoingChannels")
	WorkflowParameterLabelStatus                = WorkflowParameterLabel("status")
	WorkflowParameterLabelAllChannels           = WorkflowParameterLabel("allChannels")
	WorkflowParameterLabelEventChannels         = WorkflowParameterLabel("eventChannels")
	WorkflowParameterLabelEvents                = WorkflowParameterLabel("events")
)

type WorkflowNodeTypeParameters struct {
	WorkflowNodeType WorkflowNodeType
	RequiredInputs   map[WorkflowParameterLabel]WorkflowParameterType
	OptionalInputs   map[WorkflowParameterLabel]WorkflowParameterType
	RequiredOutputs  map[WorkflowParameterLabel]WorkflowParameterType
	OptionalOutputs  map[WorkflowParameterLabel]WorkflowParameterType
}

type RebalanceRequestOrigin int

const (
	RebalanceRequestWorkflowNode = RebalanceRequestOrigin(iota)
	RebalanceRequestManual
)

type BalanceUpdateEventOrigin int

const (
	BalanceUpdateForwardEvent = BalanceUpdateEventOrigin(iota)
	BalanceUpdateInvoiceEvent
	BalanceUpdatePaymentEvent
)
