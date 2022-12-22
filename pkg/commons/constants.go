package commons

type ServiceType int

// When adding here also add to GetServiceTypes
const (
	LndService = ServiceType(iota)
	VectorService
	AmbossService
	TorqService
	AutomationService
	LightningCommunicationService
	RebalanceService
)

const TorqDummyNodeId = -1337

type ServiceCommand int

const (
	Boot = ServiceCommand(iota)
	// Kill requires Out return channel
	// status: Active => Service kill initiated
	// status: Pending => Service is booting and cannot be killed
	// status: Inactive => Service was not running and could not be killed
	Kill
)

type ServiceChannelMessage = struct {
	ServiceType    ServiceType
	ServiceCommand ServiceCommand
	NodeId         int
	// EnforcedServiceStatus is a one time status enforcement for a service
	EnforcedServiceStatus *Status
	// NoDelay is a one time no delay enforcement for a service
	NoDelay bool
	Out     chan Status
}

const BOOTSTRAPPING_TIME_MINUTES = 8 * 60
const TOLERATED_SUBSCRIPTION_DOWNTIME_SECONDS = 15
const STREAM_LND_INVOICES_INTERVAL_SLEEP = 10
const STREAM_LND_INVOICES_SLEEP_MILLISECONDS = 1000
const STREAM_LND_INVOICES_DELTA_TIME_MILLISECONDS = 1000
const STREAM_LND_MAX_INVOICES = 1000
const STREAM_LND_MAX_FORWARDS = 50000

const STREAM_BOOTED_CHECK_SECONDS = 5
const STREAM_PAYMENTS_TICKER_SECONDS = 10
const STREAM_INFLIGHT_PAYMENTS_TICKER_SECONDS = 60
const STREAM_FORWARDS_TICKER_SECONDS = 10

const STREAM_ERROR_SLEEP_SECONDS = 60
const SERVICES_ERROR_SLEEP_SECONDS = 60

// 70 because a reconnection is attempted every 60 seconds
const AVOID_CHANNEL_AND_POLICY_IMPORT_RERUN_TIME_SECONDS = 70
const GENERIC_BOOTSTRAPPING_TIME_SECONDS = 60

const CHANNELBALANCE_TICKER_SECONDS = 150
const CHANNELBALANCE_BOOTSTRAP_TICKER_SECONDS = 10

const AMBOSS_SLEEP_SECONDS = 25
const VECTOR_SLEEP_SECONDS = 20

const WORKFLOW_LOG_COUNT = 100
const WORKFLOW_TICKER_SECONDS = 10

type Status int

const (
	Inactive = Status(iota)
	Active
	Pending
	Deleted
	Initializing
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

type PingSystem byte

const (
	Amboss PingSystem = 1 << iota
	Vector
)
const PingSystemMax = int(Vector)*2 - 1

type NodeConnectionDetailCustomSettings byte

const (
	ImportFailedPayments NodeConnectionDetailCustomSettings = 1 << iota
)
const NodeConnectionDetailCustomSettingsMax = int(ImportFailedPayments)*2 - 1

type SubscriptionStream int

const (
	TransactionStream = SubscriptionStream(iota)
	HtlcEventStream
	ChannelEventStream
	GraphEventStream
	ForwardStream
	InvoiceStream
	PaymentStream
	InFlightPaymentStream
	PeerEventStream
	ChannelBalanceCacheStream
)

var SubscriptionStreams = []SubscriptionStream{ //nolint:gochecknoglobals
	TransactionStream,
	HtlcEventStream,
	ChannelEventStream,
	GraphEventStream,
	ForwardStream,
	InvoiceStream,
	PaymentStream,
	InFlightPaymentStream,
	PeerEventStream,
	ChannelBalanceCacheStream,
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

type ImportRequest struct {
	ImportType ImportType
	Out        chan error
}

type ImportType int

const (
	ImportChannelAndRoutingPolicies = ImportType(iota)
	ImportNodeInformation
)

const (
	MEMPOOL string = "https://mempool.space/lightning/channel/"
	AMBOSS  string = "https://amboss.space/edge/"
	ONEML   string = "https://1ml.com/channel/"
)

type WorkflowNodeType int

const (
	WorkflowNodeTimeTrigger = WorkflowNodeType(iota)
	WorkflowNodeChannelBalanceEventTrigger
	WorkflowNodeChannelFilter
	WorkflowNodeRoutingPolicyParameters
	WorkflowNodeRebalanceParameters
	WorkflowNodeStageTrigger
	WorkflowNodeRebalanceRun
	WorkflowNodeRoutingPolicyRun
	WorkflowNodeSetVariable
	WorkflowNodeFilterOnVariable
)

type WorkflowTriggerType int

const (
	WorkflowTriggerTimingService = WorkflowTriggerType(iota)
	WorkflowTriggerEventService
)

type WorkflowParameter string

const (
	WorkflowParameterAny                   = WorkflowParameter("any")
	WorkflowParameterTriggered             = WorkflowParameter("triggered")
	WorkflowParameterChannelIds            = WorkflowParameter("channelIds")
	WorkflowParameterRoutingPolicySettings = WorkflowParameter("routingPolicySettings")
	WorkflowParameterRebalanceSettings     = WorkflowParameter("rebalanceSettings")
	WorkflowParameterStatus                = WorkflowParameter("status")
)

type WorkflowNodeTypeParameters struct {
	WorkflowNodeType WorkflowNodeType
	RequiredInputs   []WorkflowParameterWithLabel
	OptionalInputs   []WorkflowParameterWithLabel
	RequiredOutputs  []WorkflowParameterWithLabel
	OptionalOutputs  []WorkflowParameterWithLabel
}

type WorkflowParameterType int

const (
	WorkflowParameterTimeInSeconds = WorkflowParameterType(iota)
	WorkflowParameterBalanceShifted
	WorkflowParameterVariableName
	WorkflowParameterVariableValueString
	WorkflowParameterVariableValueNumber
)
