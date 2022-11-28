package commons

type ServiceType int

const (
	LndService = ServiceType(iota)
	VectorService
	AmbossService
	TorqService
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

const STREAM_LND_INVOICES_INTERVAL_SLEEP = 10
const STREAM_LND_INVOICES_SLEEP_MILLISECONDS = 1000
const STREAM_LND_INVOICES_DELTA_TIME_MILLISECONDS = 1000

const STREAM_BOOTED_CHECK_SECONDS = 5
const STREAM_PAYMENTS_TICKER_SECONDS = 10
const STREAM_INFLIGHT_PAYMENTS_TICKER_SECONDS = 60
const STREAM_FORWARDS_TICKER_SECONDS = 10

const STREAM_ERROR_SLEEP_SECONDS = 60
const SERVICES_ERROR_SLEEP_SECONDS = 60

// 70 because a reconnection is attempted every 60 seconds
const AVOID_CHANNEL_AND_POLICY_IMPORT_RERUN_TIME_SECONDS = 70
const INVOICE_BOOTSTRAPPING_TIME_SECONDS = 60
const GENERIC_BOOTSTRAPPING_TIME_SECONDS = 60

const AMBOSS_SLEEP_SECONDS = 25
const VECTOR_SLEEP_SECONDS = 20

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
	SigNet
	SimNet
	RegTest
	TestNet
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
)

type ImportRequest struct {
	ImportType ImportType
	Out        chan error
}

type ImportType int

const (
	ImportChannelAndRoutingPolicies = ImportType(iota)
	ImportNodeInformation
)
