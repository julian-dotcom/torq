package commons

type ServiceType int

const (
	LndSubscription = ServiceType(iota)
	VectorSubscription
	AmbossSubscription
)

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
	Out            chan Status
}

type Status int

const (
	Inactive = Status(iota)
	Active
	Pending
	Deleted
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
