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
	MaintenanceService
	CronService
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

const VECTOR_URL = "https://vector.ln.capital/"
const VECTOR_PING_URL_SUFFIX = "api/publicNodeEvents/ping"
const VECTOR_SHORTCHANNELID_URL_SUFFIX = "api/bitcoin/shortChannelId"
const VECTOR_TRANSACTIONDETAILS_URL_SUFFIX = "api/bitcoin/transactionDetails"
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

const ROUTING_POLICY_UPDATE_LIMITER_SECONDS = 5 * 60

const REBALANCE_QUEUE_TICKER_SECONDS = 10
const REBALANCE_MAXIMUM_CONCURRENCY = 10
const REBALANCE_ROUTE_FAILED_HOP_ALLOWED_DELTA_PER_MILLE = 10
const REBALANCE_REBALANCE_DELAY_MILLISECONDS = 500
const REBALANCE_SUCCESS_TIMEOUT_SECONDS = 2 * 60
const REBALANCE_TIMEOUT_SECONDS = 2 * 60 * 60
const REBALANCE_RUNNER_TIMEOUT_SECONDS = 1 * 60 * 60
const REBALANCE_ROUTES_TIMEOUT_SECONDS = 1 * 60
const REBALANCE_ROUTE_TIMEOUT_SECONDS = 10 * 60
const REBALANCE_RESULTS_TIMEOUT_SECONDS = 5 * 60
const REBALANCE_MINIMUM_DELTA_SECONDS = 10 * 60

const MAINTENANCE_QUEUE_TICKER_SECONDS = 60 * 60
const MAINTENANCE_VECTOR_DELAY_MILLISECONDS = 500

const LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS = 10

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
	return "unknown"
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
	ImportPeerEvents
	ImportTransactions
	ImportPayments
	ImportInvoices
	ImportForwards
	ImportHistoricForwards
)
const NodeConnectionDetailCustomSettingsMax = int(ImportHistoricForwards)*2 - 1

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
	WorkflowNodeDataSourceAllChannels
	WorkflowNodeDataSourceEventChannels
)

type WorkflowTriggerType int

const (
	WorkflowTriggerTimingService = WorkflowTriggerType(iota)
	WorkflowTriggerEventService
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

type ChannelStatusRequest int
